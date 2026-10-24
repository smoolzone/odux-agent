package secret

import (
	"crypto/rand"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	enc, err := encrypt(key, "sc_mysecretkey")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := decrypt(key, enc)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "sc_mysecretkey" {
		t.Fatalf("round trip mismatch: %q", plain)
	}
}

func TestWrongKeyFails(t *testing.T) {
	key := make([]byte, 32)
	other := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(other); err != nil {
		t.Fatal(err)
	}
	enc, err := encrypt(key, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decrypt(other, enc); err == nil {
		t.Fatal("expected decryption to fail with the wrong key")
	}
}
