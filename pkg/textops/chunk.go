package textops

import (
	"strings"
	"unicode"
)

// TokenEstimate holds an approximate token count for a body of text. The
// estimate uses a heuristic (word and punctuation segmentation blended with a
// ~4-characters-per-token ratio) that approximates modern BPE tokenizers such
// as OpenAI tiktoken and Anthropic's tokenizer without external dependencies.
type TokenEstimate struct {
	Characters int `json:"characters"`
	Words      int `json:"words"`
	Tokens     int `json:"tokens"`
}

// EstimateTokens approximates the number of LLM tokens in a string.
func EstimateTokens(text string) TokenEstimate {
	chars := len([]rune(text))
	words := len(strings.Fields(text))

	// Blend two common heuristics for a stable estimate:
	//  - character ratio: ~4 chars per token
	//  - word ratio:      ~1.3 tokens per word (accounts for sub-word splits)
	charBased := float64(chars) / 4.0
	wordBased := float64(words) * 1.3
	tokens := int((charBased + wordBased) / 2.0)

	if tokens == 0 && chars > 0 {
		tokens = 1
	}

	return TokenEstimate{
		Characters: chars,
		Words:      words,
		Tokens:     tokens,
	}
}

// Chunk represents a single contiguous slice of text produced by chunking.
type Chunk struct {
	Index      int    `json:"index"`
	Text       string `json:"text"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
	Characters int    `json:"characters"`
	Tokens     int    `json:"tokens"`
}

// ChunkOptions configures how text is split into chunks.
type ChunkOptions struct {
	// MaxTokens is the approximate token ceiling per chunk.
	MaxTokens int
	// OverlapTokens is the number of tokens of overlap carried into the next
	// chunk to preserve context across boundaries.
	OverlapTokens int
	// BySentence, when true, avoids splitting in the middle of a sentence.
	BySentence bool
}

// ChunkResult aggregates the outcome of a chunking operation.
type ChunkResult struct {
	TotalChunks int     `json:"total_chunks"`
	MaxTokens   int     `json:"max_tokens"`
	Overlap     int     `json:"overlap_tokens"`
	TotalTokens int     `json:"total_tokens"`
	Chunks      []Chunk `json:"chunks"`
}

// ChunkText splits text into token-bounded chunks with optional overlap. Line
// numbers are tracked so callers can map a chunk back to its source location.
func ChunkText(text string, opts ChunkOptions) ChunkResult {
	if opts.MaxTokens <= 0 {
		opts.MaxTokens = 512
	}
	if opts.OverlapTokens < 0 {
		opts.OverlapTokens = 0
	}
	if opts.OverlapTokens >= opts.MaxTokens {
		opts.OverlapTokens = opts.MaxTokens / 4
	}

	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	// Build atomic units: sentences (or lines) tagged with their line number.
	type unit struct {
		text   string
		line   int
		tokens int
	}
	var units []unit
	for i, line := range lines {
		lineNo := i + 1
		if opts.BySentence {
			for _, s := range splitSentences(line) {
				if strings.TrimSpace(s) == "" {
					continue
				}
				units = append(units, unit{text: s, line: lineNo, tokens: EstimateTokens(s).Tokens})
			}
		} else {
			units = append(units, unit{text: line, line: lineNo, tokens: EstimateTokens(line).Tokens})
		}
	}

	var chunks []Chunk
	totalTokens := 0
	i := 0
	for i < len(units) {
		var b strings.Builder
		startLine := units[i].line
		endLine := units[i].line
		curTokens := 0
		j := i
		for j < len(units) {
			u := units[j]
			if curTokens > 0 && curTokens+u.tokens > opts.MaxTokens {
				break
			}
			if b.Len() > 0 {
				if opts.BySentence {
					b.WriteString(" ")
				} else {
					b.WriteString("\n")
				}
			}
			b.WriteString(u.text)
			curTokens += u.tokens
			endLine = u.line
			j++
		}

		chunkText := b.String()
		est := EstimateTokens(chunkText)
		chunks = append(chunks, Chunk{
			Index:      len(chunks),
			Text:       chunkText,
			StartLine:  startLine,
			EndLine:    endLine,
			Characters: len([]rune(chunkText)),
			Tokens:     est.Tokens,
		})
		totalTokens += est.Tokens

		if j >= len(units) {
			break
		}

		// Advance, applying token-based overlap by walking backwards.
		if opts.OverlapTokens > 0 {
			overlap := 0
			back := j
			for back > i+1 {
				overlap += units[back-1].tokens
				if overlap >= opts.OverlapTokens {
					break
				}
				back--
			}
			i = back
		} else {
			i = j
		}
	}

	return ChunkResult{
		TotalChunks: len(chunks),
		MaxTokens:   opts.MaxTokens,
		Overlap:     opts.OverlapTokens,
		TotalTokens: totalTokens,
		Chunks:      chunks,
	}
}

// splitSentences performs lightweight sentence segmentation on a single line.
func splitSentences(line string) []string {
	var sentences []string
	var b strings.Builder
	runes := []rune(line)
	for idx, r := range runes {
		b.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			// Look ahead: sentence ends if followed by space/EOL.
			if idx+1 >= len(runes) || unicode.IsSpace(runes[idx+1]) {
				sentences = append(sentences, strings.TrimSpace(b.String()))
				b.Reset()
			}
		}
	}
	if strings.TrimSpace(b.String()) != "" {
		sentences = append(sentences, strings.TrimSpace(b.String()))
	}
	return sentences
}
