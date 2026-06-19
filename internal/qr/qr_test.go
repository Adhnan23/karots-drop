package qr

import (
	"strings"
	"testing"
)

const testURL = "https://example.com/test"

func TestTerminal(t *testing.T) {
	s, err := Terminal(testURL)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) == 0 {
		t.Fatal("expected non-empty QR string")
	}
	if !strings.Contains(s, "█") {
		t.Fatal("expected block characters in terminal QR")
	}
}

func TestTerminalDifferentContent(t *testing.T) {
	a, _ := Terminal("content-a")
	b, _ := Terminal("content-b")
	if a == b {
		t.Fatal("different content should produce different QR codes")
	}
}

func TestPNG(t *testing.T) {
	png, err := PNG(testURL)
	if err != nil {
		t.Fatal(err)
	}
	if len(png) == 0 {
		t.Fatal("expected non-empty PNG data")
	}

	// PNG magic bytes
	expected := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range expected {
		if png[i] != b {
			t.Fatalf("invalid PNG header at byte %d: got 0x%02x, want 0x%02x", i, png[i], b)
		}
	}
}

func TestPNGDimensions(t *testing.T) {
	png, _ := PNG(testURL)
	// 256x256 1-bit PNG, should be at least the header (8 bytes)
	if len(png) < 8 {
		t.Fatalf("PNG too small: %d bytes", len(png))
	}
}
