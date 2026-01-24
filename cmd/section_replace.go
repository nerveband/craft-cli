package cmd

import (
	"fmt"
	"regexp"
	"strings"
)

var headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)

// replaceSectionByHeading replaces a markdown section identified by a heading.
// If replacement doesn't start with a heading, it will be wrapped with the original heading line.
func replaceSectionByHeading(markdown, heading, replacement string) (string, error) {
	md := strings.ReplaceAll(markdown, "\r\n", "\n")
	md = strings.ReplaceAll(md, "\r", "\n")
	lines := strings.Split(md, "\n")

	target := normalizeHeadingText(heading)
	if target == "" {
		return "", fmt.Errorf("section heading is required")
	}

	inFence := false
	start := -1
	level := 0
	var originalHeadingLine string

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		m := headingRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		text := normalizeHeadingText(m[2])
		if text == target {
			start = i
			level = len(m[1])
			originalHeadingLine = strings.TrimRight(line, " ")
			break
		}
	}
	if start == -1 {
		return "", fmt.Errorf("section heading not found: %s", heading)
	}

	end := len(lines)
	inFence = false
	for i := start + 1; i < len(lines); i++ {
		trim := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trim, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		m := headingRe.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		if len(m[1]) <= level {
			end = i
			break
		}
	}

	repl := strings.TrimSpace(strings.ReplaceAll(replacement, "\r\n", "\n"))
	if repl == "" {
		return "", fmt.Errorf("replacement content is required")
	}

	// If replacement doesn't start with a heading, keep original heading line.
	firstLine := strings.SplitN(repl, "\n", 2)[0]
	if headingRe.FindStringSubmatch(strings.TrimSpace(firstLine)) == nil {
		repl = originalHeadingLine + "\n\n" + repl
	}

	newLines := make([]string, 0, len(lines)-(end-start)+strings.Count(repl, "\n")+1)
	newLines = append(newLines, lines[:start]...)
	newLines = append(newLines, strings.Split(repl, "\n")...)
	newLines = append(newLines, lines[end:]...)

	return strings.TrimSpace(strings.Join(newLines, "\n")) + "\n", nil
}

func normalizeHeadingText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}
