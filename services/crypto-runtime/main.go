package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
)

type verifyRequest struct {
	Challenge string                 `json:"challenge"`
	Seed      string                 `json:"seed"`
	Nonce     string                 `json:"nonce"`
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
		return sameStrings(payload["stages"], []string{"bootstrap", "handshake", "login", "query"}) && hmacHex(request.Seed, body) == stringValue(payload, "proof"), "four-stage login chain"
	case "K02":
		body := stringValue(payload, "account") + "|" + stringValue(payload, "order") + "|" + stringValue(payload, "amount")
		expected := hmacHex(request.Seed, "device|risk|confirm|"+body)
		return numberValue(payload, "riskScore") >= 70 && expected == stringValue(payload, "ratchet"), "risk sdk ratchet"
	case "K03":
		body := stringValue(payload, "account") + "|" + stringValue(payload, "operation")
		expected := hmacHex(request.Seed, "arena>auth>risk>"+body)
		return sameStrings(payload["state"], []string{"arena", "auth", "risk", "business"}) && expected == stringValue(payload, "businessProof"), "multi-domain state machine"
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
