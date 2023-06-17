package client_utils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"strings"
)

/*
	Generates a 16-bit cryptographic nonce to be used as the value of the "Sec-WebSocket-Key" request header during the
	initial handshake as described in section 4.1 of RFC 6455
*/
func GenerateWebSocketKey() string {
	keyBytes := make([]byte, 16)
	rand.Read(keyBytes)
	return base64.StdEncoding.EncodeToString(keyBytes)
}

func ExpectedWebSocketAccept(webSocketKey string) string {
	const GUID string = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	trimmedKey := strings.TrimSpace(webSocketKey)
	appendedBytes := []byte(trimmedKey + GUID)
	hash := sha1.Sum(appendedBytes)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func MaskMessage(messageBytes []byte) (maskedMessage []byte, mask []byte) {
	mask = make([]byte, 4)
	rand.Read(mask)

	maskedMessage = make([]byte, len(messageBytes))

	for i := 0; i < len(messageBytes); i++ {
		maskedMessage[i] = messageBytes[i] ^ mask[i%4]
	}

	return
}
