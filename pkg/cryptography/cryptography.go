// Package cryptography provides cryptographic utilities for the application
// This package implements secure hashing and encoding functions
package cryptography

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Base64HMAC generates a base64-encoded HMAC-SHA256 of the given data using the provided key
// This function:
// 1. Creates an HMAC using SHA-256 as the hash function
// 2. Writes the data to the HMAC
// 3. Computes the HMAC sum
// 4. Encodes the result in base64
// Parameters:
// - data: The data to be hashed
// - key: The secret key used for the HMAC
// Returns:
// - string: The base64-encoded HMAC-SHA256 hash
func Base64HMAC(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
