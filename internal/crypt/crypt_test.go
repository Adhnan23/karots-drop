package crypt

import (
	"bytes"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key))
	}

	// Ensure keys are random
	key2, _ := GenerateKey()
	if bytes.Equal(key, key2) {
		t.Fatal("expected different keys")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("hello world")
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext should not equal plaintext")
	}

	result, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, plaintext) {
		t.Fatalf("expected %q, got %q", plaintext, result)
	}
}

func TestEncryptDecryptEmpty(t *testing.T) {
	key, _ := GenerateKey()
	ciphertext, err := Encrypt([]byte{}, key)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Fatal("expected empty result")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	plaintext := []byte("secret data")
	ciphertext, _ := Encrypt(plaintext, key1)

	_, err := Decrypt(ciphertext, key2)
	if err == nil {
		t.Fatal("expected error with wrong key")
	}
}

func TestDecryptTamperedData(t *testing.T) {
	key, _ := GenerateKey()

	plaintext := []byte("test")
	ciphertext, _ := Encrypt(plaintext, key)

	// Flip a byte in the ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err := Decrypt(ciphertext, key)
	if err == nil {
		t.Fatal("expected error with tampered data")
	}
}

func TestEncryptLargeData(t *testing.T) {
	key, _ := GenerateKey()
	plaintext := make([]byte, 1<<20) // 1MB
	for i := range plaintext {
		plaintext[i] = byte(i)
	}

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}

	result, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(result, plaintext) {
		t.Fatal("large data roundtrip failed")
	}
}
