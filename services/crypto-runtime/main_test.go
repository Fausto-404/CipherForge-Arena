package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func testHMAC(seed, message string) string {
	mac := hmac.New(sha256.New, []byte(seed))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyK03(t *testing.T) {
	seed := "test-seed"
	payload := map[string]interface{}{
		"account":  "alice",
		"operation": "approve-transfer",
		"state":    []interface{}{"arena", "auth", "risk", "business"},
		"businessProof": testHMAC(seed, "arena>auth>risk>alice|approve-transfer"),
	}

	ok, detail := verify(verifyRequest{Challenge: "K03", Seed: seed, Payload: payload})
	if !ok || detail != "multi-domain state machine" {
		t.Fatalf("expected K03 success, got ok=%v detail=%q", ok, detail)
	}
}

func TestVerifyRejectsWrongP02Token(t *testing.T) {
	ok, detail := verify(verifyRequest{
		Challenge: "P02",
		Seed:      "test-seed",
		Payload: map[string]interface{}{
			"symbol": "BTC-USDT",
			"tier":   "pro",
			"token":  "tampered",
		},
	})
	if ok || detail != "wasm wrapper token" {
		t.Fatalf("expected P02 rejection, got ok=%v detail=%q", ok, detail)
	}
}
