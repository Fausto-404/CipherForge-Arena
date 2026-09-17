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
	key := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|cfx/2", seed, nonce)))
	plain := []byte(plainText)
	framed := append([]byte{'C', 'F', 'X', '2', byte(len(plain) >> 8), byte(len(plain))}, plain...)
	padding := 16 - len(framed)%16
	for i := 0; i < padding; i++ {
		framed = append(framed, byte(padding))
	}
	previous := make([]byte, 16)
	roundCount := 12 + int(key[0]&7)
	route := make([]int, roundCount)
	for round := range route {
		route[round] = int((key[(round*5+3)%len(key)] ^ byte((round*23+41)&255)) & 31)
	}
	box := make([]byte, 256)
	for i := range box {
		pivot := route[i%roundCount]
		box[i] = key[(i*13+pivot)%len(key)] ^ byte((i*29+17+pivot*7)&255)
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
		for round := 0; round < roundCount; round++ {
			nextRight := make([]byte, 8)
			slot := route[round]
			for i := 0; i < 8; i++ {
				a := key[(slot+i*3+round)%len(key)]
				b := key[(slot*5+i*7+round*3)%len(key)]
				tweak := byte((slot + (round+1)*(i+3) + int(key[(round+i)%len(key)])) & 255)
				value := byte((int(right[i]) + int(a) + int(tweak)) & 255)
				nextRight[i] = left[i] ^ (testRotate(box[value], uint((round+i)%7+1)) ^ b)
			}
			left, right = right, nextRight
		}
		sealed := append(append([]byte(nil), left...), right...)
		ciphertext = append(ciphertext, sealed...)
		previous = sealed
	}
	body := append([]byte{'C', 'F', 'X', '2'}, mustHexDecode(nonce)...)
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
		"probeVersion":   float64(2),
		"probeToken":     "5582b8c64de61510",
		"sourceHash":     "asset-hash",
		"score":          float64(100),
		"debuggerDelta":  float64(0),
		"timingVariance": float64(0),
		"viewportGap":    float64(0),
		"visibility":     "visible",
		"consoleGuard":   true,
		"toStringGuard":  true,
		"stackGuard":     true,
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
	value, ok := cfxUnseal("Q0ZYMgARIjNEVWZ3iJmquywM5D07IKM8mCS030PxkiH78O6pOpOsDG_tc8D3jF7xWYJQM0lC38wETWs2", "test-seed", "00112233445566778899aabb")
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
