import unittest
from unittest.mock import MagicMock, patch

from pkg.llm.types import InsightResponse

from app.config.config import settings
from app.domain.insights import DocumentInsights
from app.service.insights_service import DocumentInsightsService


class TestDocumentInsightsService(unittest.TestCase):
    def setUp(self):
        self.document_text = "Sample PDF content for testing."

    @patch("app.service.insights_service.LLMProviderFactory.create_provider")
    def test_generate_insights_with_openai_provider(self, mock_create_provider):
        """Test insights generation with OpenAI provider."""
        expected_response = InsightResponse(
            summary="This is a test summary of the document.",
            keywords=["test", "document", "sample"]
        )

        mock_provider = MagicMock()
        mock_provider.generate_insights.return_value = expected_response
        mock_create_provider.return_value = mock_provider

        with patch.object(settings, "llm_provider", "openai"), \
             patch.object(settings, "openai_api_key", "test-key"):
            service = DocumentInsightsService()
            insights: DocumentInsights = service.generate_insights(
                self.document_text,
                title="Test Title"
            )

        mock_provider.generate_insights.assert_called_once()
        self.assertEqual(insights.summary, expected_response.summary)
        self.assertEqual(insights.keywords, expected_response.keywords)

    @patch("app.service.insights_service.LLMProviderFactory.create_provider")
    def test_generate_insights_with_vertex_provider(self, mock_create_provider):
        """Test insights generation with Vertex AI provider."""
        expected_response = InsightResponse(
            summary="Vertex AI generated summary.",
            keywords=["vertex", "ai", "test"]
        )

        mock_provider = MagicMock()
        mock_provider.generate_insights.return_value = expected_response
        mock_create_provider.return_value = mock_provider

        with patch.object(settings, "llm_provider", "vertex"), \
             patch.object(settings, "vertex_project_id", "test-project"):
            service = DocumentInsightsService()
            insights: DocumentInsights = service.generate_insights(self.document_text)

        mock_provider.generate_insights.assert_called_once()
        self.assertEqual(insights.summary, expected_response.summary)
        self.assertEqual(insights.keywords, expected_response.keywords)

    @patch("app.service.insights_service.LLMProviderFactory.create_provider")
    def test_generate_insights_with_gemini_provider(self, mock_create_provider):
        """Test insights generation with Gemini provider (backward compatibility)."""
        expected_response = InsightResponse(
            summary="Gemini generated summary.",
            keywords=["gemini", "legacy", "test"]
        )

        mock_provider = MagicMock()
        mock_provider.generate_insights.return_value = expected_response
        mock_create_provider.return_value = mock_provider

        with patch.object(settings, "llm_provider", "gemini"), \
             patch.object(settings, "gemini_api_key", "test-key"):
            service = DocumentInsightsService()
            insights: DocumentInsights = service.generate_insights(self.document_text)

        mock_provider.generate_insights.assert_called_once()
        self.assertEqual(insights.summary, expected_response.summary)
        self.assertEqual(insights.keywords, expected_response.keywords)

    @patch("app.service.insights_service.LLMProviderFactory.create_provider")
    def test_generate_insights_handles_provider_error(self, mock_create_provider):
        """Test error handling when provider fails."""
        mock_provider = MagicMock()
        mock_provider.generate_insights.side_effect = RuntimeError("Provider error")
        mock_create_provider.return_value = mock_provider

        with patch.object(settings, "llm_provider", "openai"), \
             patch.object(settings, "openai_api_key", "test-key"):
            service = DocumentInsightsService()

            with self.assertRaises(RuntimeError) as context:
                service.generate_insights(self.document_text)

            self.assertIn("Failed to generate document insights", str(context.exception))

    @patch("app.service.insights_service.LLMProviderFactory.create_provider")
    def test_fallback_to_openai_on_invalid_provider(self, mock_create_provider):
        """Test fallback to OpenAI when invalid provider is specified."""
        mock_provider = MagicMock()
        mock_create_provider.return_value = mock_provider

        with patch.object(settings, "llm_provider", "invalid_provider"), \
             patch.object(settings, "openai_api_key", "test-key"):
            service = DocumentInsightsService()

            # Should fall back to OpenAI without raising error
            self.assertIsNotNone(service._provider)


if __name__ == "__main__":
    unittest.main()

