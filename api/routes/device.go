package routes

import (
	"net/http"
	"strconv"

	//	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	//	"github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetDeviceListURL = "/devices"
	GetDeviceURL     = "/devices/:uid"
	DeleteDeviceURL  = "/devices/:uid"
	RenameDeviceURL  = "/devices/:uid"
	OfflineDeviceURL = "/devices/:uid/offline"
	LookupDeviceURL  = "/lookup"
	UpdateStatusURL  = "/devices/:uid/:status"
)

const TenantIDHeader = "X-Tenant-ID"

type filterQuery struct {
	Filter string `query:"filter"`
	paginator.Query
	Status  string `query:"status"`
	SortBy  string `query:"sort_by"`
	OrderBy string `query:"order_by"`
}

func (h *Handler) GetDeviceList(c apicontext.Context) error {
	query := filterQuery{}
	if err := c.Bind(&query); err != nil {
		return err
	}

	query.Normalize()

	devices, count, err := h.service.ListDevices(c.Ctx(), query.Query, query.Filter, query.Status, query.SortBy, query.OrderBy)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, devices)
}

func (h *Handler) GetDevice(c apicontext.Context) error {
	device, err := h.service.GetDevice(c.Ctx(), models.UID(c.Param("uid")))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

func (h *Handler) DeleteDevice(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	if err := h.service.DeleteDevice(c.Ctx(), models.UID(c.Param("uid")), tenant, id); err != nil {
		if err == services.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}

		return err
	}

	return nil
}

func (h *Handler) RenameDevice(c apicontext.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	err := h.service.RenameDevice(c.Ctx(), models.UID(c.Param("uid")), req.Name, tenant, id)
	switch err {
	case services.ErrUnauthorized:
		return c.NoContent(http.StatusForbidden)
	case services.ErrDuplicatedDeviceName:
		return c.NoContent(http.StatusConflict)
	case services.ErrInvalidFormat:
		return c.NoContent(http.StatusBadRequest)
	default:
		return err
	}
}

func (h *Handler) OfflineDevice(c apicontext.Context) error {
	if err := h.service.UpdateDeviceStatus(c.Ctx(), models.UID(c.Param("uid")), false); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}

func (h *Handler) LookupDevice(c apicontext.Context) error {
	var query struct {
		Domain    string `query:"domain"`
		Name      string `query:"name"`
		Username  string `query:"username"`
		IPAddress string `query:"ip_address"`
	}

	if err := c.Bind(&query); err != nil {
		return err
	}

	device, err := h.service.LookupDevice(c.Ctx(), query.Domain, query.Name)
	if err != nil {
		return nil
	}

	return c.JSON(http.StatusOK, device)
}

func (h *Handler) UpdatePendingStatus(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	status := map[string]string{
		"accept":  "accepted",
		"reject":  "rejected",
		"pending": "pending",
		"unused":  "unused",
	}

	if err := h.service.UpdatePendingStatus(c.Ctx(), models.UID(c.Param("uid")), status[c.Param("status")], tenant, id); err != nil {
		if err == services.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}

		return err
	}

	return c.JSON(http.StatusOK, nil)
}

func IsDeviceOwner(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Get("ctx").(*apicontext.Context)
		id := ""
		if v := ctx.ID(); v != nil {
			id = v.ID
		}
		device, err := ctx.Service().(services.Service).GetDevice(ctx.Ctx(), models.UID(ctx.Param("uid")))
		if err != nil {
			return err
		}

		if err := ctx.Service().(services.Service).IsNamespaceOwner(ctx.Ctx(), device.TenantID, id); err != nil {
			return c.NoContent(http.StatusForbidden)
		}

		return next(c)
	}
}

func IsDeviceMember(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Get("ctx").(*apicontext.Context)
		id := ""
		if v := ctx.ID(); v != nil {
			id = v.ID
		}
		device, err := ctx.Service().(services.Service).GetDevice(ctx.Ctx(), models.UID(ctx.Param("uid")))
		if err != nil {
			return err
		}

		if err := ctx.Service().(services.Service).IsNamespaceMember(ctx.Ctx(), device.TenantID, id); err != nil {
			return c.NoContent(http.StatusForbidden)
		}

		return next(c)
	}
}
