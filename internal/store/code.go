package store

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateCode() (string, error) {
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = byte('0') + byte(n.Int64())
	}
	return string(code), nil
}

func (s *Store) GenerateUniqueCode() (string, error) {
	for range 100 {
		code, err := GenerateCode()
		if err != nil {
			return "", err
		}
		if !s.Exists(code) {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique code after 100 attempts")
}
