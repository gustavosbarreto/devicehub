package routes

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetPublicKeysURL    = "/sshkeys/public-keys"
	GetPublicKeyURL     = "/sshkeys/public-keys/:fingerprint/:tenant"
	CreatePublicKeyURL  = "/sshkeys/public-keys"
	UpdatePublicKeyURL  = "/sshkeys/public-keys/:fingerprint"
	DeletePublicKeyURL  = "/sshkeys/public-keys/:fingerprint"
	CreatePrivateKeyURL = "/sshkeys/private-keys"
	EvaluateKeyURL      = "/sshkeys/public-keys/evaluate/:fingerprint/:username"
)

const (
	ParamPublicKeyFingerprint = "fingerprint"
)

func (h *Handler) GetPublicKeys(c gateway.Context) error {
	paginator := query.NewPaginator()
	if err := c.Bind(paginator); err != nil {
		return err
	}

	// TODO: normalize is not required when request is privileged
	paginator.Normalize()

	list, count, err := h.service.ListPublicKeys(c.Ctx(), *paginator)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, list)
}

func (h *Handler) GetPublicKey(c gateway.Context) error {
	var req requests.PublicKeyGet
	err := c.Bind(&req)
	if err != nil {
		return err
	}

	err = c.Validate(&req)
	if err != nil {
		return err
	}

	pubKey, err := h.service.GetPublicKey(c.Ctx(), req.Fingerprint, req.Tenant)
	if err != nil {
		if err == store.ErrNoDocuments {
			return c.NoContent(http.StatusNotFound)
		}

		return err
	}

	return c.JSON(http.StatusOK, pubKey)
}

func (h *Handler) CreatePublicKey(c gateway.Context) error {
	var req requests.PublicKeyCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	var tenant string
	if c.Tenant() != nil {
		tenant = c.Tenant().ID
		req.TenantID = tenant
	}

	res, err := h.service.CreatePublicKey(c.Ctx(), req, tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdatePublicKey(c gateway.Context) error {
	var req requests.PublicKeyUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	var tenant string
	if c.Tenant() != nil {
		tenant = c.Tenant().ID
	}

	res, err := h.service.UpdatePublicKey(c.Ctx(), req.Fingerprint, tenant, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) DeletePublicKey(c gateway.Context) error {
	var req requests.PublicKeyDelete
	if err := c.Bind(&req); err != nil {
		return err
	}

	// NOTE: This is a temporary workaround.
	// TODO: Investigate why echo is not decoding the Fingerprint.
	req.Fingerprint, _ = url.QueryUnescape(req.Fingerprint)

	if err := c.Validate(&req); err != nil {
		return err
	}

	var tenant string
	if c.Tenant() != nil {
		tenant = c.Tenant().ID
	}

	if err := h.service.DeletePublicKey(c.Ctx(), req.Fingerprint, tenant); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) CreatePrivateKey(c gateway.Context) error {
	privKey, err := h.service.CreatePrivateKey(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, privKey)
}

func (h *Handler) EvaluateKey(c gateway.Context) error {
	var device models.Device
	if err := c.Bind(&device); err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	pubKey, err := h.service.GetPublicKey(c.Ctx(), c.Param(ParamPublicKeyFingerprint), device.TenantID)
	if err != nil {
		return c.JSON(http.StatusForbidden, err)
	}

	usernameOk, err := h.service.EvaluateKeyUsername(c.Ctx(), pubKey, c.Param(ParamUserName))
	if err != nil {
		return err
	}

	filterOk, err := h.service.EvaluateKeyFilter(c.Ctx(), pubKey, device)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, usernameOk && filterOk)
}
