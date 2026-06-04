package mask

import "testing"

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"13800138000", "138****8000"},
		{"12345678901", "123****8901"},
		{"12345", "*****"},
	}
	for _, tt := range tests {
		got := MaskPhone(tt.input)
		if got != tt.expected {
			t.Errorf("MaskPhone(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"john.doe@example.com", "j***@example.com"},
		{"a@b.com", "a@b.com"},
		{"ab@c.com", "a***@c.com"},
	}
	for _, tt := range tests {
		got := MaskEmail(tt.input)
		if got != tt.expected {
			t.Errorf("MaskEmail(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMaskIDNumber(t *testing.T) {
	got := MaskIDNumber("110101199001011234")
	expected := "1101**********1234"
	if got != expected {
		t.Errorf("MaskIDNumber() = %q, want %q", got, expected)
	}
}

func TestMaskBankCard(t *testing.T) {
	got := MaskBankCard("6222021234567890123")
	expected := "**** **** **** 0123"
	if got != expected {
		t.Errorf("MaskBankCard() = %q, want %q", got, expected)
	}
}

func TestMaskAPIKey(t *testing.T) {
	got := MaskAPIKey("sk-abcdefghijklmnopqrstuvwxyz1234567890")
	if len(got) < 8 || got[:2] != "sk" {
		t.Errorf("MaskAPIKey() = %q, expected masked API key", got)
	}
}
