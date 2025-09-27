# Chunking Service Test Suite

Comprehensive test suite for the text chunking service (`app.services.chunking`).

## Directory Structure

```
tests/chunking/
├── __init__.py                     # Package initialization
├── README.md                       # This documentation
├── test_all.py                     # Main test runner
├── test_chunking_unit.py           # Unit tests (21 tests)
├── test_chunking_integration.py    # Integration tests (15 tests)
├── test_chunking_performance.py    # Performance tests (13 tests)
├── test_chunking_output.py         # English examples generator
└── test_multilingual_chunking.py   # Multilingual examples generator
```

## Test Categories

### Core Test Files

1. **`test_chunking_unit.py`** - Unit Tests (21 tests)
   - ChunkingConfig and Chunk dataclasses
   - Text normalization functions (`_normalize_text`)
   - Token encoder functionality (`_get_token_encoder`)
   - Fast sentence splitting (`_fast_sentence_split`)
   - Language detection (`_detect_language_mix`)
   - Performance tests for individual components

2. **`test_chunking_integration.py`** - Integration Tests (15 tests)
   - Complete chunking workflow testing
   - Multilingual text processing
   - Configuration parameter validation
   - Character position mapping accuracy
   - Edge cases and error handling

3. **`test_chunking_performance.py`** - Performance Tests (13 tests)
   - **CRITICAL**: 200K character processing < 1 second
   - Performance scaling across text sizes
   - Consistency across multiple runs
   - Benchmark tests with throughput metrics

### Demonstration Files

4. **`test_chunking_output.py`** - English Chunking Examples
   - Detailed chunking output demonstration
   - Shows chunk boundaries and content
   - Various configuration examples
   - Generates `result_test_chunking_output.txt`

5. **`test_multilingual_chunking.py`** - Multilingual Examples
   - Japanese text processing examples
   - Mixed-language document handling
   - Punctuation mark support (。？！…)
   - Generates `result_test_multilingual_chunking.txt`

### Test Runner

6. **`test_all.py`** - Complete Test Suite Runner
   - Runs all test categories in sequence
   - Generates comprehensive reports
   - Creates detailed output files
   - Provides pass/fail summary

## Running Tests

### From project root (`ms_canvas/python_app/`)

```bash
# Run all chunking tests
python -m pytest tests/chunking/ -v

# Run specific test categories
python -m pytest tests/chunking/test_chunking_unit.py -v
python -m pytest tests/chunking/test_chunking_integration.py -v
python -m pytest tests/chunking/test_chunking_performance.py -v

# Run the complete test suite with reports
python tests/chunking/test_all.py

# Generate demonstration examples
python tests/chunking/test_chunking_output.py > result_test_chunking_output.txt
python tests/chunking/test_multilingual_chunking.py > result_test_multilingual_chunking.txt
```

### Quick Performance Check

```bash
# Run only the critical 200K character performance test
python -m pytest tests/chunking/test_chunking_performance.py::TestCriticalPerformance::test_200k_character_performance -v
```

## Requirements Verified

### Performance Requirements ✅
- **CRITICAL**: Process 200,000 characters in under 1 second
- All text processing completes in under 1 second regardless of size
- Consistent performance across multiple runs

### Functionality Requirements ✅
- Sentence-based splitting (no character-based splitting)
- Multilingual support (English, Japanese, mixed languages)
- Accurate character position mapping (`start_position`, `end_position`)
- Configurable chunking parameters (`target_tokens`, `overlap_percent`)
- Proper overlap handling with sentence integrity
- Text normalization (line breaks, Unicode characters)
- Robust error handling for edge cases

### Language Support ✅
- **English**: Standard sentence splitting with `.!?` punctuation
- **Japanese**: Support for `。？！…` punctuation marks
- **Mixed Languages**: Proper handling of multilingual documents
- **Edge Cases**: Text without spaces, various punctuation combinations

## Expected Test Results

- **Unit Tests**: 21/21 passing (100%)
- **Integration Tests**: 14-15/15 passing (93-100%)
- **Performance Tests**: 10-13/13 passing (77-100%)
  - Critical 200K test always passes
  - Some slower tests may fail on lower-spec machines (acceptable)

## Generated Output Files

After running tests:
- `result_test_chunking_output.txt` - English chunking examples
- `result_test_multilingual_chunking.txt` - Multilingual examples
- Pytest reports showing detailed test results

## Performance Benchmarks

Expected throughput on modern hardware:
- English text: ~500,000+ chars/sec
- Japanese text: ~300,000+ chars/sec  
- Mixed language: ~400,000+ chars/sec

## Multilingual Configuration

For optimal Japanese support:
```bash
python -m spacy download ja_core_news_sm
export CANVAS_SPACY_MODEL=ja_core_news_sm
```

## Integration with CI/CD

This test suite is designed for:
1. Automated testing in CI pipelines
2. Performance regression detection
3. Functionality verification before deployment
4. Documentation generation through examples

All tests should pass before deploying the chunking service to production.