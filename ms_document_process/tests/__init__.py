"""Test package configuration for ms_document_process."""

import os


# Unset env vars that are not recognized by the service settings to avoid validation errors.
for _env_var in (
    "next_public_clerk_publishable_key",
    "clerk_secret_key",
    "clerk_webhook_secret",
    "ngrok_authtoken",
    "ngrok_static_domain",
):
    os.environ.pop(_env_var, None)

