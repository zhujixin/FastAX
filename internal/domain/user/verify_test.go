package user

import (
	"strings"
	"testing"
)

func TestVerifyService_GenerateCode_NilCache(t *testing.T) {
	vs := NewVerifyService(nil)

	code, err := vs.GenerateCode("test@test.com")
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}
	if len(code) != 6 {
		t.Errorf("code length = %v, want 6", len(code))
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			t.Errorf("code contains non-digit: %c", ch)
		}
	}
}

func TestVerifyService_VerifyCode_NilCache_AlwaysPasses(t *testing.T) {
	vs := NewVerifyService(nil)

	ok, err := vs.VerifyCode("test@test.com", "123456")
	if err != nil {
		t.Fatalf("VerifyCode() error = %v", err)
	}
	if !ok {
		t.Error("VerifyCode() should return true with nil cache (dev mode)")
	}
}

func TestVerifyService_VerifyCode_NilCache_WrongCode(t *testing.T) {
	vs := NewVerifyService(nil)

	ok, err := vs.VerifyCode("test@test.com", "wrong")
	if err != nil {
		t.Fatalf("VerifyCode() error = %v", err)
	}
	if !ok {
		t.Error("VerifyCode() should return true with nil cache regardless of code")
	}
}

func TestGenerateRandomCode_Length(t *testing.T) {
	for _, length := range []int{4, 6, 8} {
		code, err := generateRandomCode(length)
		if err != nil {
			t.Fatalf("generateRandomCode(%d) error = %v", length, err)
		}
		if len(code) != length {
			t.Errorf("generateRandomCode(%d) length = %d", length, len(code))
		}
	}
}

func TestGenerateRandomCode_OnlyDigits(t *testing.T) {
	code, err := generateRandomCode(100)
	if err != nil {
		t.Fatalf("generateRandomCode() error = %v", err)
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			t.Errorf("non-digit character: %c", ch)
		}
	}
}

func TestGenerateRandomCode_Unique(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, _ := generateRandomCode(6)
		codes[code] = true
	}
	// With 6 digits, 100 codes should almost certainly be unique
	if len(codes) < 90 {
		t.Errorf("too many collisions: got %d unique out of 100", len(codes))
	}
}

func TestGenerateRandomCode_NonEmpty(t *testing.T) {
	code, err := generateRandomCode(6)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(code) == "" {
		t.Error("code should not be empty")
	}
}
