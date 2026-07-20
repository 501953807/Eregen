package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// PayloadCrypto handles AES-128-CTR + HMAC-SHA256 for device message encryption.
// Matches the firmware implementation in firmware/common/src/payload_crypto.c.
type PayloadCrypto struct {
	aesKey  [16]byte
	hmacKey [16]byte
}

// NewPayloadCrypto derives per-device keys from a factory seed.
func NewPayloadCrypto(masterKey []byte) (*PayloadCrypto, error) {
	if len(masterKey) < 32 {
		return nil, fmt.Errorf("master key must be >= 32 bytes")
	}

	p := &PayloadCrypto{}

	/* Derive AES key: SHA-256(master_key || "eregen-device-v1") */
	deriver := sha256.New()
	deriver.Write(masterKey[:16])
	deriver.Write([]byte("eregen-device-v1"))
	copy(p.aesKey[:], deriver.Sum(nil))

	/* Derive HMAC key: SHA-256(aes_key || "eregen-hmac-v1") */
	hderiv := sha256.New()
	hderiv.Write(p.aesKey[:])
	hderiv.Write([]byte("eregen-hmac-v1"))
	copy(p.hmacKey[:], hderiv.Sum(nil))

	return p, nil
}

// Encrypt encrypts plaintext using AES-128-CTR with HMAC-SHA256 integrity.
// Output format: [nonce:12][len:16][ciphertext:padded][hmac:32]
func (p *PayloadCrypto) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(p.aesKey[:])
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, make([]byte, 16)) // first 12 bytes = nonce

	encPadded := (len(plaintext) + 15) & ^15
	out := make([]byte, 12+16+encPadded+32)

	// Random nonce (bytes 0-11)
	if _, err := rand.Read(out[:12]); err != nil {
		return nil, err
	}

	// Encrypted data (starts at byte 28)
	if len(plaintext) > 0 {
		stream.XORKeyStream(out[28:], plaintext)
	}

	// PKCS#7 padding
	padCount := encPadded - len(plaintext)
	for i := 0; i < padCount; i++ {
		out[28+len(plaintext)+i] = byte(padCount)
	}

	// Store original length as big-endian uint16
	binary.BigEndian.PutUint16(out[12:], uint16(len(plaintext)))

	// HMAC over nonce + length + ciphertext
	mac := hmac.New(sha256.New, p.hmacKey[:])
	mac.Write(out[:12+16+encPadded])
	mac.Sum(out[12+16+encPadded : 12+16+encPadded])

	return out, nil
}

// Decrypt verifies HMAC and returns plaintext. Returns -2 on HMAC mismatch.
func (p *PayloadCrypto) Decrypt(encrypted []byte) ([]byte, error) {
	if len(encrypted) < 60 {
		return nil, fmt.Errorf("payload too short")
	}

	encPadded := len(encrypted) - 32 // remove HMAC
	origLen := int(binary.BigEndian.Uint16(encrypted[12:]))
	if origLen > encPadded-28 {
		return nil, fmt.Errorf("invalid payload length")
	}

	// Verify HMAC
	mac := hmac.New(sha256.New, p.hmacKey[:])
	mac.Write(encrypted[:encPadded])
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(expectedMAC, encrypted[encPadded:]) {
		return nil, fmt.Errorf("HMAC mismatch: payload may be tampered")
	}

	block, err := aes.NewCipher(p.aesKey[:])
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, make([]byte, 16))
	plaintext := make([]byte, origLen)
	if origLen > 0 {
		stream.XORKeyStream(plaintext, encrypted[28:28+origLen])
	}

	return plaintext, nil
}
