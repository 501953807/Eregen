// Package crypto provides encryption utilities for Eregen device communications.
// All algorithms use MIT/BSD-compatible implementations.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrInvalidKey    = errors.New("crypto: invalid key length")
	ErrDecryptionFailed = errors.New("crypto: decryption failed")
)

// GenerateKeyPair generates an Ed25519 key pair for device authentication.
func GenerateKeyPair() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pub.(ed25519.PublicKey), priv.(ed25519.PrivateKey), nil
}

// Sign signs a message with the device private key.
func Sign(privKey ed25519.PrivateKey, msg []byte) []byte {
	sig := ed25519.Sign(privKey, msg)
	return sig
}

// Verify verifies a signature using the device public key.
func Verify(pubKey ed25519.PublicKey, msg, sig []byte) bool {
	return ed25519.Verify(pubKey, msg, sig)
}

// HashPassword hashes a password using SHA-256 with a salt.
// Note: For production password hashing, use bcrypt or argon2.
// This is only used for device token generation.
func HashPassword(password string, salt []byte) string {
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// GenerateSalt returns a 16-byte random salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// EncryptAES encrypts plaintext using AES-256-GCM.
func EncryptAES(plaintext []byte, key [32]byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = aesGCM.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// DecryptAES decrypts ciphertext using AES-256-GCM.
func DecryptAES(ciphertext, nonce []byte, key [32]byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	return plaintext, nil
}

// GenerateDeviceToken generates a unique device provisioning token.
func GenerateDeviceToken() (string, error) {
	b := make([]byte, 24)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return "dt_" + base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateProvisioningCode generates a 6-digit pairing code for WiFi setup.
func GenerateProvisioningCode() (string, error) {
	b := make([]byte, 4)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	code := uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
	return base36Encode(code % 0x1000000), nil
}

func base36Encode(n uint32) string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if n == 0 {
		return string(chars[0])
	}
	var result []byte
	for n > 0 {
		result = append(result, chars[n%36])
		n /= 36
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}
