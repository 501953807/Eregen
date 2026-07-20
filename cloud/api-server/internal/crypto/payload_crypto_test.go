package crypto

import (
	"testing"
)

func TestNewPayloadCrypto_ValidKey(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	_, err := NewPayloadCrypto(key)
	if err != nil {
		t.Fatalf("NewPayloadCrypto failed: %v", err)
	}
}

func TestNewPayloadCrypto_ShortKey(t *testing.T) {
	_, err := NewPayloadCrypto([]byte("short"))
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	crypto, err := NewPayloadCrypto(key)
	if err != nil {
		t.Fatalf("NewPayloadCrypto failed: %v", err)
	}

	plaintext := []byte(`{"type":"heartbeat","dev_id":"BR-1234","bat":85}`)
	encrypted, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if len(encrypted) < 60 {
		t.Errorf("encrypted too short: %d bytes", len(encrypted))
	}

	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecrypt_EmptyPayload(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	crypto, err := NewPayloadCrypto(key)
	if err != nil {
		t.Fatalf("NewPayloadCrypto failed: %v", err)
	}

	plaintext := []byte("")
	encrypted, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if len(decrypted) != 0 {
		t.Errorf("empty payload round-trip failed: got %d bytes", len(decrypted))
	}
}

func TestEncrypt_DifferentCiphertext(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	crypto, err := NewPayloadCrypto(key)
	if err != nil {
		t.Fatalf("NewPayloadCrypto failed: %v", err)
	}

	plaintext := []byte("same message")
	enc1, _ := crypto.Encrypt(plaintext)
	enc2, _ := crypto.Encrypt(plaintext)
	if string(enc1) == string(enc2) {
		t.Error("encrypt should produce different ciphertext each time (random nonce)")
	}
}

func TestDecrypt_TamperedPayload(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	crypto, err := NewPayloadCrypto(key)
	if err != nil {
		t.Fatalf("NewPayloadCrypto failed: %v", err)
	}

	plaintext := []byte("sensitive data")
	encrypted, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Tamper with the ciphertext
	encrypted[30] ^= 0xFF
	_, err = crypto.Decrypt(encrypted)
	if err == nil {
		t.Error("expected HMAC mismatch error on tampered payload")
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	crypto, err := NewPayloadCrypto(key)
	if err != nil {
		t.Fatalf("NewPayloadCrypto failed: %v", err)
	}

	_, err = crypto.Decrypt([]byte("too short"))
	if err == nil {
		t.Error("expected error for too-short payload")
	}
}
