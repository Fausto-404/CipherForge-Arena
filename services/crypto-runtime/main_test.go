package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"
)

func testHMAC(seed, message string) string {
	mac := hmac.New(sha256.New, []byte(seed))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func testRotate(value byte, shift uint) byte {
	return (value << shift) | (value >> (8 - shift))
}

func testSeal(plainText, seed, nonce string) string {
	key := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|cfx/1", seed, nonce)))
	plain := []byte(plainText)
	framed := append([]byte{'C', 'F', 'X', '1', byte(len(plain) >> 8), byte(len(plain))}, plain...)
	padding := 16 - len(framed)%16
	for i := 0; i < padding; i++ {
		framed = append(framed, byte(padding))
	}
	previous := make([]byte, 16)
	box := make([]byte, 256)
	for i := range box {
		box[i] = key[(i*13)%len(key)] ^ byte((i*29+17)&255)
	}
	ciphertext := make([]byte, 0, len(framed))
	for offset := 0; offset < len(framed); offset += 16 {
		block := make([]byte, 16)
		copy(block, framed[offset:offset+16])
		for i := range block {
			block[i] ^= previous[i]
		}
		left := append([]byte(nil), block[:8]...)
		right := append([]byte(nil), block[8:]...)
		for round := 0; round < 16; round++ {
			nextRight := make([]byte, 8)
			for i := 0; i < 8; i++ {
				a := key[(round*7+i)%len(key)]
				b := key[(round*11+i*3)%len(key)]
				tweak := byte(((round + 1) * (i + 3)) & 255)
				value := byte((int(right[i]) + int(a) + int(tweak)) & 255)
				nextRight[i] = left[i] ^ (testRotate(box[value], uint((round+i)%7+1)) ^ b)
			}
			left, right = right, nextRight
		}
		sealed := append(append([]byte(nil), left...), right...)
		ciphertext = append(ciphertext, sealed...)
		previous = sealed
	}
	body := append([]byte{'C', 'F', 'X', '1'}, mustHexDecode(nonce)...)
	body = append(body, ciphertext...)
	digest := sha256.Sum256(append(key[:], body...))
	body = append(body, digest[:12]...)
	return base64.RawURLEncoding.EncodeToString(body)
}

func mustHexDecode(value string) []byte {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return decoded
}

func TestVerifyK03(t *testing.T) {
	seed := "test-seed"
	nonce := "00112233445566778899aabb"
	guard := map[string]interface{}{
		"sourceHash":    "asset-hash",
		"score":         float64(100),
		"debuggerDelta": float64(0),
		"viewportGap":   float64(0),
		"visibility":    "visible",
		"consoleGuard":  true,
	}
	payload := map[string]interface{}{
		"account":       "alice",
		"operation":     "approve-transfer",
		"state":         []interface{}{"arena", "auth", "risk", "business"},
		"businessProof": testHMAC(seed, "arena>auth>risk>alice|approve-transfer"),
		"envelope":      testSeal("alice|approve-transfer", seed, nonce),
		"guard":         guard,
		"guardProof":    testHMAC(seed, guardMessage(guard)),
	}

	ok, detail := verify(verifyRequest{Challenge: "K03", Seed: seed, Nonce: nonce, AssetHash: "asset-hash", Payload: payload})
	if !ok || detail != "multi-domain encrypted state machine" {
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

func TestCfxUnsealMatchesBrowserFixture(t *testing.T) {
	value, ok := cfxUnseal("Q0ZYMQARIjNEVWZ3iJmquxOtG9w9SYuv6HtoAM-thuA1_WKifW7diTD9KwOaEIJCFjscw0GkxUwlI5Oo", "test-seed", "00112233445566778899aabb")
	if !ok || value != "alice|approve-transfer" {
		t.Fatalf("expected browser envelope to open, got ok=%v value=%q", ok, value)
	}
}

func TestCfxUnsealRejectsTamperedEnvelope(t *testing.T) {
	envelope := []byte(testSeal("alice|approve-transfer", "test-seed", "00112233445566778899aabb"))
	for index := range envelope {
		if envelope[index] == '-' || envelope[index] == '_' || envelope[index] == 'A' {
			continue
		}
		envelope[index] = 'A'
		break
	}
	if value, ok := cfxUnseal(string(envelope), "test-seed", "00112233445566778899aabb"); ok || value != "" {
		t.Fatalf("expected tampered envelope rejection, got ok=%v value=%q", ok, value)
	}
}
