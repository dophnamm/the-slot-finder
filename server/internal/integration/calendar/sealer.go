package calendar

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// keyBytes is the AES-256 key length CALENDAR_FEED_KEY must decode to.
const keyBytes = 32

// ErrBadKey means CALENDAR_FEED_KEY is not 32 bytes of base64.
var ErrBadKey = errors.New("CALENDAR_FEED_KEY must be 32 bytes, base64-encoded")

// errSealed means a stored link could not be opened: a wrong key or damaged data.
var errSealed = errors.New("calendar link could not be decrypted")

// Sealer encrypts calendar links at rest with AES-GCM. A sealed link is nonce ‖ ciphertext.
type Sealer struct {
	aead cipher.AEAD
}

// NewSealer reads a base64 key of 32 bytes.
func NewSealer(keyBase64 string) (*Sealer, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil || len(key) != keyBytes {
		return nil, ErrBadKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrBadKey
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrBadKey
	}
	return &Sealer{aead: aead}, nil
}

// Seal encrypts a link with a fresh random nonce.
func (s *Sealer) Seal(link string) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, []byte(link), nil), nil
}

// Open decrypts a sealed link.
func (s *Sealer) Open(sealed []byte) (string, error) {
	n := s.aead.NonceSize()
	if len(sealed) < n {
		return "", errSealed
	}
	plain, err := s.aead.Open(nil, sealed[:n], sealed[n:], nil)
	if err != nil {
		return "", errSealed
	}
	return string(plain), nil
}
