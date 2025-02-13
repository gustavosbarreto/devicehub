package services

import (
	"crypto/rsa"
	"os"

	jwt "github.com/golang-jwt/jwt/v4"
)

func LoadKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	signBytes, err := os.ReadFile(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		return nil, nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, nil, err
	}

	verifyBytes, err := os.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, nil, err
	}

	return privKey, pubKey, nil
}
