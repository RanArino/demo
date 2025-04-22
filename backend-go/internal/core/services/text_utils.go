package services

import (
	"fmt"
	"regexp"
	"strings"
)

// TODO: Currently uses whitespace splitting; to switch to LLM tokenizer, update here.
func CountTokens(text string) int {
	return len(strings.Fields(text))
}

// SegmentText splits raw text into chunks by grouping sentences until maxTokens is reached.
func SegmentText(raw string, maxTokens int) ([]string, error) {
	if maxTokens <= 0 {
		return nil, fmt.Errorf("maxTokens must be > 0")
	}
	// Split into sentences (RE2-compatible): match sequences ending with punctuation
	re := regexp.MustCompile(`[^.!?]+[.!?]`)
	sentences := re.FindAllString(raw, -1)
	var chunks []string
	var curr []string
	currCount := 0
	for _, sent := range sentences {
		tokCount := CountTokens(sent)
		if tokCount == 0 {
			continue
		}
		// if adding this sentence exceeds maxTokens, start new chunk
		if currCount > 0 && currCount+tokCount > maxTokens {
			chunks = append(chunks, strings.Join(curr, " "))
			curr = []string{sent}
			currCount = tokCount
		} else {
			curr = append(curr, sent)
			currCount += tokCount
		}
	}
	if len(curr) > 0 {
		chunks = append(chunks, strings.Join(curr, " "))
	}
	return chunks, nil
}
