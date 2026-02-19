package crypto

import "golang.org/x/crypto/scrypt"

const (
	scryptN      = 32768
	scryptR      = 8
	scryptP      = 1
	scryptKeyLen = 32
)

func DeriveKey(password []byte, salt []byte) ([]byte, error) {
	return scrypt.Key(password, salt, scryptN, scryptR, scryptP, scryptKeyLen)
}
