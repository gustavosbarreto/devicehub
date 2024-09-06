package host

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/go-playground/assert/v2"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/osauth"
	osauthMocks "github.com/shellhub-io/shellhub/pkg/agent/pkg/osauth/mocks"
	clientMocks "github.com/shellhub-io/shellhub/pkg/api/client/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	gossh "golang.org/x/crypto/ssh"
)

func TestPublicKey(t *testing.T) {
	// stringToRef is a helper function to convert a string to a pointer to a string.
	stringToRef := func(s string) *string { return &s }

	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	key, _ := gossh.NewPublicKey(&privKey.PublicKey)

	tests := []struct {
		ctx           gliderssh.Context
		authenticator *Authenticator
		name          string
		user          string
		key           gliderssh.PublicKey
		requiredMocs  func(apiMock *clientMocks.Client)
		expected      bool
	}{
		{
			ctx: &testSSHContext{
				user: "",
			},
			authenticator: &Authenticator{
				deviceName: stringToRef("device"),
				api:        new(clientMocks.Client),
			},
			name:         "return false when user is not found",
			user:         "",
			key:          nil,
			requiredMocs: func(apiMock *clientMocks.Client) {},
			expected:     false,
		},
		{
			ctx: &testSSHContext{
				user: "test",
			},
			authenticator: &Authenticator{
				authData: &models.DeviceAuthResponse{
					Token: "token",
				},
				singleUserPassword: "",
				deviceName:         stringToRef("device"),
				api:                new(clientMocks.Client),
			},
			name: "return false when public key api request fails",
			user: "",
			key:  key,
			requiredMocs: func(apiMock *clientMocks.Client) {
				apiMock.On("AuthPublicKey", mock.Anything, "token").Return(nil, errors.New("error")).Once()
			},
			expected: false,
		},
		{
			ctx: &testSSHContext{
				user: "test",
			},
			authenticator: &Authenticator{
				authData: &models.DeviceAuthResponse{
					Token: "token",
				},
				singleUserPassword: "",
				deviceName:         stringToRef("device"),
				api:                new(clientMocks.Client),
			},
			name: "return false when public key signature is invalid",
			user: "",
			key:  key,
			requiredMocs: func(apiMock *clientMocks.Client) {
				apiMock.On("AuthPublicKey", mock.Anything, "token").Return(&models.PublicKeyAuthResponse{
					Signature: "signature",
				}, nil).Once()
			},
			expected: false,
		},
		{
			ctx: &testSSHContext{
				user: "test",
			},
			authenticator: &Authenticator{
				authData: &models.DeviceAuthResponse{
					Token: "token",
				},
				singleUserPassword: "",
				deviceName:         stringToRef("device"),
				api:                new(clientMocks.Client),
			},
			name: "return true when public key signature does not implement crypto.PublicKey",
			user: "",
			key:  key,
			requiredMocs: func(apiMock *clientMocks.Client) {
				apiMock.On("AuthPublicKey", mock.Anything, "token").Return(&models.PublicKeyAuthResponse{
					Signature: base64.StdEncoding.EncodeToString([]byte("signature")),
				}, nil).Once()
			},
			expected: false,
		},
		{
			ctx: &testSSHContext{
				user: "test",
			},
			authenticator: &Authenticator{
				authData: &models.DeviceAuthResponse{
					Token: "token",
				},
				singleUserPassword: "",
				deviceName:         stringToRef("device"),
				api:                new(clientMocks.Client),
			},
			name: "fail when public key returned by crypto.PublicKey is not a pointer to a rsa.PublicKey",
			user: "",
			key:  key,
			requiredMocs: func(apiMock *clientMocks.Client) {
				apiMock.On("AuthPublicKey", mock.Anything, "token").Return(&models.PublicKeyAuthResponse{
					Signature: base64.StdEncoding.EncodeToString([]byte("signature")),
				}, nil).Once()
			},
			expected: false,
		},
		{
			ctx: &testSSHContext{
				user: "test",
			},
			authenticator: &Authenticator{
				authData: &models.DeviceAuthResponse{
					Token: "token",
				},
				singleUserPassword: "",
				deviceName:         stringToRef("device"),
				api:                new(clientMocks.Client),
			},
			name: "return false when public key returned by crypto.PublicKey does not pass on rsa.VerifyPKCS1v15",
			user: "",
			key:  key,
			requiredMocs: func(apiMock *clientMocks.Client) {
				apiMock.On("AuthPublicKey", mock.Anything, "token").Return(&models.PublicKeyAuthResponse{
					Signature: base64.StdEncoding.EncodeToString([]byte("signature")),
				}, nil).Once()
			},
			expected: false,
		},
		{
			ctx: &testSSHContext{
				user: "test",
			},
			authenticator: &Authenticator{
				authData: &models.DeviceAuthResponse{
					Token: "token",
				},
				singleUserPassword: "",
				deviceName:         stringToRef("device"),
				api:                new(clientMocks.Client),
			},
			name: "return true when public key signature is valid",
			user: "",
			key:  key,
			requiredMocs: func(apiMock *clientMocks.Client) {
				type Signature struct {
					Username  string
					Namespace string
				}

				sigBytes, _ := json.Marshal(&Signature{
					Username:  "test",
					Namespace: "device",
				})

				digest := sha256.Sum256(sigBytes)

				signature, _ := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, digest[:])

				apiMock.On("AuthPublicKey", &models.PublicKeyAuthRequest{
					Fingerprint: gossh.FingerprintLegacyMD5(key),
					Data:        string(sigBytes),
				}, "token").Return(&models.PublicKeyAuthResponse{
					Signature: base64.StdEncoding.EncodeToString(signature),
				}, nil).Once()
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.requiredMocs(tt.authenticator.api.(*clientMocks.Client))

			ok := tt.authenticator.PublicKey(tt.ctx, tt.user, tt.key)
			assert.Equal(t, tt.expected, ok)
		})
	}
}

func TestPassword(t *testing.T) {
	mock := &osauthMocks.Backend{}
	osauth.DefaultBackend = mock

	tests := []struct {
		ctx           gliderssh.Context
		authenticator *Authenticator
		name          string
		user          string
		password      string
		requiredMocks func()
		expected      bool
	}{
		{
			ctx:           &testSSHContext{user: "test"},
			authenticator: &Authenticator{},
			name:          "return false when user or password are invalid",
			user:          "",
			password:      "password",
			requiredMocks: func() {
				mock.On("AuthUser", "test", "password").Return(false).Once()
			},
			expected: false,
		},
		{
			ctx:           &testSSHContext{user: "test"},
			authenticator: &Authenticator{},
			name:          "return true when user and password are valid",
			user:          "",
			password:      "test",
			requiredMocks: func() {
				mock.On("AuthUser", "test", "test").Return(true).Once()
			},
			expected: true,
		},
		{
			ctx: &testSSHContext{user: "test"},
			authenticator: &Authenticator{
				singleUserPassword: "test",
			},
			name:          "return false when single user is enabled and password is invalid",
			user:          "",
			password:      "password",
			requiredMocks: func() {},
			expected:      false,
		},
		{
			ctx: &testSSHContext{user: "test"},
			authenticator: &Authenticator{
				singleUserPassword: "$6$Ntq5PynhGPFJuhxn$emiTnyA.GTsvK6JjjrecwDSB3jywkoHky9ZuJAYwSGFlZU2npTFOEMVPYG7CsDLRyvUE7OzbqFidYuKO274DC.",
			},
			name:          "return true when single user is enabled and password is valid",
			user:          "",
			password:      "test",
			requiredMocks: func() {},
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.requiredMocks()

			got := tt.authenticator.Password(tt.ctx, tt.user, tt.password)
			assert.Equal(t, tt.expected, got)
		})
	}
}
