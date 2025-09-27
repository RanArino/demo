"""
Unit tests for the chunking service.
Tests individual components and functions.
"""
import pytest
import time
from unittest.mock import patch, Mock
from app.services.chunking import (
    ChunkingConfig,
    Chunk,
    _normalize_text,
    _get_token_encoder,
    _fast_sentence_split,
    _detect_language_mix,
)


class TestChunkingConfig:
    """Test ChunkingConfig dataclass."""
    
    def test_default_config(self):
        config = ChunkingConfig()
        assert config.target_tokens == 300
        assert config.overlap_percent == 10
        assert config.tokenizer == "tiktoken:cl100k_base"

    def test_custom_config(self):
        config = ChunkingConfig(
            target_tokens=500,
            overlap_percent=20,
            tokenizer="tiktoken:gpt2"
        )
        assert config.target_tokens == 500
        assert config.overlap_percent == 20
        assert config.tokenizer == "tiktoken:gpt2"


class TestChunk:
    """Test Chunk dataclass."""
    
    def test_chunk_creation(self):
        chunk = Chunk(
            position=0,
            start_position=10,
            end_position=50,
            content="Test content"
        )
        assert chunk.position == 0
        assert chunk.start_position == 10
        assert chunk.end_position == 50
        assert chunk.content == "Test content"


class TestNormalizeText:
    """Test text normalization function."""
    
    def test_normalize_line_breaks(self):
        text = "Line 1\r\nLine 2\rLine 3\nLine 4"
        normalized = _normalize_text(text)
        assert normalized == "Line 1\nLine 2\nLine 3\nLine 4"

    def test_normalize_unicode_line_separators(self):
        text = "Line 1\u2028Line 2\u2029Line 3"
        normalized = _normalize_text(text)
        assert normalized == "Line 1\nLine 2\nLine 3"

    def test_preserve_content_integrity(self):
        text = "Hello\tworld\n  with   spaces"
        normalized = _normalize_text(text)
        assert "\t" in normalized
        assert "   " in normalized


class TestTokenEncoder:
    """Test token encoder functionality."""
    
    def test_default_tiktoken_encoder(self):
        encoder = _get_token_encoder("tiktoken:cl100k_base")
        assert encoder is not None
        tokens = encoder.encode("Hello world")
        assert len(tokens) > 0

    def test_gpt2_tiktoken_encoder(self):
        encoder = _get_token_encoder("tiktoken:gpt2")
        assert encoder is not None
        tokens = encoder.encode("Hello world")
        assert len(tokens) > 0

    def test_fallback_to_default(self):
        encoder = _get_token_encoder("invalid:encoder")
        assert encoder is not None
        tokens = encoder.encode("Hello world")
        assert len(tokens) > 0

    def test_empty_tokenizer_spec(self):
        encoder = _get_token_encoder("tiktoken:")
        assert encoder is not None
        tokens = encoder.encode("Hello world")
        assert len(tokens) > 0


class TestFastSentenceSplit:
    """Test fast sentence splitting functionality."""
    
    def test_english_sentences(self):
        text = "First sentence. Second sentence! Third sentence?"
        sentences = _fast_sentence_split(text)
        assert len(sentences) >= 3
        
    def test_japanese_sentences(self):
        text = "これは最初の文です。これは二番目の文です！これは三番目の文ですか？"
        sentences = _fast_sentence_split(text)
        assert len(sentences) >= 3
        
    def test_mixed_punctuation(self):
        text = "English sentence. 日本語の文です。Another English! もう一つの日本語？"
        sentences = _fast_sentence_split(text)
        assert len(sentences) >= 4
        
    def test_ellipsis(self):
        text = "This is incomplete... But this is complete."
        sentences = _fast_sentence_split(text)
        assert len(sentences) >= 2


class TestLanguageDetection:
    """Test language detection functionality."""
    
    def test_english_only(self):
        text = "This is only English text."
        assert not _detect_language_mix(text)
        
    def test_japanese_only(self):
        text = "これは日本語だけのテキストです。"
        assert not _detect_language_mix(text)
        
    def test_mixed_english_japanese(self):
        text = "This is English. これは日本語です。"
        assert _detect_language_mix(text)
        
    def test_mixed_with_chinese(self):
        text = "English text. 中文文本。"
        assert _detect_language_mix(text)


class TestPerformanceRequirements:
    """Test performance requirements for individual components."""
    
    def test_normalize_text_performance(self):
        text = "Line 1\r\nLine 2\rLine 3\u2028Line 4\u2029Line 5\n" * 1000
        
        start_time = time.time()
        normalized = _normalize_text(text)
        end_time = time.time()
        
        assert end_time - start_time < 0.1
        assert "\r" not in normalized
        assert "\u2028" not in normalized
        assert "\u2029" not in normalized
        
    def test_fast_sentence_split_performance(self):
        sentence = "This is a test sentence. "
        text = sentence * 1000  # 1000 sentences
        
        start_time = time.time()
        sentences = _fast_sentence_split(text)
        end_time = time.time()
        
        assert end_time - start_time < 0.1
        assert len(sentences) > 0
        
    def test_language_detection_performance(self):
        text = "English text mixed with 日本語 text. " * 1000
        
        start_time = time.time()
        is_mixed = _detect_language_mix(text)
        end_time = time.time()
        
        assert end_time - start_time < 0.01
        assert is_mixed is True