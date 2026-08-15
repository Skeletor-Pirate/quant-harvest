package main

import "crypto/sha256"

// EnvelopeIdentifier is a non-cryptographic integration placeholder.
// A vetted ML-KEM implementation and key-management design are required before handling secrets.
func EnvelopeIdentifier(plaintext []byte) [32]byte {
	return sha256.Sum256(append([]byte("qharvest-envelope:"), plaintext...))
}
