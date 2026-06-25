package auth

import (
	"github.com/pquerna/otp/totp"
)

type TOTPKey struct {
	Secret string
	URL    string
}

func GenerateTOTP(username string) (*TOTPKey, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Containerized CLI Login System",
		AccountName: username,
	})
	if err != nil {
		return nil, err
	}

	return &TOTPKey{Secret: key.Secret(), URL: key.URL()}, nil
}

func ValidateTOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}
