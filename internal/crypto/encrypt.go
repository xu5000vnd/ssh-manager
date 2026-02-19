package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

const (
	saltSize  = 32
	nonceSize = 12
)

func Encrypt(plaintext []byte, password []byte) ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	key, err := DeriveKey(password, salt)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}
	defer zeroBytes(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	output := make([]byte, 0, saltSize+nonceSize+len(ciphertext))
	output = append(output, salt...)
	output = append(output, nonce...)
	output = append(output, ciphertext...)

	return output, nil
}

func Decrypt(data []byte, password []byte) ([]byte, error) {
	if len(data) < saltSize+nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	key, err := DeriveKey(password, salt)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}
	defer zeroBytes(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
