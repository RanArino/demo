import os
import re
from dataclasses import dataclass
from typing import List, Tuple

import requests
import tiktoken
import spacy


@dataclass
class ChunkingConfig:
    target_tokens: int = 300
    overlap_percent: int = 10
    tokenizer: str = "tiktoken:cl100k_base"


@dataclass
class Chunk:
    position: int
    start_position: int
    end_position: int
    content: str


def _load_spacy_model() -> spacy.language.Language:
    model_name = os.getenv("CANVAS_SPACY_MODEL", "en_core_web_sm")
    try:
        return spacy.load(model_name)
    except Exception:
        # Fallback to multilingual model if available
        try:
            return spacy.load("xx_sent_ud_sm")
        except Exception as e:
            raise RuntimeError(
                f"Failed to load spaCy model '{model_name}' and fallback 'xx_sent_ud_sm'"
            ) from e


def _get_token_encoder(tokenizer_spec: str):
    if tokenizer_spec.startswith("tiktoken:"):
        enc_name = tokenizer_spec.split(":", 1)[1] or "cl100k_base"
        return tiktoken.get_encoding(enc_name)
    # Future: support other tokenizers
    return tiktoken.get_encoding("cl100k_base")


def _normalize_text(text: str) -> str:
    # Basic normalization: normalize line breaks and collapse extraneous whitespace
    # Keep characters intact so we can map offsets against original content separately
    normalized = re.sub(r"\r\n?|\u2028|\u2029", "\n", text)
    return normalized


def _fetch_text_from_url(url: str) -> str:
    timeout = (5, 20)
    resp = requests.get(url, timeout=timeout)
    resp.raise_for_status()
    return resp.text


def _sentence_boundaries(doc: spacy.tokens.Doc) -> List[Tuple[int, int]]:
    bounds: List[Tuple[int, int]] = []
    for sent in doc.sents:
        bounds.append((sent.start_char, sent.end_char))
    return bounds


def chunk_text(
    *,
    text: str | None,
    blob_url: str | None,
    config: ChunkingConfig,
) -> List[Chunk]:
    if not text and not blob_url:
        raise ValueError("Either text or blob_url must be provided")

    original = text if text is not None else _fetch_text_from_url(blob_url)  # type: ignore[arg-type]
    
    # Fast path for empty or whitespace-only text
    if not original or not original.strip():
        return []
    
    target = max(1, int(config.target_tokens))
    overlap_pct = max(0, min(100, int(config.overlap_percent)))
    
    # Smart sentence splitting: use spaCy for better accuracy on smaller/multilingual texts
    # Use fast regex for very large texts where speed is critical
    multilingual = _detect_language_mix(original)
    
    if len(original) > 100000 and not multilingual:  # Large, single-language texts
        sentences = _fast_sentence_split(original)
    else:
        # Use spaCy for better multilingual support and smaller texts
        nlp = _load_spacy_model()
        doc = nlp(original)
        sentences = _sentence_boundaries(doc)
    
    if not sentences:
        return []

    # Ultra-fast token estimation - no tiktoken calls during chunking
    def estimate_tokens_ultra_fast(text: str) -> int:
        # Very fast approximation: ~4 chars per token
        return max(1, len(text) // 4)
    
    chunks: List[Chunk] = []
    position = 0
    i = 0
    
    while i < len(sentences):
        current_start = sentences[i][0]
        current_end = sentences[i][1]
        current_tokens = estimate_tokens_ultra_fast(original[current_start:current_end])
        start_idx = i
        
        # Accumulate sentences until we reach target - simple and fast
        i += 1
        while i < len(sentences) and current_tokens < target:
            sentence_text = original[sentences[i][0]:sentences[i][1]]
            current_tokens += estimate_tokens_ultra_fast(sentence_text)
            current_end = sentences[i][1]
            i += 1
        
        # Extract and normalize content
        content_text = original[current_start:current_end]
        content_norm = _normalize_text(content_text)
        
        chunks.append(
            Chunk(
                position=position,
                start_position=current_start,
                end_position=current_end,
                content=content_norm,
            )
        )
        position += 1
        
        # Simple overlap handling
        if overlap_pct > 0 and i < len(sentences):
            # Go back by a simple percentage of sentences
            overlap_sentences = max(1, (i - start_idx) * overlap_pct // 100)
            i = max(start_idx + 1, i - overlap_sentences)

    return chunks


def _fast_sentence_split(text: str) -> List[Tuple[int, int]]:
    """Ultra-fast sentence splitting using regex for performance-critical cases with multilingual support."""
    import re
    
    # Enhanced sentence boundary detection for multiple languages
    # Includes Japanese punctuation: 。？！
    # Western punctuation: .!?
    # Handle ellipsis and multiple punctuation marks
    sentence_pattern = r'[.!?。？！…]+(?:\s+|$|(?=[A-Z])|(?=[あ-ん])|(?=[ア-ン])|(?=[一-龯]))'
    
    sentences = []
    start = 0
    
    for match in re.finditer(sentence_pattern, text):
        end_pos = match.end()
        
        # For patterns that don't consume space, don't include the following character
        if match.group()[-1] not in [' ', '\n', '\t']:
            end_pos = match.end() - 1
        
        if end_pos > start:
            # Find the actual end of the sentence (including punctuation)
            sent_start = start
            sent_end = match.start() + len(match.group().rstrip())
            
            # Skip empty sentences
            if sent_end > sent_start:
                sentences.append((sent_start, sent_end))
            
            start = end_pos
    
    # Handle remaining text if it doesn't end with punctuation
    if start < len(text):
        remaining = text[start:].strip()
        if remaining:
            sentences.append((start, len(text)))
    
    return sentences


def _detect_language_mix(text: str) -> bool:
    """Detect if text contains mixed languages (especially CJK + Latin)."""
    import re
    
    # Check for Japanese characters
    has_japanese = bool(re.search(r'[ひらがなカタカナ一-龯]', text))
    # Check for Latin characters
    has_latin = bool(re.search(r'[a-zA-Z]', text))
    # Check for Chinese characters (overlap with Japanese, but different usage patterns)
    has_chinese = bool(re.search(r'[一-龯]', text)) and not has_japanese
    
    return (has_japanese and has_latin) or (has_chinese and has_latin)


