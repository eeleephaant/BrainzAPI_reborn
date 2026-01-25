import json
import logging
import os
from typing import Awaitable, Callable
from fastapi import FastAPI, Request, Response, UploadFile, File
from fastapi.responses import JSONResponse
import httpx

from parser_app import ttsiigh_utils, utils, database
from parser_app.logging_config import configure_logging

from ttsiigh_utils import XLSXParser  # type: ignore

app = FastAPI()

AUTH_SERVICE_URL = os.getenv("AUTH_SERVICE_URL", "http://brainz-auth:8080/auth")

configure_logging()
log = logging.getLogger("parser_app")


@app.middleware("http")
async def api_key_middleware(
        request: Request,
        call_next: Callable[[Request], Awaitable[Response]]
):
    api_key = request.headers.get("X-API-Key")
    if not api_key:
        return JSONResponse(
            status_code=401,
            content={"status": False, "error_message": "Missing API key"}
        )
    institution_id = request.query_params.get("institution_id")
    if institution_id is None:
        return JSONResponse(
            status_code=400,
            content={"status": False, "error_message": "Institution ID is required"}
        )

    needed_rights = {"perm": {"action": "write", "institution_id": int(institution_id)}}

    async with httpx.AsyncClient() as client:
        try:
            resp = await client.get(AUTH_SERVICE_URL, params={"key": api_key}, timeout=5.0, )
            data = resp.json()
        except httpx.RequestError:
            return JSONResponse(
                status_code=500,
                content={"status": False, "error_message": "Auth service unavailable"}
            )

    if not data.get("status", False):
        return JSONResponse(
            status_code=401,
            content={"status": False, "error_message": data.get("error_message", "Invalid API key")}
        )

    request.state.user = data.get("user")

    response = await call_next(request)
    return response


@app.post("/parse")
async def parse_rsp(file: UploadFile = File(...)):
    try:
        content = await file.read()

        if file.content_type not in ["application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"]:
            return JSONResponse({"status": False, "message": "Invalid file type"}, status_code=400)

        parser = XLSXParser(content)

        groups = parser.extract_groups()
        raw_date = parser.extract_date()

        if ttsiigh_utils.schedule_exists(raw_date):
            return JSONResponse(
                content={
                    "status": False,
                    "message": "Schedule already exists"
                },
                status_code=409
            )

        ttsiigh_utils.add_groups(groups)

        lessons = parser.extract_lessons(ttsiigh_utils.local_timings)

        database.write_lessons_to_bd(lessons)

        channel = "info_stream:1"
        message = {"type": "new schedule", "date": raw_date.strftime("%Y-%m-%d")}
        utils.redis_client.publish(channel, json.dumps(message))

        return JSONResponse(
            content={
                "status": True,
                "date": raw_date.isoformat(),
                "institution_id": 1,
                "groups_count": len(groups),
                "lesson_count": len(lessons)
            },
            status_code=200
        )
    except Exception:
        log.error("Unexpected error during parsing", exc_info=True)
        return JSONResponse(
            content={"status": False, "message": "internal error"},
            status_code=500
        )
