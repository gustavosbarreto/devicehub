package services

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"strings"
	"time"

	"github.com/cnf/structhash"
	"github.com/go-playground/validator/v10"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type AuthService interface {
	AuthDevice(ctx context.Context, req *models.DeviceAuthRequest, remoteAddr string) (*models.DeviceAuthResponse, error)
	AuthUser(ctx context.Context, req models.UserAuthRequest) (*models.UserAuthResponse, error)
	AuthGetToken(ctx context.Context, tenant string) (*models.UserAuthResponse, error)
	AuthPublicKey(ctx context.Context, req *models.PublicKeyAuthRequest) (*models.PublicKeyAuthResponse, error)
	AuthSwapToken(ctx context.Context, ID, tenant string) (*models.UserAuthResponse, error)
	AuthUserInfo(ctx context.Context, username, tenant, token string) (*models.UserAuthResponse, error)
	PublicKey() *rsa.PublicKey
}

func (s *service) AuthDevice(ctx context.Context, req *models.DeviceAuthRequest, remoteAddr string) (*models.DeviceAuthResponse, error) {
	uid := sha256.Sum256(structhash.Dump(req.DeviceAuth, 1))

	key := hex.EncodeToString(uid[:])

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.DeviceAuthClaims{
		UID: key,
		AuthClaims: models.AuthClaims{
			Claims: "device",
		},
	})

	tokenStr, err := token.SignedString(s.privKey)
	if err != nil {
		return nil, err
	}

	type Device struct {
		Name      string
		Namespace string
	}

	var value *Device

	if err := s.cache.Get(ctx, strings.Join([]string{"auth_device", key}, "/"), &value); err == nil && value != nil {
		return &models.DeviceAuthResponse{
			UID:       key,
			Token:     tokenStr,
			Name:      value.Name,
			Namespace: value.Namespace,
		}, nil
	}
	device := models.Device{
		UID:        key,
		Identity:   req.Identity,
		Info:       req.Info,
		PublicKey:  req.PublicKey,
		TenantID:   req.TenantID,
		LastSeen:   clock.Now(),
		RemoteAddr: remoteAddr,
	}

	// The order here is critical as we don't want to register devices if the tenant id is invalid
	namespace, err := s.store.NamespaceGet(ctx, device.TenantID)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, err
	}
	hostname := strings.ToLower(req.DeviceAuth.Hostname)

	if err := s.store.DeviceCreate(ctx, device, hostname); err != nil {
		return nil, err
	}

	if err := s.store.DeviceSetOnline(ctx, models.UID(device.UID), true); err != nil {
		return nil, err
	}

	for _, uid := range req.Sessions {
		if err := s.store.SessionSetLastSeen(ctx, models.UID(uid)); err != nil {
			continue
		}
	}

	dev, err := s.store.DeviceGetByUID(ctx, models.UID(device.UID), device.TenantID)
	if err != nil {
		return nil, err
	}
	if err := s.cache.Set(ctx, strings.Join([]string{"auth_device", key}, "/"), &Device{Name: dev.Name, Namespace: namespace.Name}, time.Second*30); err != nil {
		return nil, err
	}

	return &models.DeviceAuthResponse{
		UID:       key,
		Token:     tokenStr,
		Name:      dev.Name,
		Namespace: namespace.Name,
	}, nil
}

func (s *service) AuthUser(ctx context.Context, req models.UserAuthRequest) (*models.UserAuthResponse, error) {
	user, err := s.store.UserGetByUsername(ctx, strings.ToLower(req.Username))
	if err != nil {
		user, err = s.store.UserGetByEmail(ctx, strings.ToLower(req.Username))
		if err != nil {
			return nil, err
		}
	}

	if !user.Confirmed {
		return nil, ErrForbidden
	}

	namespace, _ := s.store.NamespaceGetFirst(ctx, user.ID)

	role := ""
	tenant := ""
	if namespace != nil {
		tenant = namespace.TenantID

		for _, member := range namespace.Members {
			if member.ID == user.ID {
				role = member.Role

				break
			}
		}
	}

	password := sha256.Sum256([]byte(req.Password))
	if user.Password == hex.EncodeToString(password[:]) {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
			Username: user.Username,
			Admin:    true,
			Tenant:   tenant,
			Role:     role,
			ID:       user.ID,
			AuthClaims: models.AuthClaims{
				Claims: "user",
			},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(clock.Now().Add(time.Hour * 72)),
			},
		})

		tokenStr, err := token.SignedString(s.privKey)
		if err != nil {
			return nil, err
		}

		user.LastLogin = clock.Now()

		if err := s.store.UserUpdateData(ctx, user, user.ID); err != nil {
			return nil, err
		}

		return &models.UserAuthResponse{
			Token:  tokenStr,
			Name:   user.Name,
			ID:     user.ID,
			User:   user.Username,
			Tenant: tenant,
			Role:   role,
			Email:  user.Email,
		}, nil
	}

	return nil, ErrUnauthorized
}

func (s *service) AuthGetToken(ctx context.Context, id string) (*models.UserAuthResponse, error) {
	user, _, err := s.store.UserGetByID(ctx, id, false)
	if err != nil {
		return nil, err
	}

	namespace, _ := s.store.NamespaceGetFirst(ctx, user.ID)

	role := ""
	tenant := ""
	if namespace != nil {
		tenant = namespace.TenantID

		for _, member := range namespace.Members {
			if member.ID == user.ID {
				role = member.Role

				break
			}
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
		Username: user.Username,
		Admin:    true,
		Tenant:   tenant,
		Role:     role,
		ID:       user.ID,
		AuthClaims: models.AuthClaims{
			Claims: "user",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(clock.Now().Add(time.Hour * 72)),
		},
	})

	tokenStr, err := token.SignedString(s.privKey)
	if err != nil {
		return nil, err
	}

	return &models.UserAuthResponse{
		Token:  tokenStr,
		Name:   user.Name,
		ID:     user.ID,
		User:   user.Username,
		Tenant: tenant,
		Role:   role,
		Email:  user.Email,
	}, nil
}

func (s *service) AuthPublicKey(ctx context.Context, req *models.PublicKeyAuthRequest) (*models.PublicKeyAuthResponse, error) {
	privKey, err := s.store.PrivateKeyGet(ctx, req.Fingerprint)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privKey.Data)
	if block == nil {
		return nil, err
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	digest := sha256.Sum256([]byte(req.Data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return nil, err
	}

	return &models.PublicKeyAuthResponse{
		Signature: base64.StdEncoding.EncodeToString(signature),
	}, nil
}

func (s *service) AuthSwapToken(ctx context.Context, id, tenant string) (*models.UserAuthResponse, error) {
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return nil, err
	}

	user, _, err := s.store.UserGetByID(ctx, id, false)
	if err != nil {
		return nil, err
	}

	var role string
	for _, member := range namespace.Members {
		if member.ID == user.ID {
			role = member.Role

			break
		}
	}

	for _, i := range namespace.Members {
		if user.ID == i.ID {
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
				Username: user.Username,
				Admin:    true,
				Tenant:   namespace.TenantID,
				Role:     role,
				ID:       user.ID,
				AuthClaims: models.AuthClaims{
					Claims: "user",
				},
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(clock.Now().Add(time.Hour * 72)),
				},
			})

			tokenStr, err := token.SignedString(s.privKey)
			if err != nil {
				return nil, err
			}

			return &models.UserAuthResponse{
				Token:  tokenStr,
				Name:   user.Name,
				ID:     user.ID,
				User:   user.Username,
				Role:   role,
				Tenant: namespace.TenantID,
				Email:  user.Email,
			}, nil
		}
	}

	return nil, nil
}

func (s *service) AuthUserInfo(ctx context.Context, username, tenant, token string) (*models.UserAuthResponse, error) {
	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		if err == store.ErrNoDocuments {
			return nil, ErrUnauthorized
		}

		return nil, err
	}

	namespace, _ := s.store.NamespaceGet(ctx, tenant)

	role := ""
	if namespace != nil {
		for _, member := range namespace.Members {
			if member.ID == user.ID {
				role = member.Role

				break
			}
		}
	}

	return &models.UserAuthResponse{
		Token:  token,
		Name:   user.Name,
		User:   user.Username,
		Tenant: tenant,
		Role:   role,
		ID:     user.ID,
		Email:  user.Email,
	}, nil
}

func (s *service) PublicKey() *rsa.PublicKey {
	return s.pubKey
}
