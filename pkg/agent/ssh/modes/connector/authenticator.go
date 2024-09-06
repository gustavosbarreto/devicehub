package connector

import (
	"archive/tar"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"

	dockerclient "github.com/docker/docker/client"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/pkg/agent/ssh/modes"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// NOTICE: Ensures the Authenticator interface is implemented.
var _ modes.Authenticator = (*Authenticator)(nil)

// Authenticator implements the Authenticator interface when the server is running in connector mode.
type Authenticator struct {
	// api is a client to communicate with the ShellHub's API.
	api client.Client
	// authData is the authentication data received from the API to authenticate the device.
	authData *models.DeviceAuthResponse
	// container is the device name.
	//
	// NOTICE: Uses a pointer for later assignment.
	container *string
	// docker is a client to communicate with the Docker's API.
	docker dockerclient.APIClient
}

// NewAuthenticator creates a new instance of Authenticator for the connector mode.
func NewAuthenticator(api client.Client, docker dockerclient.APIClient, authData *models.DeviceAuthResponse, container *string) *Authenticator {
	return &Authenticator{
		api:       api,
		authData:  authData,
		container: container,
		docker:    docker,
	}
}

// getPasswd return a [io.Reader] for the container's passwd file.
func getPasswd(ctx context.Context, cli dockerclient.APIClient, container string) (io.Reader, error) {
	passwdTar, _, err := cli.CopyFromContainer(ctx, container, "/etc/passwd")
	if err != nil {
		return nil, err
	}

	passwd := tar.NewReader(passwdTar)
	if _, err := passwd.Next(); err != nil {
		return nil, err
	}

	return passwd, nil
}

// getShadow return a [io.Reader] for the container's shadow file.
func getShadow(ctx context.Context, cli dockerclient.APIClient, container string) (io.Reader, error) {
	shadowTar, _, err := cli.CopyFromContainer(ctx, container, "/etc/shadow")
	if err != nil {
		return nil, err
	}

	shadow := tar.NewReader(shadowTar)
	if _, err := shadow.Next(); err != nil {
		return nil, err
	}

	return shadow, nil
}

// Password handles the server's SSH password authentication when server is running in connector mode.
func (a *Authenticator) Password(ctx gliderssh.Context, username string, password string) bool {
	passwd, err := getPasswd(ctx, a.docker, *a.container)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to get the passwd file from container")

		return false
	}

	user, err := osauth.LookupUserFromPasswd(username, passwd)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to lookup for the user on passwd file")

		return false
	}

	if user.Password == "" {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("user passwd is empty, so the authentication via password is blocked")

		// NOTICE(r): when the user doesn't have password, we block the login.
		return false
	}

	shadow, err := getShadow(ctx, a.docker, *a.container)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to get the shadow file from container")

		return false
	}

	if !osauth.AuthUserFromShadow(username, password, shadow) {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to authenticate the user on the device")

		return false
	}

	// NOTICE: set the osauth.User to the context to be obtained later on.
	ctx.SetValue("user", user)

	log.WithFields(
		log.Fields{
			"container": *a.container,
			"username":  username,
		},
	).Info("using password authentication")

	return true
}

// PublicKey handles the server's SSH public key authentication when server is running in connector mode.
func (a *Authenticator) PublicKey(ctx gliderssh.Context, username string, key gliderssh.PublicKey) bool {
	passwd, err := getPasswd(ctx, a.docker, *a.container)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to get the passwd file from container")

		return false
	}

	user, err := osauth.LookupUserFromPasswd(username, passwd)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to lookup for the user on passwd file")

		return false
	}

	type Signature struct {
		Username  string
		Namespace string
	}

	sig := &Signature{
		Username:  username,
		Namespace: *a.container,
	}

	sigBytes, err := json.Marshal(sig)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container": *a.container,
				"username":  username,
			},
		).WithError(err).Error("failed to marshal signature")

		return false
	}

	sigHash := sha256.Sum256(sigBytes)

	fingerprint := gossh.FingerprintLegacyMD5(key)
	res, err := a.api.AuthPublicKey(&models.PublicKeyAuthRequest{
		Fingerprint: fingerprint,
		Data:        string(sigBytes),
	}, a.authData.Token)
	if err != nil {
		log.WithFields(
			log.Fields{
				"container":   *a.container,
				"username":    username,
				"fingerprint": fingerprint,
			},
		).WithError(err).Error("failed to authenticate the user via public key")

		return false
	}

	digest, err := base64.StdEncoding.DecodeString(res.Signature)
	if err != nil {
		if err != nil {
			log.WithFields(
				log.Fields{
					"container":   *a.container,
					"username":    username,
					"fingerprint": fingerprint,
				},
			).WithError(err).Error("failed to decode the signature")

			return false
		}

		return false
	}

	cryptoKey, ok := key.(gossh.CryptoPublicKey)
	if !ok {
		log.WithFields(
			log.Fields{
				"container":   *a.container,
				"username":    username,
				"fingerprint": fingerprint,
			},
		).Error("failed to get the crypto public key")

		return false
	}

	pubCrypto := cryptoKey.CryptoPublicKey()

	pubKey, ok := pubCrypto.(*rsa.PublicKey)
	if !ok {
		log.WithFields(
			log.Fields{
				"container":   *a.container,
				"username":    username,
				"fingerprint": fingerprint,
			},
		).Error("failed to convert the crypto public key")

		return false
	}

	if err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, sigHash[:], digest); err != nil {
		log.WithFields(
			log.Fields{
				"container":   *a.container,
				"username":    username,
				"fingerprint": fingerprint,
			},
		).WithError(err).Error("failed to verify the signature")

		return false
	}

	// NOTICE: set the osauth.User to the context to be obtained later on.
	ctx.SetValue("user", user)

	log.WithFields(
		log.Fields{
			"container":   *a.container,
			"username":    username,
			"fingerprint": fingerprint,
		},
	).Info("using public key authentication")

	return true
}
