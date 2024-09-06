package vpn

import (
	"net"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func handler(handler func(net.Conn, *Settings) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		log.Debug("handler started")
		defer log.Debug("handler done")

		conn, _, err := c.Response().Hijack()
		if err != nil {
			log.Error(err)

			return err
		}

		defer conn.Close()

		settings, err := ParseSettings(c.Request().Body)
		if err != nil {
			log.WithError(err).Error("faild to parse the settings")

			return err
		}

		// NOTE: the [handler] is called to handler the core logic of the VPN client, while this handler is used to extract
		// the connection and the settings data.
		if err := handler(conn, settings); err != nil {
			log.WithError(err).Error("failed to handler the vpn connection between server and agent")

			return err
		}

		return nil
	}
}

func closeHandler(callback func() error) func(c echo.Context) error {
	return func(c echo.Context) error {
		log.Trace("close handler called")

		return callback()
	}
}
