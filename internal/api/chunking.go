package api

import "strings"

// SplitMarkdownIntoChunks splits markdown into chunks that are likely safe for the Craft API.
// It prefers splitting on paragraph boundaries ("\n\n"), then line boundaries, and falls back
// to hard byte splits.
func SplitMarkdownIntoChunks(markdown string, maxBytes int) []string {
	if strings.TrimSpace(markdown) == "" {
		return nil
	}
	if maxBytes <= 0 {
		maxBytes = defaultInsertChunkBytes
	}

	// Normalize newlines for consistent chunking.
	md := strings.ReplaceAll(markdown, "\r\n", "\n")
	md = strings.ReplaceAll(md, "\r", "\n")

	paras := strings.Split(md, "\n\n")
	var chunks []string
	var cur strings.Builder
	curLen := 0

	flush := func() {
		text := strings.TrimSpace(cur.String())
		if text != "" {
			chunks = append(chunks, text)
		}
		cur.Reset()
		curLen = 0
	}

	appendWithDelim := func(part, delim string) {
		if curLen == 0 {
			cur.WriteString(part)
			curLen = len(part)
			return
		}
		cur.WriteString(delim)
		cur.WriteString(part)
		curLen += len(delim) + len(part)
	}

	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if len(p) > maxBytes {
			// Flush current chunk first, then split this paragraph.
			flush()
			lines := strings.Split(p, "\n")
			var lineCur strings.Builder
			lineLen := 0
			lineFlush := func() {
				text := strings.TrimSpace(lineCur.String())
				if text != "" {
					chunks = append(chunks, text)
				}
				lineCur.Reset()
				lineLen = 0
			}
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					continue
				}
				if len(line) > maxBytes {
					// Hard split a single oversized line.
					lineFlush()
					b := []byte(line)
					for start := 0; start < len(b); start += maxBytes {
						end := start + maxBytes
						if end > len(b) {
							end = len(b)
						}
						piece := strings.TrimSpace(string(b[start:end]))
						if piece != "" {
							chunks = append(chunks, piece)
						}
					}
					continue
				}

				addLen := len(line)
				if lineLen > 0 {
					addLen += 1 // "\n"
				}
				if lineLen+addLen > maxBytes {
					lineFlush()
				}
				if lineLen > 0 {
					lineCur.WriteString("\n")
					lineLen++
				}
				lineCur.WriteString(line)
				lineLen += len(line)
			}
			lineFlush()
			continue
		}

		addLen := len(p)
		if curLen > 0 {
			addLen += 2 // "\n\n"
		}
		if curLen+addLen > maxBytes {
			flush()
		}
		appendWithDelim(p, "\n\n")
	}

	flush()
	return chunks
}
