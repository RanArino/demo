# LLM Provider Package

Shared package for multiple LLM providers across all microservices.

## Overview

This package provides a unified interface for generating document insights (summaries and keywords) using different LLM providers:

- **OpenAI** (default): gpt-5-mini-2025-08-07 or other OpenAI models
- **Vertex AI**: Google Cloud Vertex AI with Gemini models
- **Gemini**: Google Gemini API (backward compatibility)

## Architecture

```
pkg/llm/
├── __init__.py          # Package exports
├── base.py              # Abstract LLMProvider base class
├── types.py             # Shared types (InsightRequest, InsightResponse)
├── factory.py           # Provider factory and configuration
└── providers/
    ├── __init__.py
    ├── openai.py        # OpenAI Chat Completions API
    ├── vertex.py        # Google Vertex AI
    └── gemini.py        # Google Gemini API
```

## Usage

### Basic Usage

```python
from pkg.llm import LLMProviderFactory, LLMProviderConfig, LLMProviderType

# Configure OpenAI provider (default)
config = LLMProviderConfig(
    provider_type=LLMProviderType.OPENAI,
    openai_api_key="your-api-key",
    openai_model="gpt-5-mini-2025-08-07",
    temperature=0.2,
    top_p=0.95,
)

# Create provider
provider = LLMProviderFactory.create_provider(config)

# Generate insights
response = provider.generate_insights(
    document_text="Your document content here...",
    title="Document Title",
    summary_tokens=512,
    keyword_count=10,
)

print(response.summary)
print(response.keywords)
```

### Using Vertex AI

```python
config = LLMProviderConfig(
    provider_type=LLMProviderType.VERTEX,
    vertex_project_id="your-gcp-project",
    vertex_location="us-central1",
    vertex_model="gemini-1.5-flash",
)

provider = LLMProviderFactory.create_provider(config)
response = provider.generate_insights(document_text="...")
```

### Using Gemini API

```python
config = LLMProviderConfig(
    provider_type=LLMProviderType.GEMINI,
    gemini_api_key="your-gemini-key",
    gemini_model="models/gemini-1.5-flash",
)

provider = LLMProviderFactory.create_provider(config)
response = provider.generate_insights(document_text="...")
```

## Configuration

### Environment Variables

Set these in your microservice's `.env.local`:

```bash
# Provider selection (openai | vertex | gemini)
LLM_PROVIDER=openai

# Common settings
LLM_TEMPERATURE=0.2
LLM_TOP_P=0.95
LLM_SUMMARY_TOKENS=512
LLM_KEYWORD_COUNT=10

# OpenAI (default)
OPENAI_API_KEY=your_openai_api_key
OPENAI_MODEL=gpt-5-mini-2025-08-07

# Vertex AI (optional)
VERTEX_PROJECT_ID=your_gcp_project
VERTEX_LOCATION=us-central1
VERTEX_MODEL=gemini-1.5-flash

# Gemini API (optional)
GEMINI_API_KEY=your_gemini_key
GEMINI_MODEL=models/gemini-1.5-flash
```

## Integration with Microservices

To use this package in a microservice:

1. Add the pkg path to Python path:

```python
import sys
from pathlib import Path

pkg_path = Path(__file__).resolve().parents[4] / "pkg"
if str(pkg_path) not in sys.path:
    sys.path.insert(0, str(pkg_path))

from pkg.llm import LLMProviderFactory, LLMProviderConfig, LLMProviderType
```

2. Configure and use the provider based on your settings.

## Response Format

All providers return a consistent `InsightResponse`:

```python
class InsightResponse(BaseModel):
    summary: str              # Generated summary
    keywords: List[str]       # Extracted keywords
```

## Dependencies

Required packages (add to your microservice's `pyproject.toml`):

```toml
openai = "^1.50.0"                    # For OpenAI provider
google-cloud-aiplatform = "^1.70.0"   # For Vertex AI provider
google-generativeai = "^0.8.5"        # For Gemini provider
```

## Error Handling

All providers raise `RuntimeError` on failure:

```python
try:
    response = provider.generate_insights(document_text="...")
except RuntimeError as e:
    logger.error(f"Failed to generate insights: {e}")
```

## Adding New Providers

To add a new provider:

1. Create a new file in `providers/` (e.g., `anthropic.py`)
2. Implement the `LLMProvider` base class
3. Add provider type to `LLMProviderType` enum in `factory.py`
4. Add provider creation logic to `LLMProviderFactory`
5. Export from `providers/__init__.py`

## Testing

Mock the provider in tests:

```python
from unittest.mock import MagicMock, patch
from pkg.llm.types import InsightResponse

@patch("your_module.LLMProviderFactory.create_provider")
def test_with_mocked_provider(mock_create_provider):
    mock_provider = MagicMock()
    mock_provider.generate_insights.return_value = InsightResponse(
        summary="Test summary",
        keywords=["test", "keywords"]
    )
    mock_create_provider.return_value = mock_provider

    # Your test code here
```