package framework

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseDotenvBasic(t *testing.T) {
	input := "API_KEY=abc123\nDB_URL=postgres://localhost/mydb\n"
	got, err := ParseDotenv(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["API_KEY"] != "abc123" {
		t.Errorf("API_KEY = %q, want %q", got["API_KEY"], "abc123")
	}
	if got["DB_URL"] != "postgres://localhost/mydb" {
		t.Errorf("DB_URL = %q, want %q", got["DB_URL"], "postgres://localhost/mydb")
	}
}

func TestParseDotenvComments(t *testing.T) {
	input := "# this is a comment\nKEY=val\n# another comment\n"
	got, err := ParseDotenv(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 key, got %d", len(got))
	}
	if got["KEY"] != "val" {
		t.Errorf("KEY = %q, want %q", got["KEY"], "val")
	}
}

func TestParseDotenvBlankLines(t *testing.T) {
	input := "\nKEY=val\n\n\nKEY2=val2\n"
	got, err := ParseDotenv(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 keys, got %d", len(got))
	}
}

func TestParseDotenvSingleQuoted(t *testing.T) {
	input := "SECRET='hello world'\n"
	got, err := ParseDotenv(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["SECRET"] != "hello world" {
		t.Errorf("SECRET = %q, want %q", got["SECRET"], "hello world")
	}
}

func TestParseDotenvDoubleQuoted(t *testing.T) {
	input := `KEY="hello\nworld"` + "\n"
	got, err := ParseDotenv(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["KEY"] != "hello\nworld" {
		t.Errorf("KEY = %q, want %q", got["KEY"], "hello\nworld")
	}
}

func TestParseDotenvDoubleQuotedEscapes(t *testing.T) {
	input := `KEY="say \"hi\" and \\done"` + "\n"
	got, err := ParseDotenv(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `say "hi" and \done`
	if got["KEY"] != want {
		t.Errorf("KEY = %q, want %q", got["KEY"], want)
	}
}

func TestParseDotenvExportPrefix(t *testing.T) {
	input := "export API_KEY=abc123\n"
	got, err := ParseDotenv(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["API_KEY"] != "abc123" {
		t.Errorf("API_KEY = %q, want %q", got["API_KEY"], "abc123")
	}
}

func TestParseDotenvEmptyValue(t *testing.T) {
	input := "KEY=\n"
	got, err := ParseDotenv(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["KEY"] != "" {
		t.Errorf("KEY = %q, want empty", got["KEY"])
	}
}

func TestParseDotenvInlineComment(t *testing.T) {
	input := "KEY=value # this is a comment\n"
	got, err := ParseDotenv(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["KEY"] != "value" {
		t.Errorf("KEY = %q, want %q", got["KEY"], "value")
	}
}

func TestParseDotenvMissingEquals(t *testing.T) {
	input := "NOEQUALS\n"
	_, err := ParseDotenv(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for missing '='")
	}
}

func TestWriteDotenvSorted(t *testing.T) {
	secrets := map[string]string{
		"ZEBRA": "z",
		"APPLE": "a",
		"MANGO": "m",
	}

	var buf bytes.Buffer
	if err := WriteDotenv(&buf, secrets, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "APPLE=a" {
		t.Errorf("line 0 = %q, want %q", lines[0], "APPLE=a")
	}
	if lines[1] != "MANGO=m" {
		t.Errorf("line 1 = %q, want %q", lines[1], "MANGO=m")
	}
	if lines[2] != "ZEBRA=z" {
		t.Errorf("line 2 = %q, want %q", lines[2], "ZEBRA=z")
	}
}

func TestWriteDotenvRedacted(t *testing.T) {
	secrets := map[string]string{
		"KEY": "secret_value",
	}

	var buf bytes.Buffer
	if err := WriteDotenv(&buf, secrets, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(buf.String()) != "KEY=****" {
		t.Errorf("got %q, want %q", strings.TrimSpace(buf.String()), "KEY=****")
	}
}

func TestWriteDotenvRevealed(t *testing.T) {
	secrets := map[string]string{
		"KEY": "secret_value",
	}

	var buf bytes.Buffer
	if err := WriteDotenv(&buf, secrets, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(buf.String()) != "KEY=secret_value" {
		t.Errorf("got %q, want %q", strings.TrimSpace(buf.String()), "KEY=secret_value")
	}
}
