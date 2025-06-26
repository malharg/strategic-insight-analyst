package processing

const (
	// The target size for each text chunk in characters.
	ChunkSize = 1500
	// The number of characters to overlap between chunks to maintain context.
	ChunkOverlap = 200
)

// ChunkText splits a large text into smaller, overlapping chunks.
func ChunkText(text string) []string {
	// If the text is smaller than the chunk size, return it as a single chunk.
	if len(text) <= ChunkSize {
		return []string{text}
	}

	var chunks []string
	// Use runes to correctly handle multi-byte characters (like emojis or other languages).
	runes := []rune(text)

	for i := 0; i < len(runes); {
		end := i + ChunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunks = append(chunks, string(runes[i:end]))

		// Move the starting point for the next chunk.
		i += ChunkSize - ChunkOverlap

		// A safety check to prevent creating a tiny, mostly-overlapping final chunk.
		if i+ChunkOverlap >= len(runes) {
			break
		}
	}

	return chunks
}
