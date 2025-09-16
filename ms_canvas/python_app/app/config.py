from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    # Server
    CANVAS_PY_HOST: str = "0.0.0.0"
    CANVAS_PY_PORT: int = 50051
    CANVAS_MAX_GRPC_MSG_BYTES: int = 64 * 1024 * 1024

    # Embeddings / LLM
    OPENAI_API_KEY: Optional[str] = None
    AZURE_OPENAI_API_KEY: Optional[str] = None
    OPENAI_ORG_ID: Optional[str] = None
    OPENAI_PROJECT_ID: Optional[str] = None

    # Neo4j
    NEO4J_URI: str = "neo4j://localhost:7687"
    NEO4J_USERNAME: str = "neo4j"
    NEO4J_PASSWORD: str = "password"

    # Tests
    INTEGRATION_TESTS: Optional[int] = 0
    RUN_BENCHMARKS: Optional[int] = 0

    class Config:
        env_file = ".env.local"
        env_file_encoding = "utf-8"


settings = Settings()


