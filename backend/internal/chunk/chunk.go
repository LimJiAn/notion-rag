package chunk

import "strings"

func Text(text string, size, overlap int) []string {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return nil
	}
	if len(clean) <= size {
		return []string{clean}
	}

	step := size - overlap
	chunks := make([]string, 0, len(clean)/step+1)

	for start := 0; start < len(clean); start += step {
		end := start + size
		if end > len(clean) {
			end = len(clean)
		}

		part := strings.TrimSpace(clean[start:end])
		if part != "" {
			chunks = append(chunks, part)
		}
		if end == len(clean) {
			break
		}
	}

	return chunks
}
