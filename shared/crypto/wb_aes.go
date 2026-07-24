// Package crypto provides BLE encryption utilities for medical wristband communication.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrInvalidPairingCode = errors.New("crypto: invalid pairing code")
	ErrDecryptionFailed   = errors.New("crypto: decryption failed")
	ErrHMACMismatch       = errors.New("crypto: HMAC verification failed")
)

const (
	// PairingCodeLength is the number of digits in the pairing code
	PairingCodeLength = 4
	// SessionTimeoutSeconds is the maximum duration a BLE session is valid
	SessionTimeoutSeconds = 300
	// AESKeySize is the size of the AES key in bytes (AES-128)
	AESKeySize = 16
	// IVSize is the size of the initialization vector
	IVSize = 16
	// HMACKeySize is the size of the HMAC key
	HMACKeySize = 32
)

// GeneratePairingCode creates a random 4-digit pairing code
func GeneratePairingCode() (string, error) {
	b := make([]byte, 2)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	code := binary.BigEndian.Uint16(b) % 10000
	return fmt.Sprintf("%04d", code), nil
}

// ValidatePairingCode checks if the pairing code is valid (4 digits)
func ValidatePairingCode(code string) bool {
	if len(code) != PairingCodeLength {
		return false
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// DeriveSessionKey derives an AES-128 key from a pairing code and device ID
func DeriveSessionKey(pairingCode, deviceID string) ([]byte, error) {
	if !ValidatePairingCode(pairingCode) {
		return nil, ErrInvalidPairingCode
	}

	// Combine pairing code and device ID to create a deterministic key
	keyMaterial := []byte(pairingCode + ":" + deviceID)
	h := sha256.New()
	h.Write(keyMaterial)
	key := h.Sum(nil)[:AESKeySize]
	return key, nil
}

// EncryptPatientInfo encrypts patient information using AES-128-CBC
func EncryptPatientInfo(plaintext []byte, key []byte) (ciphertext, iv []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	iv = make([]byte, IVSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	padded := pkcs7Pad(plaintext, block.BlockSize())
	ciphertext = make([]byte, len(padded))
	mode.CryptBlocks(ciphertext, padded)

	return ciphertext, iv, nil
}

// DecryptPatientInfo decrypts patient information using AES-128-CBC
func DecryptPatientInfo(ciphertext, iv, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, ErrDecryptionFailed
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateHMAC generates an HMAC-SHA256 signature for challenge-response authentication
func GenerateHMAC(key []byte, message []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return mac.Sum(nil)
}

// VerifyHMAC verifies an HMAC-SHA256 signature
func VerifyHMAC(key []byte, message, signature []byte) bool {
	expected := GenerateHMAC(key, message)
	return hmac.Equal(expected, signature)
}

// CreateChallengeResponse creates a challenge-response pair for authentication
func CreateChallengeResponse(secretKey []byte) (challenge, response []byte, err error) {
	challenge = make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, challenge); err != nil {
		return nil, nil, err
	}

	response = GenerateHMAC(secretKey, challenge)
	return challenge, response, nil
}

// VerifyChallengeResponse verifies a challenge-response authentication
func VerifyChallengeResponse(secretKey, challenge, response []byte) bool {
	expected := GenerateHMAC(secretKey, challenge)
	return hmac.Equal(expected, response)
}

// SerializeBLEMessage serializes a BLE message into bytes for transmission
func SerializeBLEMessage(msgType BLEMessageType, payload []byte) ([]byte, error) {
	header := make([]byte, 2)
	header[0] = byte(msgType >> 8)
	header[1] = byte(msgType)

	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(payload)))

	result := append(header, length...)
	result = append(result, payload...)
	return result, nil
}

// DeserializeBLEMessage deserializes a BLE message from bytes
func DeserializeBLEMessage(data []byte) (BLEMessageType, []byte, error) {
	if len(data) < 6 {
		return 0, nil, errors.New("crypto: message too short")
	}

	msgType := BLEMessageType(binary.BigEndian.Uint16(data[:2]))
	length := binary.BigEndian.Uint32(data[2:6])

	if uint32(len(data)-6) != length {
		return 0, nil, errors.New("crypto: payload length mismatch")
	}

	payload := data[6:]
	return msgType, payload, nil
}

// EncodeBase64 encodes bytes to base64 string
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decodes a base64 string to bytes
func DecodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// pkcs7Pad adds PKCS7 padding to plaintext
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpad removes PKCS7 padding from plaintext
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("crypto: empty data")
	}

	padding := int(data[len(data)-1])
	if padding <= 0 || padding > len(data) {
		return nil, errors.New("crypto: invalid padding")
	}

	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("crypto: invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// FormatPairingCode formats a pairing code for display
func FormatPairingCode(code string) string {
	if len(code) == PairingCodeLength {
		return strings.Join(strings.Split(code, ""), " ")
	}
	return code
}
