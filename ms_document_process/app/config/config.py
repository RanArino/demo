from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    grpc_port: int = 50053
    kafka_brokers: str = "localhost:9092"
    kafka_sasl_username: str = ""
    kafka_sasl_password: str = ""
    kafka_security_protocol: str = "PLAINTEXT"
    r2_access_key_id: str
    r2_secret_access_key: str
    r2_account_id: str
    # Two buckets: one for original files, one for processed markdown
    r2_bucket_source_name: str
    r2_bucket_processed_name: str
    ms_knowledge_grpc_url: str = "localhost:50052"

    class Config:
        env_file = ".env.local"
        env_file_encoding = "utf-8"

settings = Settings()
