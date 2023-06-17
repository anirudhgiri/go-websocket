package url_utils

import (
	"errors"
	"net/url"
)

func FormConnectionURL(serverURLString string) (*url.URL, error) {
	serverURL, err := url.Parse(serverURLString)
	if err != nil {
		return nil, err
	}

	err = setURIForTLSConnection(serverURL)
	if err != nil {
		return nil, err
	}

	return serverURL, nil
}

func GetServerPort(serverURL *url.URL) (string, error) {
	switch serverURL.Scheme {
	case "http", "ws":
		return ":80", nil
	case "https", "wss":
		return ":443", nil
	}

	return "", errors.New("websocket: invalid server URI")
}

func setURIForTLSConnection(serverURL *url.URL) error {
	switch serverURL.Scheme {
	case "ws", "http":
		serverURL.Scheme = "http"
	case "wss", "https":
		serverURL.Scheme = "https"
	default:
		return errors.New("websocket: invalid server URI")
	}

	return nil
}
