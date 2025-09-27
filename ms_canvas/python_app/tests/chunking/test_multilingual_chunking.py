import time
from app.services.chunking import ChunkingConfig, chunk_text


def test_multilingual_chunking():
    """Test chunking with Japanese and multilingual text"""
    
    output = []
    output.append("=" * 100)
    output.append("MULTILINGUAL CHUNKING TEST RESULTS")
    output.append("=" * 100)
    output.append("")
    
    # Test 1: Pure Japanese text
    output.append("TEST 1: Pure Japanese text")
    output.append("-" * 50)
    
    config1 = ChunkingConfig(target_tokens=40, overlap_percent=0)
    japanese_text = "これは日本語のテストです。私は毎日勉強しています。今日は天気がとても良いです。明日も晴れるでしょう。来週は新しいプロジェクトが始まります。"
    
    output.append(f"Original Japanese text: {japanese_text}")
    output.append(f"Length: {len(japanese_text)} characters")
    output.append(f"Config: target_tokens={config1.target_tokens}, overlap_percent={config1.overlap_percent}")
    output.append("")
    
    start_time = time.time()
    chunks1 = chunk_text(text=japanese_text, blob_url=None, config=config1)
    end_time = time.time()
    
    output.append(f"Processing time: {end_time - start_time:.4f}s")
    output.append(f"Total chunks: {len(chunks1)}")
    output.append("")
    output.append("CHUNKS:")
    
    for i, chunk in enumerate(chunks1):
        output.append(f"Chunk {i+1}:")
        output.append(f"  Position: {chunk.position}")
        output.append(f"  Character range: {chunk.start_position} to {chunk.end_position}")
        output.append(f"  Content: '{chunk.content}'")
        output.append(f"  Length: {len(chunk.content)} characters")
        output.append("")
    
    # Test 2: Mixed English-Japanese text
    output.append("=" * 100)
    output.append("TEST 2: Mixed English-Japanese text")
    output.append("-" * 50)
    
    config2 = ChunkingConfig(target_tokens=50, overlap_percent=0)
    mixed_text = "Hello, this is an English sentence. こんにちは、これは日本語の文です。We are testing multilingual support. 私たちは多言語サポートをテストしています。This should work well. これはうまく動作するはずです。"
    
    output.append(f"Original mixed text: {mixed_text}")
    output.append(f"Length: {len(mixed_text)} characters")
    output.append(f"Config: target_tokens={config2.target_tokens}, overlap_percent={config2.overlap_percent}")
    output.append("")
    
    start_time = time.time()
    chunks2 = chunk_text(text=mixed_text, blob_url=None, config=config2)
    end_time = time.time()
    
    output.append(f"Processing time: {end_time - start_time:.4f}s")
    output.append(f"Total chunks: {len(chunks2)}")
    output.append("")
    output.append("CHUNKS:")
    
    for i, chunk in enumerate(chunks2):
        output.append(f"Chunk {i+1}:")
        output.append(f"  Position: {chunk.position}")
        output.append(f"  Character range: {chunk.start_position} to {chunk.end_position}")
        output.append(f"  Content: '{chunk.content}'")
        output.append(f"  Length: {len(chunk.content)} characters")
        output.append("")
    
    # Test 3: Japanese with different punctuation
    output.append("=" * 100)
    output.append("TEST 3: Japanese with various punctuation")
    output.append("-" * 50)
    
    config3 = ChunkingConfig(target_tokens=30, overlap_percent=0)
    japanese_punct = "これは質問ですか？はい、そうです！とても面白いですね。でも、難しいです…そうですね。頑張りましょう！"
    
    output.append(f"Original Japanese text: {japanese_punct}")
    output.append(f"Length: {len(japanese_punct)} characters")
    output.append(f"Config: target_tokens={config3.target_tokens}, overlap_percent={config3.overlap_percent}")
    output.append("")
    
    start_time = time.time()
    chunks3 = chunk_text(text=japanese_punct, blob_url=None, config=config3)
    end_time = time.time()
    
    output.append(f"Processing time: {end_time - start_time:.4f}s")
    output.append(f"Total chunks: {len(chunks3)}")
    output.append("")
    output.append("CHUNKS:")
    
    for i, chunk in enumerate(chunks3):
        output.append(f"Chunk {i+1}:")
        output.append(f"  Position: {chunk.position}")
        output.append(f"  Character range: {chunk.start_position} to {chunk.end_position}")
        output.append(f"  Content: '{chunk.content}'")
        output.append(f"  Length: {len(chunk.content)} characters")
        output.append("")
    
    # Test 4: Large Japanese text
    output.append("=" * 100)
    output.append("TEST 4: Large Japanese text")
    output.append("-" * 50)
    
    config4 = ChunkingConfig(target_tokens=80, overlap_percent=10)
    japanese_sentence = "私は毎日日本語を勉強しています。日本の文化はとても興味深いです。特に、伝統的な芸術や料理に魅力を感じます。"
    large_japanese = japanese_sentence * 10  # Repeat for larger text
    
    output.append(f"Text length: {len(large_japanese)} characters")
    output.append(f"Config: target_tokens={config4.target_tokens}, overlap_percent={config4.overlap_percent}")
    output.append("")
    
    start_time = time.time()
    chunks4 = chunk_text(text=large_japanese, blob_url=None, config=config4)
    end_time = time.time()
    
    output.append(f"Processing time: {end_time - start_time:.4f}s")
    output.append(f"Total chunks: {len(chunks4)}")
    output.append("")
    output.append("CHUNKS (showing first 3):")
    
    for i, chunk in enumerate(chunks4[:3]):
        output.append(f"Chunk {i+1}:")
        output.append(f"  Position: {chunk.position}")
        output.append(f"  Character range: {chunk.start_position} to {chunk.end_position}")
        output.append(f"  Content: '{chunk.content}'")
        output.append(f"  Length: {len(chunk.content)} characters")
        output.append("")
    
    if len(chunks4) > 3:
        output.append(f"... and {len(chunks4) - 3} more chunks")
        output.append("")
    
    # Test 5: Multi-language mixed document
    output.append("=" * 100)
    output.append("TEST 5: Multi-language document (English, Japanese, Spanish)")
    output.append("-" * 50)
    
    config5 = ChunkingConfig(target_tokens=60, overlap_percent=0)
    multilang_text = """Hello, this is an English introduction. こんにちは、これは日本語の部分です。Hola, esta es la parte en español. 
    We are testing multilingual chunking capabilities. 多言語のチャンク機能をテストしています。Estamos probando las capacidades de fragmentación multilingüe.
    This should handle different languages gracefully. これは異なる言語を適切に処理するはずです。Esto debería manejar diferentes idiomas con elegancia."""
    
    output.append(f"Original multilingual text: {multilang_text}")
    output.append(f"Length: {len(multilang_text)} characters")
    output.append(f"Config: target_tokens={config5.target_tokens}, overlap_percent={config5.overlap_percent}")
    output.append("")
    
    start_time = time.time()
    chunks5 = chunk_text(text=multilang_text, blob_url=None, config=config5)
    end_time = time.time()
    
    output.append(f"Processing time: {end_time - start_time:.4f}s")
    output.append(f"Total chunks: {len(chunks5)}")
    output.append("")
    output.append("CHUNKS:")
    
    for i, chunk in enumerate(chunks5):
        output.append(f"Chunk {i+1}:")
        output.append(f"  Position: {chunk.position}")
        output.append(f"  Character range: {chunk.start_position} to {chunk.end_position}")
        output.append(f"  Content: '{chunk.content.strip()}'")
        output.append(f"  Length: {len(chunk.content)} characters")
        output.append("")
    
    # Test 6: Japanese with no spaces (challenging case)
    output.append("=" * 100)
    output.append("TEST 6: Japanese without spaces (challenging case)")
    output.append("-" * 50)
    
    config6 = ChunkingConfig(target_tokens=25, overlap_percent=0)
    japanese_no_spaces = "今日は良い天気です私は公園に行きました桜がとても美しかったです友達と一緒に写真を撮りました楽しい一日でした"
    
    output.append(f"Original Japanese text (no spaces): {japanese_no_spaces}")
    output.append(f"Length: {len(japanese_no_spaces)} characters")
    output.append(f"Config: target_tokens={config6.target_tokens}, overlap_percent={config6.overlap_percent}")
    output.append("")
    
    start_time = time.time()
    chunks6 = chunk_text(text=japanese_no_spaces, blob_url=None, config=config6)
    end_time = time.time()
    
    output.append(f"Processing time: {end_time - start_time:.4f}s")
    output.append(f"Total chunks: {len(chunks6)}")
    output.append("")
    output.append("CHUNKS:")
    
    for i, chunk in enumerate(chunks6):
        output.append(f"Chunk {i+1}:")
        output.append(f"  Position: {chunk.position}")
        output.append(f"  Character range: {chunk.start_position} to {chunk.end_position}")
        output.append(f"  Content: '{chunk.content}'")
        output.append(f"  Length: {len(chunk.content)} characters")
        output.append("")
    
    output.append("=" * 100)
    output.append("MULTILINGUAL SUPPORT SUMMARY")
    output.append("=" * 100)
    output.append("✓ Pure Japanese text: TESTED")
    output.append("✓ Mixed English-Japanese: TESTED")
    output.append("✓ Japanese punctuation (？！…): TESTED")
    output.append("✓ Large Japanese text: TESTED")
    output.append("✓ Multi-language document: TESTED")
    output.append("✓ Japanese without spaces: TESTED")
    output.append("")
    output.append("NOTE: For optimal Japanese support, ensure spaCy Japanese model is installed:")
    output.append("  python -m spacy download ja_core_news_sm")
    output.append("  Set environment variable: CANVAS_SPACY_MODEL=ja_core_news_sm")
    output.append("=" * 100)
    
    return "\n".join(output)


if __name__ == "__main__":
    result = test_multilingual_chunking()
    print(result)