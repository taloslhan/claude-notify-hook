package ui

import "testing"

func TestMaskToken_Normal(t *testing.T) {
	got := MaskToken("1234567890ABCDEF")
	if got != "1234...CDEF" {
		t.Errorf("MaskToken normal = %q, want '1234...CDEF'", got)
	}
}

func TestMaskToken_Short(t *testing.T) {
	got := MaskToken("short")
	if got != "****" {
		t.Errorf("MaskToken short = %q, want '****'", got)
	}
}

func TestMaskToken_ExactlyTen(t *testing.T) {
	got := MaskToken("1234567890")
	if got != "****" {
		t.Errorf("MaskToken 10 chars = %q, want '****'", got)
	}
}

func TestMaskToken_ElevenChars(t *testing.T) {
	got := MaskToken("12345678901")
	if got != "1234...8901" {
		t.Errorf("MaskToken 11 chars = %q, want '1234...8901'", got)
	}
}

func TestMaskToken_Empty(t *testing.T) {
	got := MaskToken("")
	if got != "****" {
		t.Errorf("MaskToken empty = %q, want '****'", got)
	}
}
