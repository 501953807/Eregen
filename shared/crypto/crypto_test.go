package crypto

import (
	"crypto/ed25519"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pub) != ed25519.PublicKeySize {
		t.Errorf("expected public key size %d, got %d", ed25519.PublicKeySize, len(pub))
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Errorf("expected private key size %d, got %d", ed25519.PrivateKeySize, len(priv))
	}
}

func TestSignAndVerify(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("key generation failed: %v", err)
	}

	msg := []byte("hello eregen")
	sig := Sign(priv, msg)

	if !Verify(pub, msg, sig) {
		t.Error("signature verification failed for valid signature")
	}

	// Tampered message should fail
	tampered := []byte("hello eregeN")
	if Verify(pub, tampered, sig) {
		t.Error("verification passed for tampered message")
	}

	// Wrong key should fail
	pub2, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("second key generation failed: %v", err)
	}
	if Verify(pub2, msg, sig) {
		t.Error("verification passed with wrong key")
	}
}

func TestHashPassword(t *testing.T) {
	salt := []byte("test-salt-16bytes!")
	hash := HashPassword("mypassword", salt)
	if hash == "" {
		t.Fatal("empty password hash")
	}

	// Same input → same output
	hash2 := HashPassword("mypassword", salt)
	if hash != hash2 {
		t.Error("same input produced different hashes")
	}

	// Different password → different output
	hash3 := HashPassword("different", salt)
	if hash == hash3 {
		t.Error("different passwords produced same hash")
	}

	// Different salt → different output
	salt2 := []byte("different-salt!!")
	hash4 := HashPassword("mypassword", salt2)
	if hash == hash4 {
		t.Error("different salts produced same hash")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(salt1) != 16 {
		t.Errorf("expected salt length 16, got %d", len(salt1))
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Two salts should be different
	if string(salt1) == string(salt2) {
		t.Error("two generated salts are identical")
	}
}

func TestEncryptDecryptAES(t *testing.T) {
	var key [32]byte
	copy(key[:], "01234567890123456789012345678901")

	plaintext := []byte("sensitive health data")
	ciphertext, nonce, err := EncryptAES(plaintext, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatal("empty ciphertext")
	}
	if len(nonce) == 0 {
		t.Fatal("empty nonce")
	}

	decrypted, err := DecryptAES(ciphertext, nonce, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecryptAESWrongKey(t *testing.T) {
	var key1 [32]byte
	var key2 [32]byte
	copy(key1[:], "01234567890123456789012345678901")
	copy(key2[:], "12345678901234567890123456789012")

	plaintext := []byte("secret")
	ciphertext, nonce, err := EncryptAES(plaintext, key1)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	_, err = DecryptAES(ciphertext, nonce, key2)
	if err == nil {
		t.Fatal("expected decryption failure with wrong key")
	}
}

func TestEncryptAESNonceUniqueness(t *testing.T) {
	var key [32]byte
	copy(key[:], "01234567890123456789012345678901")

	plaintext := []byte("same message twice")
	_, nonce1, _ := EncryptAES(plaintext, key)
	_, nonce2, _ := EncryptAES(plaintext, key)

	if string(nonce1) == string(nonce2) {
		t.Error("two encryptions of same plaintext produced same nonce")
	}
}

func TestGenerateDeviceToken(t *testing.T) {
	token1, err := GenerateDeviceToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token1) < 5 || token1[:3] != "dt_" {
		t.Errorf("token has invalid prefix: %q", token1)
	}

	token2, err := GenerateDeviceToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1 == token2 {
		t.Error("two generated tokens are identical")
	}
}

func TestGenerateProvisioningCode(t *testing.T) {
	code1, err := GenerateProvisioningCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code1) < 4 || len(code1) > 6 {
		t.Errorf("expected short code (4-6 chars), got length %d: %q", len(code1), code1)
	}

	code2, err := GenerateProvisioningCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code1 == code2 {
		t.Error("two generated codes are identical")
	}

	// All chars should be base36
	for _, c := range code1 {
		valid := (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')
		if !valid {
			t.Errorf("invalid char in provisioning code: %q", string(c))
		}
	}
}

func TestBase36Encode(t *testing.T) {
	tests := []struct {
		input    uint32
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{35, "Z"},
		{36, "10"},
		{1000000, "LFLS"},
		{16777215, "9ZLDR"}, // max 4-digit base36
	}

	for _, tt := range tests {
		result := base36Encode(tt.input)
		if result != tt.expected {
			t.Errorf("base36Encode(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
