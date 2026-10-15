"""Settings and environment configuration for AI Service."""

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    ENVIRONMENT: str = "development"
    PORT: int = 8000
    HOST: str = "0.0.0.0"
    DATABASE_URL: str = ""

    class Config:
        env_file = ".env"


settings = Settings()
