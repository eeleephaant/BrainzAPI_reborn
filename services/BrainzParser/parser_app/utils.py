import os
import redis


class Vector2:
    def __init__(self, x: int, y: int):
        self.x = x
        self.y = y


DB_HOST = os.getenv("DB_HOST", "localhost")
DB_PORT = int(os.getenv("DB_PORT", 5432))
DB_USER = os.getenv("DB_USER", "postgres")
DB_PASSWORD = os.getenv("DB_PASSWORD", "")
DB_NAME = os.getenv("DB_NAME", "postgres")

DSN = f"postgresql://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"

redis_client = redis.Redis(os.getenv("REDIS_HOST", "redis"), int(os.getenv("REDIS_PORT", 6379)), db=0)
