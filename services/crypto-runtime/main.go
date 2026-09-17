package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type verifyRequest struct {
	Challenge string                 `json:"challenge"`
	Seed      string                 `json:"seed"`
	Nonce     string                 `json:"nonce"`
	AssetHash string                 `json:"assetHash"`
	Payload   map[string]interface{} `json:"payload"`
}

type verifyResponse struct {
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

func stringValue(payload map[string]interface{}, key string) string {
	value, _ := payload[key].(string)
	return value
}

func numberValue(payload map[string]interface{}, key string) float64 {
	value, _ := payload[key].(float64)
	return value
}

func boolValue(payload map[string]interface{}, key string) bool {
	value, _ := payload[key].(bool)
	return value
}

func rotate(value byte, shift uint) byte {
	return (value << shift) | (value >> (8 - shift))
}

func guardMessage(guard map[string]interface{}) string {
	consoleGuard := "0"
	if boolValue(guard, "consoleGuard") {
		consoleGuard = "1"
	}
	toStringGuard := "0"
	if boolValue(guard, "toStringGuard") {
		toStringGuard = "1"
	}
	stackGuard := "0"
	if boolValue(guard, "stackGuard") {
		stackGuard = "1"
	}
	return strconv.Itoa(int(numberValue(guard, "probeVersion"))) + "|" +
		stringValue(guard, "probeToken") + "|" +
		stringValue(guard, "sourceHash") + "|" +
		strconv.Itoa(int(numberValue(guard, "score"))) + "|" +
		strconv.Itoa(int(numberValue(guard, "debuggerDelta"))) + "|" +
		strconv.Itoa(int(numberValue(guard, "timingVariance"))) + "|" +
		strconv.Itoa(int(numberValue(guard, "viewportGap"))) + "|" +
		stringValue(guard, "visibility") + "|" + consoleGuard + "|" + toStringGuard + "|" + stackGuard
}

func verifyGuard(request verifyRequest) bool {
	guard, ok := request.Payload["guard"].(map[string]interface{})
	if !ok || request.AssetHash == "" {
		return false
	}
	probeDigest := sha256.Sum256([]byte(request.Nonce + "|guard/2"))
	probeToken := hex.EncodeToString(probeDigest[:])[:16]
	return stringValue(guard, "sourceHash") == request.AssetHash &&
		int(numberValue(guard, "probeVersion")) == 2 &&
		stringValue(guard, "probeToken") == probeToken &&
		numberValue(guard, "score") >= 60 &&
		numberValue(guard, "debuggerDelta") <= 120 &&
		numberValue(guard, "timingVariance") <= 60 &&
		numberValue(guard, "viewportGap") <= 280 &&
		stringValue(guard, "visibility") == "visible" &&
		boolValue(guard, "consoleGuard") &&
		boolValue(guard, "toStringGuard") &&
		boolValue(guard, "stackGuard") &&
		hmacHex(request.Seed, guardMessage(guard)) == stringValue(request.Payload, "guardProof")
}

func cfxUnseal(envelope, seed, nonce string) (string, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(envelope)
	if err != nil || len(raw) < 4+12+16+12 || string(raw[:4]) != "CFX2" {
		return "", false
	}
	nonceBytes, err := hex.DecodeString(nonce)
	if err != nil || len(nonceBytes) != 12 || !bytes.Equal(nonceBytes, raw[4:16]) {
		return "", false
	}
	key := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|cfx/2", seed, nonce)))
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
	cipherLength := len(raw) - 4 - 12 - 12
	if cipherLength < 16 || cipherLength%16 != 0 {
		return "", false
	}
	body := raw[:4+12+cipherLength]
	tagDigest := sha256.Sum256(append(key[:], body...))
	if !hmac.Equal(tagDigest[:12], raw[len(raw)-12:]) {
		return "", false
	}

	plain := make([]byte, 0, cipherLength)
	previous := make([]byte, 16)
	for offset := 0; offset < cipherLength; offset += 16 {
		block := raw[16+offset : 16+offset+16]
		left := append([]byte(nil), block[:8]...)
		right := append([]byte(nil), block[8:]...)
		for round := roundCount - 1; round >= 0; round-- {
			oldRight := append([]byte(nil), left...)
			oldLeft := make([]byte, 8)
			slot := route[round]
			for i := 0; i < 8; i++ {
				a := key[(slot+i*3+round)%len(key)]
				b := key[(slot*5+i*7+round*3)%len(key)]
				tweak := byte((slot + (round+1)*(i+3) + int(key[(round+i)%len(key)])) & 255)
				value := byte((int(oldRight[i]) + int(a) + int(tweak)) & 255)
				mixed := rotate(box[value], uint((round+i)%7+1)) ^ b
				oldLeft[i] = right[i] ^ mixed
			}
			left, right = oldLeft, oldRight
		}
		decoded := make([]byte, 16)
		for i := 0; i < 16; i++ {
			if i < 8 {
				decoded[i] = left[i] ^ previous[i]
			} else {
				decoded[i] = right[i-8] ^ previous[i]
			}
		}
		plain = append(plain, decoded...)
		copy(previous, block)
	}

	padding := int(plain[len(plain)-1])
	if padding < 1 || padding > 16 || !bytes.Equal(plain[len(plain)-padding:], bytes.Repeat([]byte{byte(padding)}, padding)) {
		return "", false
	}
	plain = plain[:len(plain)-padding]
	if len(plain) < 6 || string(plain[:4]) != "CFX2" {
		return "", false
	}
	length := int(plain[4])<<8 | int(plain[5])
	value := plain[6:]
	return string(value), len(value) == length
}

func hmacHex(seed, message string) string {
	mac := hmac.New(sha256.New, []byte(seed))
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func sameStrings(value interface{}, expected []string) bool {
	items, ok := value.([]interface{})
	if !ok || len(items) != len(expected) {
		return false
	}
	for index, item := range items {
		text, ok := item.(string)
		if !ok || text != expected[index] {
			return false
		}
	}
	return true
}

func verify(request verifyRequest) (bool, string) {
	payload := request.Payload
	switch request.Challenge {
	case "P01":
		body := stringValue(payload, "account") + "|" + stringValue(payload, "amount")
		return hmacHex(request.Seed, body) == stringValue(payload, "workerProof"), "worker envelope"
	case "P02":
		body := stringValue(payload, "symbol") + "|" + stringValue(payload, "tier")
		mac := hmac.New(sha256.New, []byte(request.Seed))
		_, _ = mac.Write([]byte(body))
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		return expected == stringValue(payload, "token"), "wasm wrapper token"
	case "P03":
		body := stringValue(payload, "account") + "|" + stringValue(payload, "resource")
		return boolValue(payload, "viaWorker") && hmacHex(request.Seed, body) == stringValue(payload, "shadowProof"), "service worker rewritten request"
	case "K01":
		body := "bootstrap|" + stringValue(payload, "account") + "|login|" + stringValue(payload, "password") + "|query"
		return verifyGuard(request) && sameStrings(payload["stages"], []string{"bootstrap", "handshake", "login", "query"}) && hmacHex(request.Seed, body) == stringValue(payload, "proof"), "four-stage login chain"
	case "K02":
		body := stringValue(payload, "account") + "|" + stringValue(payload, "order") + "|" + stringValue(payload, "amount")
		expected := hmacHex(request.Seed, "device|risk|confirm|"+body)
		return verifyGuard(request) && numberValue(payload, "riskScore") >= 70 && expected == stringValue(payload, "ratchet"), "risk sdk ratchet"
	case "K03":
		body := stringValue(payload, "account") + "|" + stringValue(payload, "operation")
		expected := hmacHex(request.Seed, "arena>auth>risk>"+body)
		opened, openedOK := cfxUnseal(stringValue(payload, "envelope"), request.Seed, request.Nonce)
		return verifyGuard(request) && sameStrings(payload["state"], []string{"arena", "auth", "risk", "business"}) && expected == stringValue(payload, "businessProof") && openedOK && opened == body, "multi-domain encrypted state machine"
	default:
		return false, "unsupported runtime challenge"
	}
}

func verifyHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input verifyRequest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	ok, detail := verify(input)
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(writer).Encode(verifyResponse{OK: ok, Detail: detail})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = writer.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/verify", verifyHandler)
	log.Println("CipherForge crypto runtime listening on :9090")
	log.Fatal(http.ListenAndServe(":9090", mux))
}
