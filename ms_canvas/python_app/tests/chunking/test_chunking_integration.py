"""
Integration tests for the chunking service.
Tests the full chunking workflow with real text processing.
"""
import pytest
import time
from app.services.chunking import ChunkingConfig, chunk_text


class TestBasicChunking:
    """Test basic chunking functionality."""
    
    def test_requires_text_or_url(self):
        config = ChunkingConfig()
        with pytest.raises(ValueError, match="Either text or blob_url must be provided"):
            chunk_text(text=None, blob_url=None, config=config)

    def test_empty_text_handling(self):
        config = ChunkingConfig()
        chunks = chunk_text(text="   ", blob_url=None, config=config)
        assert len(chunks) == 0

    def test_simple_sentence_chunking(self):
        config = ChunkingConfig(target_tokens=50, overlap_percent=0)
        text = "This is the first sentence. This is the second sentence. This is the third sentence."
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        assert len(chunks) > 0
        for i, chunk in enumerate(chunks):
            assert chunk.position == i
            assert chunk.start_position >= 0
            assert chunk.end_position > chunk.start_position
            assert len(chunk.content) > 0
            assert chunk.start_position < len(text)
            assert chunk.end_position <= len(text)


class TestMultilingualChunking:
    """Test multilingual chunking capabilities."""
    
    def test_japanese_text(self):
        config = ChunkingConfig(target_tokens=40, overlap_percent=0)
        text = "これは日本語のテストです。私は毎日勉強しています。今日は天気がとても良いです。"
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        assert len(chunks) >= 1
        assert all(chunk.content.strip() for chunk in chunks)
        
    def test_mixed_language_text(self):
        config = ChunkingConfig(target_tokens=50, overlap_percent=0)
        text = "Hello, this is English. こんにちは、これは日本語です。We are testing multilingual support."
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        assert len(chunks) >= 1
        total_content = " ".join(chunk.content for chunk in chunks)
        assert "Hello" in total_content
        assert "こんにちは" in total_content
        
    def test_japanese_punctuation(self):
        config = ChunkingConfig(target_tokens=30, overlap_percent=0)
        text = "これは質問ですか？はい、そうです！とても面白いですね。"
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        assert len(chunks) >= 1
        assert all(chunk.content.strip() for chunk in chunks)


class TestChunkingConfiguration:
    """Test different chunking configurations."""
    
    def test_different_target_tokens(self):
        text = "This is a test sentence. " * 10
        
        configs = [
            ChunkingConfig(target_tokens=20, overlap_percent=0),
            ChunkingConfig(target_tokens=50, overlap_percent=0),
            ChunkingConfig(target_tokens=100, overlap_percent=0),
        ]
        
        results = []
        for config in configs:
            chunks = chunk_text(text=text, blob_url=None, config=config)
            results.append(len(chunks))
        
        # Higher target tokens should generally result in fewer chunks
        assert results[0] >= results[1] >= results[2]
        
    def test_overlap_functionality(self):
        config_no_overlap = ChunkingConfig(target_tokens=30, overlap_percent=0)
        config_with_overlap = ChunkingConfig(target_tokens=30, overlap_percent=25)
        text = "First sentence. Second sentence. Third sentence. Fourth sentence. Fifth sentence."
        
        chunks_no_overlap = chunk_text(text=text, blob_url=None, config=config_no_overlap)
        chunks_with_overlap = chunk_text(text=text, blob_url=None, config=config_with_overlap)
        
        if len(chunks_no_overlap) > 1 and len(chunks_with_overlap) > 1:
            # With overlap, total content should be larger due to repeated content
            total_no_overlap = sum(len(chunk.content) for chunk in chunks_no_overlap)
            total_with_overlap = sum(len(chunk.content) for chunk in chunks_with_overlap)
            assert total_with_overlap >= total_no_overlap


class TestCharacterPositionMapping:
    """Test accuracy of character position mapping."""
    
    def test_position_accuracy(self):
        config = ChunkingConfig(target_tokens=20, overlap_percent=0)
        text = "First sentence. Second sentence. Third sentence."
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        for chunk in chunks:
            extracted = text[chunk.start_position:chunk.end_position]
            normalized_extracted = extracted.replace('\r\n', '\n').replace('\r', '\n').replace('\u2028', '\n').replace('\u2029', '\n')
            assert chunk.content == normalized_extracted
            
    def test_position_bounds(self):
        config = ChunkingConfig(target_tokens=30, overlap_percent=0)
        text = "Test sentence one. Test sentence two. Test sentence three."
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        for chunk in chunks:
            assert 0 <= chunk.start_position < len(text)
            assert chunk.start_position < chunk.end_position <= len(text)
            
    def test_multilingual_position_accuracy(self):
        config = ChunkingConfig(target_tokens=25, overlap_percent=0)
        text = "English sentence. 日本語の文章です。Another English sentence."
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        for chunk in chunks:
            extracted = text[chunk.start_position:chunk.end_position]
            # Basic check that positions make sense
            assert len(extracted) > 0
            assert chunk.start_position >= 0
            assert chunk.end_position <= len(text)


class TestPerformanceIntegration:
    """Test performance requirements in integration scenarios."""
    
    def test_medium_text_performance(self):
        config = ChunkingConfig(target_tokens=100, overlap_percent=10)
        sentence = "This is a test sentence for performance evaluation. "
        text = sentence * 200  # About 10KB
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        assert end_time - start_time < 1.0
        assert len(chunks) > 0
        
    def test_large_text_performance(self):
        config = ChunkingConfig(target_tokens=300, overlap_percent=15)
        sentence = "Performance test sentence with reasonable length. "
        text = sentence * 2000  # About 100KB
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        assert end_time - start_time < 1.0
        assert len(chunks) > 0
        
    def test_multilingual_performance(self):
        config = ChunkingConfig(target_tokens=200, overlap_percent=10)
        mixed_sentence = "English sentence here. 日本語の文章がここにあります。"
        text = mixed_sentence * 500  # Mixed language text
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        assert end_time - start_time < 1.0
        assert len(chunks) > 0


class TestEdgeCases:
    """Test edge cases and error conditions."""
    
    def test_single_very_long_sentence(self):
        config = ChunkingConfig(target_tokens=50, overlap_percent=0)
        text = "This is a very long sentence that goes on and on " * 100 + "and finally ends here."
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        assert len(chunks) >= 1
        assert all(len(chunk.content) > 0 for chunk in chunks)
        
    def test_many_short_sentences(self):
        config = ChunkingConfig(target_tokens=30, overlap_percent=0)
        text = "Short. " * 100
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        assert len(chunks) >= 1
        assert all(len(chunk.content.strip()) > 0 for chunk in chunks)
        
    def test_mixed_punctuation_edge_cases(self):
        config = ChunkingConfig(target_tokens=40, overlap_percent=0)
        text = "Question? Answer! Statement. Ellipsis... 日本語？日本語！日本語。"
        
        chunks = chunk_text(text=text, blob_url=None, config=config)
        
        assert len(chunks) >= 1
        assert all(chunk.content.strip() for chunk in chunks)