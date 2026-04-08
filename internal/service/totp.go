package service

import (
	"bytes"
	"encoding/base64"
	"image/png"

	"github.com/pquerna/otp/totp"
)

func GenerateTOTP(username, issuer string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
	})
	if err != nil {
		return "", "", err
	}

	img, err := key.Image(200, 200)
	if err != nil {
		return "", "", err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", err
	}
	qrBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	return key.Secret(), qrBase64, nil
}

func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}
