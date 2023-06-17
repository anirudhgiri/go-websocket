package websocket

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"

	"github.com/anirudhgiri/go-websocket/internal/utils/client_utils"
	"github.com/anirudhgiri/go-websocket/internal/utils/url_utils"
)

type WebSocketConnection struct {
	ServerURL     *url.URL
	TLSConnection *tls.Conn
}

func Dial(serverURL string) (*WebSocketConnection, error) {
	var socket WebSocketConnection

	serverURLStruct, err := url_utils.FormConnectionURL(serverURL)
	if err != nil {
		return nil, err
	}
	socket.ServerURL = serverURLStruct

	serverPort, err := url_utils.GetServerPort(serverURLStruct)
	if err != nil {
		return nil, err
	}
	var tlsDialURL string = serverURLStruct.Host + serverPort
	config := &tls.Config{}

	connection, err := tls.Dial("tcp", tlsDialURL, config)
	if err != nil {
		return nil, err
	}
	socket.TLSConnection = connection

	websocketKey := client_utils.GenerateWebSocketKey()

	request := &http.Request{
		Method: http.MethodGet,
		URL:    serverURLStruct,
		Header: http.Header{
			"Connection":            {"Upgrade"},
			"Upgrade":               {"websocket"},
			"Sec-WebSocket-Key":     {websocketKey},
			"Sec-WebSocket-Version": {"13"},
		},
	}

	request = request.WithContext(context.Background())
	if err := request.Write(connection); err != nil {
		return nil, err
	}

	response, err := http.ReadResponse(bufio.NewReader(connection), request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusSwitchingProtocols {
		return nil, errors.New("unable to upgrade connection to websocket. Server returned status:" + response.Status)
	}
	if response.Header.Get("Sec-Websocket-Accept") != client_utils.ExpectedWebSocketAccept(websocketKey) {
		return nil, errors.New("unable to upgrade connection to websocket. Server failed WebSocket key challenge")
	}

	return &socket, nil
}

func (connection WebSocketConnection) SendMessage(message string) error {
	_, err := connection.TLSConnection.Write([]byte{0x81})
	if err != nil {
		return err
	}

	var messageBytes []byte = []byte(message)
	var payloadLength int = len(messageBytes)

	if payloadLength <= 125 {
		_, err := connection.TLSConnection.Write([]byte{byte(128 ^ len(messageBytes))})
		if err != nil {
			return err
		}
	} else {
		if payloadLength < 65536 {
			_, err := connection.TLSConnection.Write([]byte{126, byte(payloadLength >> 8), byte(payloadLength)})
			if err != nil {
				return err
			}
		} else {
			_, err := connection.TLSConnection.Write([]byte{127, byte(payloadLength >> 8), byte(payloadLength)})
			if err != nil {
				return err
			}
		}
	}
	var maskedMessage, mask []byte = client_utils.MaskMessage(messageBytes)
	_, err = connection.TLSConnection.Write(mask)
	if err != nil {
		return err
	}
	_, err = connection.TLSConnection.Write(maskedMessage)
	if err != nil {
		return err
	}
	return nil
}

func (connection WebSocketConnection) RecieveMessage() (string, error) {
	header := make([]byte, 2)

	if _, err := connection.TLSConnection.Read(header); err != nil {
		return "", err
	}

	payloadLength := int(header[1] & 0x7F)
	payload := make([]byte, payloadLength)

	if _, err := connection.TLSConnection.Read(payload); err != nil {
		return "", err
	}

	return string(payload), nil
}
