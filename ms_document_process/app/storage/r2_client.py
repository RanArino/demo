from __future__ import annotations

import hashlib
from typing import Optional

import boto3
from botocore.config import Config as BotoConfig


class R2Client:
	def __init__(self, *, endpoint: str, region: str = "auto", access_key_id: str = "", secret_access_key: str = "", use_path_style: bool = True):
		if not endpoint:
			raise ValueError("R2 endpoint is required")
		if not access_key_id or not secret_access_key:
			raise ValueError("R2 credentials are required")
		self._endpoint = endpoint
		self._client = boto3.client(
			"s3",
			endpoint_url=endpoint,
			aws_access_key_id=access_key_id,
			aws_secret_access_key=secret_access_key,
			region_name=region or "auto",
			config=BotoConfig(s3={"addressing_style": "path" if use_path_style else "virtual"}),
		)

	def upload(self, bucket: str, key: str, data: bytes, content_type: Optional[str] = None) -> None:
		kwargs = {"Bucket": bucket, "Key": key, "Body": data}
		if content_type:
			kwargs["ContentType"] = content_type
		self._client.put_object(**kwargs)

	def download(self, bucket: str, key: str) -> bytes:
		resp = self._client.get_object(Bucket=bucket, Key=key)
		return resp["Body"].read()

	def generate_presigned_upload_url(self, bucket: str, key: str, expires_seconds: int = 900, content_type: Optional[str] = None) -> str:
		params = {"Bucket": bucket, "Key": key}
		if content_type:
			params["ContentType"] = content_type
		return self._client.generate_presigned_url(
			ClientMethod="put_object",
			Params=params,
			ExpiresIn=expires_seconds,
		)

	@staticmethod
	def calculate_sha256(data: bytes) -> str:
		return hashlib.sha256(data).hexdigest()
