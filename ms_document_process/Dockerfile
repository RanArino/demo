FROM python:3.13-slim

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

WORKDIR /app

RUN pip install --no-cache-dir poetry

# Copy and install the shared pkg as a dependency
COPY pkg /tmp/pkg
RUN pip install --no-cache-dir /tmp/pkg

# Install only dependency metadata for the service for better caching
COPY ms_document_process/pyproject.toml ms_document_process/poetry.lock ./

# Force Poetry to install into the global environment, not a venv
ENV POETRY_VIRTUALENVS_CREATE=false \
    POETRY_VIRTUALENVS_IN_PROJECT=false

RUN poetry install --no-root --no-interaction --no-ansi

# Now copy the actual application code
COPY ms_document_process/ .

# Ensure no stray project venv remains
RUN rm -rf /app/.venv

EXPOSE 50051

ENV PYTHONPATH=/app
CMD ["python", "-m", "app.main"]
