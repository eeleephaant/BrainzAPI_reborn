FROM python:3.13-slim-bookworm AS builder

WORKDIR /parser_app

RUN apt-get update && apt-get install -y \
	libpq-dev \
	python3-dev \
	gcc \
	--no-install-recommends && rm -rf /var/lib/apt/lists/*

COPY ./requirements.txt /parser_app/requirements.txt

RUN pip install --no-compile --no-cache-dir --upgrade -r /parser_app/requirements.txt

COPY . /parser_app

FROM python:3.13-slim-bookworm

WORKDIR /parser_app

COPY --from=builder /usr/local /usr/local
COPY --from=builder /parser_app /parser_app

ENV PYTHONPATH=/parser_app
ENV PYTHONDONTWRITEBYTECODE=1

CMD ["fastapi", "run", "api.py", "--port", "80"]