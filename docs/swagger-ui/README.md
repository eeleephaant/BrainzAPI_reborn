### Swagger UI

Запуск локально (из корня репозитория):

```bash
docker compose -f docs/swagger-ui/docker-compose.swagger.yml up -d
```

Открыть Swagger UI:
- `http://127.0.0.1:8099`

Спека лежит в `docs/openapi/openapi.yaml`.

Остановить:

```bash
docker compose -f docs/swagger-ui/docker-compose.swagger.yml down
```

