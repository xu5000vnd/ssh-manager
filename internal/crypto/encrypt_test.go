package crypto

import "testing"

func TestEncryptDecryptRoundtrip(t *testing.T) {
	password := []byte("secret-password")
	plaintext := []byte("hello ssh manager")

	ciphertext, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	result, err := Decrypt(ciphertext, password)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(result) != string(plaintext) {
		t.Fatalf("got %q, want %q", string(result), string(plaintext))
	}
}

func TestDecryptWrongPasswordFails(t *testing.T) {
	ciphertext, err := Encrypt([]byte("secret"), []byte("pw1"))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if _, err := Decrypt(ciphertext, []byte("pw2")); err == nil {
		t.Fatal("expected decrypt error for wrong password")
	}
}

func TestDecryptCorruptedDataFails(t *testing.T) {
	ciphertext, err := Encrypt([]byte("secret"), []byte("pw"))
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	ciphertext[len(ciphertext)-1] ^= 0xFF

	if _, err := Decrypt(ciphertext, []byte("pw")); err == nil {
		t.Fatal("expected decrypt error for corrupted ciphertext")
	}
}

func TestDecryptEmptyInputFails(t *testing.T) {
	if _, err := Decrypt(nil, []byte("pw")); err == nil {
		t.Fatal("expected decrypt error for empty input")
	}
}
