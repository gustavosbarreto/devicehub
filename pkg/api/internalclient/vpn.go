package internalclient

import "net/http"

type vpnAPI interface {
	// VPNStopRouter sends a rquest to VPN service to stop the namespace router.
	VPNStopRouter(tenant string) error
}

func (c *client) VPNStopRouter(tenant string) error {
	res, err := c.http.
		R().
		Delete("http://vpn:8080/vpn/router/" + tenant)
	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		return err
	}

	return nil
}
