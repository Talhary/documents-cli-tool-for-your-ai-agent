package textops

import (
	"strings"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."
	est := EstimateTokens(text)

	if est.Characters != len(text) {
		t.Errorf("expected chars %d, got %d", len(text), est.Characters)
	}
	if est.Words != 9 {
		t.Errorf("expected 9 words, got %d", est.Words)
	}
	if est.Tokens <= 0 {
		t.Errorf("expected positive token estimate, got %d", est.Tokens)
	}
}

func TestChunkText(t *testing.T) {
	lines := []string{
		"Paragraph 1 line 1.",
		"Paragraph 1 line 2.",
		"",
		"Paragraph 2 line 1.",
		"Paragraph 2 line 2.",
		"Paragraph 2 line 3.",
		"",
		"Paragraph 3 line 1.",
	}
	content := strings.Join(lines, "\n")

	res := ChunkText(content, ChunkOptions{
		MaxTokens:     10,
		OverlapTokens: 0,
		BySentence:    false,
	})

	if res.TotalChunks <= 1 {
		t.Errorf("expected multiple chunks, got %d", res.TotalChunks)
	}

	for i, c := range res.Chunks {
		if c.Text == "" {
			t.Errorf("chunk %d is empty", i)
		}
		if c.Tokens <= 0 {
			t.Errorf("chunk %d has invalid token count: %d", i, c.Tokens)
		}
	}
}
