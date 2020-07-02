package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/deviceadm"
	"github.com/shellhub-io/shellhub/api/firewall"
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
)

const TenantIDHeader = "X-Tenant-ID"

type filterQuery struct {
	Filter string `query:"filter"`
	paginator.Query
}

func GetDeviceList(c apicontext.Context) error {
	svc := deviceadm.NewService(c.Store())

	query := filterQuery{}
	if err := c.Bind(&query); err != nil {
		return err
	}

	query.Normalize()

	devices, count, err := svc.ListDevices(c.Ctx(), query.Query, query.Filter)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, devices)
}

func GetDevice(c apicontext.Context) error {
	svc := deviceadm.NewService(c.Store())

	device, err := svc.GetDevice(c.Ctx(), models.UID(c.Param("uid")))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

func DeleteDevice(c apicontext.Context) error {
	svc := deviceadm.NewService(c.Store())

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	if err := svc.DeleteDevice(c.Ctx(), models.UID(c.Param("uid")), tenant); err != nil {
		if err == deviceadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}

		return err
	}

	return nil
}

func RenameDevice(c apicontext.Context) error {
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

	svc := deviceadm.NewService(c.Store())

	if err := svc.RenameDevice(c.Ctx(), models.UID(c.Param("uid")), req.Name, tenant); err != nil {
		if err == deviceadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}

		return err
	}

	return nil
}

func OfflineDevice(c apicontext.Context) error {
	svc := deviceadm.NewService(c.Store())

	if err := svc.UpdateDeviceStatus(c.Ctx(), models.UID(c.Param("uid")), false); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}

func LookupDevice(c apicontext.Context) error {
	var query struct {
		Domain    string `query:"domain"`
		Name      string `query:"name"`
		Username  string `query:"username"`
		IPAddress string `query:"ip_address"`
	}

	if err := c.Bind(&query); err != nil {
		return err
	}

	svc := deviceadm.NewService(c.Store())
	fw := firewall.NewService(c.Store())

	device, err := svc.LookupDevice(c.Ctx(), query.Domain, query.Name)
	if err != nil {
		return nil
	}

	ok, err := fw.Evaluate(c.Ctx(), firewall.Request{
		Hostname:  query.Name,
		Namespace: query.Domain,
		Username:  query.Username,
		IPAddress: query.IPAddress,
	})
	if err != nil {
		return err
	}

	if !ok {
		return c.NoContent(http.StatusForbidden)
	}

	return c.JSON(http.StatusOK, device)
}
