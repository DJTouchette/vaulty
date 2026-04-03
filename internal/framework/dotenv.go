package framework

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ParseDotenv parses a .env file from the reader. It handles:
//   - KEY=value pairs
//   - Comments (lines starting with #)
//   - Blank lines
//   - Single-quoted values (literal, no escape processing)
//   - Double-quoted values (with \n, \\, \" escape sequences)
//   - Unquoted values (trimmed)
//   - "export KEY=value" prefix
//
// Limitation: multiline values are not supported.
func ParseDotenv(r io.Reader) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip optional "export " prefix
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
		}

		// Find the first '='
		idx := strings.Index(line, "=")
		if idx < 0 {
			return nil, fmt.Errorf("line %d: missing '=' in %q", lineNum, line)
		}

		key := strings.TrimSpace(line[:idx])
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}

		raw := line[idx+1:]

		val, err := parseDotenvValue(raw)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		result[key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading .env: %w", err)
	}

	return result, nil
}

func parseDotenvValue(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return "", nil
	}

	// Single-quoted: literal value, no escapes
	if strings.HasPrefix(raw, "'") {
		if !strings.HasSuffix(raw, "'") || len(raw) < 2 {
			return "", fmt.Errorf("unterminated single quote")
		}
		return raw[1 : len(raw)-1], nil
	}

	// Double-quoted: process escape sequences
	if strings.HasPrefix(raw, "\"") {
		if !strings.HasSuffix(raw, "\"") || len(raw) < 2 {
			return "", fmt.Errorf("unterminated double quote")
		}
		inner := raw[1 : len(raw)-1]
		return unescapeDoubleQuoted(inner), nil
	}

	// Unquoted: trim and strip inline comments
	if idx := strings.Index(raw, " #"); idx >= 0 {
		raw = strings.TrimSpace(raw[:idx])
	}

	return raw, nil
}

func unescapeDoubleQuoted(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
				i++
			case '\\':
				b.WriteByte('\\')
				i++
			case '"':
				b.WriteByte('"')
				i++
			default:
				b.WriteByte(s[i])
			}
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// WriteDotenv writes secrets in .env format to the writer.
// Keys are sorted alphabetically. If reveal is false, values are masked as "****".
func WriteDotenv(w io.Writer, secrets map[string]string, reveal bool) error {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		val := "****"
		if reveal {
			val = secrets[k]
		}
		if _, err := fmt.Fprintf(w, "%s=%s\n", k, val); err != nil {
			return fmt.Errorf("writing .env: %w", err)
		}
	}
	return nil
}
