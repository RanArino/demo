import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

from app.config.config import settings
from app.domain.insights import DocumentInsights
from app.service.insights_service import DocumentInsightsService


class TestDocumentInsightsService(unittest.TestCase):
    def setUp(self):
        self.fixture_path = Path(__file__).resolve().parent.parent / "fixtures" / "gemini_structured_output.json"
        self.document_text = "Sample PDF content for testing."

    def test_generate_insights_uses_structured_output_schema(self):
        expected_payload = {
            "summary": "Fixture summary describing the sample PDF content for testing structured output.",
            "keywords": ["fixture", "structured output", "summary"],
        }

        fake_response = type("FakeGeneration", (), {"parsed": expected_payload})()

        with patch.object(settings, "gemini_api_key", "test-key"), patch(
            "google.generativeai.configure"
        ) as mock_configure, patch("google.generativeai.GenerativeModel") as mock_model_cls:
            mock_model = MagicMock()
            mock_model.generate_content.return_value = fake_response
            mock_model_cls.return_value = mock_model

            service = DocumentInsightsService()
            insights: DocumentInsights = service.generate_insights(self.document_text, title="Fixture Title")

        mock_configure.assert_called_once()
        mock_model.generate_content.assert_called_once()

        _, kwargs = mock_model.generate_content.call_args
        generation_config = kwargs["generation_config"]
        self.assertEqual(generation_config.response_mime_type, "application/json")

        schema = generation_config.response_schema
        self.assertEqual(schema["type"], "object")
        self.assertIn("summary", schema["properties"])
        self.assertIn("keywords", schema["properties"])
        self.assertEqual(schema["properties"]["keywords"]["max_items"], settings.gemini_keyword_count)

        self.assertEqual(insights.summary, expected_payload["summary"].strip())
        self.assertEqual(insights.keywords, expected_payload["keywords"])


if __name__ == "__main__":
    unittest.main()

