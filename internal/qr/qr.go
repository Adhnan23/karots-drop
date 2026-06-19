package qr

import (
	qrcode "github.com/skip2/go-qrcode"
)

func Terminal(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}
	return qr.ToString(false), nil
}

func PNG(content string) ([]byte, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	return qr.PNG(256)
}
