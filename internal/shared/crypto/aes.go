// AES-256-GCM 加密工具
//
// 本文件实现 AES-256-GCM 加解密，用于敏感数据保护：
//   - BYOK Key 加密存储
//   - 身份证号加密
//   - API Key 加密
//
// 每次加密使用随机 nonce，密文格式: nonce(12字节) + ciphertext + tag(16字节)
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Encrypt 使用 AES-256-GCM 加密明文
// key 必须为 32 字节（256 位）
// 返回 base64 编码的密文（含 nonce）
func Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("key must be 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 使用 AES-256-GCM 解密密文
// key 必须为 32 字节（256 位）
// ciphertext 为 base64 编码的密文（含 nonce）
func Decrypt(encoded string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("key must be 32 bytes, got %d", len(key))
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// GenerateKey 生成 32 字节随机 AES-256 密钥
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return key, nil
}
