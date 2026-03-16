package chunk

import "testing"

func TestText(t *testing.T) {
	input := "abcdefghijklmnopqrstuvwxyz"
	chunks := Text(input, 10, 2)

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	if chunks[0] != "abcdefghij" {
		t.Fatalf("unexpected first chunk: %q", chunks[0])
	}
	if chunks[1] != "ijklmnopqr" {
		t.Fatalf("unexpected second chunk: %q", chunks[1])
	}
	if chunks[2] != "qrstuvwxyz" {
		t.Fatalf("unexpected third chunk: %q", chunks[2])
	}
}
