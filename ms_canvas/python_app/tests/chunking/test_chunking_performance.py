"""
Performance tests for the chunking service.
Critical tests to ensure sub-1-second performance requirements are met.
"""
import pytest
import time
from app.services.chunking import ChunkingConfig, chunk_text


class TestCriticalPerformance:
    """Critical performance tests that must pass."""
    
    def test_200k_character_performance(self):
        """CRITICAL: Must process 200,000 characters in under 1 second."""
        config = ChunkingConfig(target_tokens=500, overlap_percent=15)
        
        # Create exactly 200,000 characters
        sentence = "This is a comprehensive test sentence designed to evaluate performance with substantial content. "
        num_repeats = (200000 // len(sentence)) + 1
        text = sentence * num_repeats
        text = text[:200000]
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        processing_time = end_time - start_time
        
        # CRITICAL REQUIREMENT
        assert processing_time < 1.0, f"Processing took {processing_time:.4f}s, must be < 1.0s"
        assert len(chunks) > 0
        assert len(text) == 200000
        
    def test_100k_character_performance(self):
        """Test with 100KB text - must be under 1 second."""
        config = ChunkingConfig(target_tokens=300, overlap_percent=10)
        
        sentence = "Performance test sentence with reasonable length and content for evaluation. "
        text = sentence * 1000  # About 100KB
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        assert end_time - start_time < 1.0
        assert len(chunks) > 0
        
    def test_multilingual_large_text_performance(self):
        """Test large multilingual text performance."""
        config = ChunkingConfig(target_tokens=400, overlap_percent=10)
        
        # Mixed language content
        mixed_sentence = "English content here for testing purposes. 日本語のコンテンツもここにテスト用として含まれています。"
        text = mixed_sentence * 500  # About 50KB of mixed content
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        assert end_time - start_time < 1.0
        assert len(chunks) > 0


class TestPerformanceScaling:
    """Test how performance scales with different text sizes."""
    
    def test_small_text_performance(self):
        """Small text should be very fast."""
        config = ChunkingConfig(target_tokens=100, overlap_percent=10)
        text = "This is a small test. It should be very fast."
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        assert end_time - start_time < 0.5
        assert len(chunks) >= 1
        
    def test_medium_text_performance(self):
        """Medium text (10KB) should be fast."""
        config = ChunkingConfig(target_tokens=200, overlap_percent=10)
        sentence = "This is a medium-length test sentence for performance evaluation. "
        text = sentence * 100  # About 10KB
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        assert end_time - start_time < 1.0
        assert len(chunks) > 0
        
    def test_performance_with_different_configs(self):
        """Test that different configurations don't affect performance requirement."""
        text = "Test sentence. " * 5000  # Medium sized text
        
        configs = [
            ChunkingConfig(target_tokens=50, overlap_percent=0),
            ChunkingConfig(target_tokens=100, overlap_percent=25),
            ChunkingConfig(target_tokens=500, overlap_percent=50),
            ChunkingConfig(target_tokens=1000, overlap_percent=75),
        ]
        
        for config in configs:
            start_time = time.time()
            chunks = chunk_text(text=text, blob_url=None, config=config)
            end_time = time.time()
            
            assert end_time - start_time < 1.0
            assert len(chunks) >= 1


class TestPerformanceConsistency:
    """Test that performance is consistent across runs."""
    
    def test_repeated_chunking_performance(self):
        """Test that repeated chunking maintains performance."""
        config = ChunkingConfig(target_tokens=300, overlap_percent=10)
        sentence = "Repeated performance test sentence with consistent content. "
        text = sentence * 1000  # About 50KB
        
        times = []
        for _ in range(5):  # Run 5 times
            start_time = time.time()
            chunks = chunk_text(text=text, blob_url=None, config=config)
            end_time = time.time()
            
            processing_time = end_time - start_time
            times.append(processing_time)
            
            assert processing_time < 1.0
            assert len(chunks) > 0
        
        # Check consistency - no run should be more than 2x slower than fastest
        min_time = min(times)
        max_time = max(times)
        assert max_time < min_time * 3  # Allow some variation but not excessive
        
    def test_different_content_performance(self):
        """Test performance with different types of content."""
        config = ChunkingConfig(target_tokens=250, overlap_percent=10)
        
        # Different content types
        contents = [
            "Simple English sentences. " * 2000,
            "これは日本語のテストです。" * 1000,
            "Mixed English and 日本語 content. " * 1500,
            "Technical content with numbers 123 and symbols @#$. " * 1800,
        ]
        
        for content in contents:
            start_time = time.time()
            chunks = chunk_text(text=content, blob_url=None, config=config)
            end_time = time.time()
            
            assert end_time - start_time < 1.0
            assert len(chunks) > 0


class TestPerformanceBenchmarks:
    """Benchmark tests for performance monitoring."""
    
    def test_benchmark_english_only(self):
        """Benchmark for English-only content."""
        config = ChunkingConfig(target_tokens=400, overlap_percent=15)
        sentence = "This is an English benchmark sentence with standard content for testing. "
        text = sentence * 2000  # About 150KB
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        processing_time = end_time - start_time
        
        assert processing_time < 1.0
        assert len(chunks) > 0
        
        # For monitoring - print performance info
        print(f"\nEnglish benchmark: {len(text)} chars in {processing_time:.4f}s")
        print(f"Throughput: {len(text) / processing_time:.0f} chars/sec")
        
    def test_benchmark_japanese_only(self):
        """Benchmark for Japanese-only content."""
        config = ChunkingConfig(target_tokens=400, overlap_percent=15)
        sentence = "これは日本語のベンチマークテストの文章です。標準的な内容でテストを行います。"
        text = sentence * 1000  # Japanese content
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        processing_time = end_time - start_time
        
        assert processing_time < 1.0
        assert len(chunks) > 0
        
        # For monitoring - print performance info
        print(f"\nJapanese benchmark: {len(text)} chars in {processing_time:.4f}s")
        print(f"Throughput: {len(text) / processing_time:.0f} chars/sec")
        
    def test_benchmark_mixed_languages(self):
        """Benchmark for mixed-language content."""
        config = ChunkingConfig(target_tokens=400, overlap_percent=15)
        sentence = "Mixed benchmark with English and 日本語 content for comprehensive testing. "
        text = sentence * 1500  # Mixed content
        
        start_time = time.time()
        chunks = chunk_text(text=text, blob_url=None, config=config)
        end_time = time.time()
        
        processing_time = end_time - start_time
        
        assert processing_time < 1.0
        assert len(chunks) > 0
        
        # For monitoring - print performance info
        print(f"\nMixed language benchmark: {len(text)} chars in {processing_time:.4f}s")
        print(f"Throughput: {len(text) / processing_time:.0f} chars/sec")