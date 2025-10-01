import os

from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    grpc_port: int = 50053
    kafka_brokers: str = "localhost:9092"
    kafka_sasl_username: str = ""
    kafka_sasl_password: str = ""
    kafka_security_protocol: str = "PLAINTEXT"
    r2_endpoint: str = ""
    r2_access_key_id: str = ""
    r2_secret_access_key: str = ""
    r2_account_id: str = ""
    r2_region: str = "auto"
    # Two buckets: one for original files, one for processed markdown
    r2_bucket_source_name: str = ""
    r2_bucket_processed_name: str = ""
    ms_knowledge_grpc_url: str = "localhost:50052"

    # LLM Provider Configuration
    llm_provider: str = "openai"  # openai | vertex | gemini
    llm_temperature: float = 0.2
    llm_top_p: float = 0.95
    llm_summary_tokens: int = 512
    llm_keyword_count: int = 10

    # OpenAI Configuration (default provider)
    openai_api_key: str = ""
    openai_model: str = "gpt-5-mini-2025-08-07"

    # Vertex AI Configuration (optional)
    vertex_project_id: str = ""
    vertex_location: str = "us-central1"
    vertex_model: str = "gemini-1.5-flash"

    # Gemini Configuration (backward compatibility)
    gemini_api_key: str = ""
    gemini_model: str = "models/gemini-1.5-flash"
    gemini_summary_tokens: int = 512  # deprecated, use llm_summary_tokens
    gemini_keyword_count: int = 10    # deprecated, use llm_keyword_count

    model_config = SettingsConfigDict(
        env_file=".env.local",
        env_file_encoding="utf-8",
        extra="ignore" if os.environ.get("ALLOW_EXTRA_SETTINGS") == "1" else "forbid",
    )

settings = Settings()
