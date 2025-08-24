FROM python:3.13-slim

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

WORKDIR /app

RUN pip install --no-cache-dir poetry

# Install only dependency metadata first for better caching
COPY pyproject.toml poetry.lock ./

# Force Poetry to install into the global environment, not a venv
ENV POETRY_VIRTUALENVS_CREATE=false \
    POETRY_VIRTUALENVS_IN_PROJECT=false

RUN poetry install --no-root --no-interaction --no-ansi

# Now copy the actual application code
COPY . .

# Ensure no stray project venv remains
RUN rm -rf /app/.venv

EXPOSE 50051

ENV PYTHONPATH=/app
CMD ["python", "-m", "app.main"]
