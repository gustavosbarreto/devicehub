package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetSessionsURL             = "/sessions"
	GetSessionURL              = "/sessions/:uid"
	SetSessionAuthenticatedURL = "/sessions/:uid"
	CreateSessionURL           = "/sessions"
	FinishSessionURL           = "/sessions/:uid/finish"
	KeepAliveSessionURL        = "/sessions/:uid/keepalive"
	RecordSessionURL           = "/sessions/:uid/record"
	PlaySessionURL             = "/sessions/:uid/play"
)

const (
	ParamSessionID = "uid"
)

func (h *Handler) GetSessionList(c gateway.Context) error {
	query := paginator.NewQuery()
	if err := c.Bind(query); err != nil {
		return err
	}

	// TODO: normalize is not required when request is privileged
	query.Normalize()

	sessions, count, err := h.service.ListSessions(c.Ctx(), *query)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, sessions)
}

func (h *Handler) GetSession(c gateway.Context) error {
	session, err := h.service.GetSession(c.Ctx(), models.UID(c.Param(ParamSessionID)))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}

func (h *Handler) SetSessionAuthenticated(c gateway.Context) error {
	var req struct {
		Authenticated bool `json:"authenticated"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	return h.service.SetSessionAuthenticated(c.Ctx(), models.UID(c.Param(ParamSessionID)), req.Authenticated)
}

func (h *Handler) CreateSession(c gateway.Context) error {
	session := new(models.Session)

	if err := c.Bind(&session); err != nil {
		return err
	}

	session, err := h.service.CreateSession(c.Ctx(), *session)
	if err != nil {
		return err
	}

	ip := c.Request().Header.Get("X-Real-IP")
	err = h.service.SetDevicePosition(c.Ctx(), session.DeviceUID, ip)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}

func (h *Handler) FinishSession(c gateway.Context) error {
	return h.service.DeactivateSession(c.Ctx(), models.UID(c.Param(ParamSessionID)))
}

func (h *Handler) KeepAliveSession(c gateway.Context) error {
	return h.service.KeepAliveSession(c.Ctx(), models.UID(c.Param(ParamSessionID)))
}

func (h *Handler) RecordSession(c gateway.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *Handler) PlaySession(c gateway.Context) error {
	return c.NoContent(http.StatusOK)
}

func (h *Handler) DeleteRecordedSession(c gateway.Context) error {
	return c.NoContent(http.StatusOK)
}
