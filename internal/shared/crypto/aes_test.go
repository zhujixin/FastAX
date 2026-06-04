package crypto

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	plaintext := "sk-this-is-a-secret-api-key-12345"
	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if encrypted == plaintext || encrypted == "" {
		t.Error("Encrypt() returned unchanged or empty string")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	encrypted, _ := Encrypt("hello", key1)
	_, err := Decrypt(encrypted, key2)
	if err == nil {
		t.Error("Decrypt() with wrong key should error")
	}
}

func TestEncryptInvalidKeySize(t *testing.T) {
	_, err := Encrypt("hello", []byte("short"))
	if err == nil {
		t.Error("Encrypt() with short key should error")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	key, _ := GenerateKey()
	_, err := Decrypt("not-base64!!!", key)
	if err == nil {
		t.Error("Decrypt() with invalid base64 should error")
	}
}
