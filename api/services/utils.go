package services

import (
	"crypto/rsa"
	"os"
	"slices"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/models"
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

func containsTags(list []models.Tags, item string) bool {
	return slices.ContainsFunc(list, func(n models.Tags) bool {
		return n.Name == item
	})
}

func contains(list []string, item string) bool {
	return slices.ContainsFunc(list, func(n string) bool {
		return n == item
	})
}
