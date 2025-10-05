FROM python:3.13-alpine3.22 AS builder

WORKDIR /parser_app

RUN apk add --no-cache \
    build-base \
    libpq-dev \
    python3-dev \
    linux-headers \
    bash \
	tini


COPY ./parser_app/requirements.txt /parser_app/requirements.txt

RUN pip install --no-compile --no-cache-dir --upgrade -r /parser_app/requirements.txt

COPY ./parser_app /parser_app

FROM python:3.13-alpine3.22

WORKDIR /

COPY --from=builder /usr/local /usr/local
COPY --from=builder /parser_app /parser_app

COPY --from=builder /sbin/tini /sbin/tini

ENV PYTHONPATH=/parser_app
ENV PYTHONDONTWRITEBYTECODE=1

ENTRYPOINT ["/sbin/tini", "--"]

CMD ["sh", "-c", "python -m parser_app.worker & fastapi run parser_app/api.py --port 80"]
