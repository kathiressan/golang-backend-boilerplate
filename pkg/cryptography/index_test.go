package cryptographyHelper

import (
	"testing"
)

func TestBase64HMAC_ValidInput(t *testing.T) {
	data := "message"
	key := "secret-key"

	result := Base64HMAC(data, key)

	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestBase64HMAC_SameInputsProduceSameOutput(t *testing.T) {
	data := "message"
	key := "secret-key"

	result1 := Base64HMAC(data, key)
	result2 := Base64HMAC(data, key)

	if result1 != result2 {
		t.Error("expected same inputs to produce same output")
	}
}

func TestBase64HMAC_DifferentKeysProduceDifferentOutput(t *testing.T) {
	data := "message"

	result1 := Base64HMAC(data, "key1")
	result2 := Base64HMAC(data, "key2")

	if result1 == result2 {
		t.Error("expected different keys to produce different output")
	}
}

func TestBase64HMAC_DifferentDataProduceDifferentOutput(t *testing.T) {
	key := "secret-key"

	result1 := Base64HMAC("data1", key)
	result2 := Base64HMAC("data2", key)

	if result1 == result2 {
		t.Error("expected different data to produce different output")
	}
}

func TestBase64HMAC_EmptyData(t *testing.T) {
	key := "secret-key"

	result := Base64HMAC("", key)

	if result == "" {
		t.Error("expected non-empty result for empty data")
	}
}

func TestBase64HMAC_EmptyKey(t *testing.T) {
	data := "message"

	result := Base64HMAC(data, "")

	if result == "" {
		t.Error("expected non-empty result for empty key")
	}
}
