package tunnel

import (
	"context"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/revdial"
)

type Tunnel struct {
	router       *echo.Echo
	srv          *http.Server
	HTTPHandler  func(e echo.Context) error
	ConnHandler  func(e echo.Context) error
	CloseHandler func(e echo.Context) error
}

type Builder struct {
	tunnel *Tunnel
}

func NewBuilder() *Builder {
	return &Builder{
		tunnel: NewTunnel(),
	}
}

func (t *Builder) WithHTTPHandler(handler func(e echo.Context) error) *Builder {
	t.tunnel.HTTPHandler = handler

	return t
}

func (t *Builder) WithConnHandler(handler func(e echo.Context) error) *Builder {
	t.tunnel.ConnHandler = handler

	return t
}

func (t *Builder) WithCloseHandler(handler func(e echo.Context) error) *Builder {
	t.tunnel.CloseHandler = handler

	return t
}

func (t *Builder) Build() *Tunnel {
	return t.tunnel
}

func NewTunnel() *Tunnel {
	e := echo.New()

	t := &Tunnel{
		router: e,
		srv: &http.Server{
			Handler: e,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				return context.WithValue(ctx, "http-conn", c) //nolint:revive
			},
		},
		HTTPHandler: func(e echo.Context) error {
			panic("HTTPHandler can not be nil")
		},
		ConnHandler: func(e echo.Context) error {
			panic("connHandler can not be nil")
		},
		CloseHandler: func(e echo.Context) error {
			panic("closeHandler can not be nil")
		},
	}
	e.GET("/ssh/http", func(e echo.Context) error {
		return t.HTTPHandler(e)
	})
	e.GET("/ssh/:id", func(e echo.Context) error {
		return t.ConnHandler(e)
	})
	e.GET("/ssh/close/:id", func(e echo.Context) error {
		return t.CloseHandler(e)
	})

	return t
}

const ContextKeyHTTPConn string = "http-conn"

// NewCustomTunnel creates a new [Tunnel] with the route to the connect, in a POST, and close, in a DELETE, actions.
func NewCustomTunnel(connPath string, closePath string) *Tunnel {
	router := echo.New()

	t := &Tunnel{
		router: router,
		srv: &http.Server{
			Handler: router,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				return context.WithValue(ctx, ContextKeyHTTPConn, c) //nolint:revive
			},
		},
		ConnHandler: func(e echo.Context) error {
			panic("connHandler can not be nil")
		},
		CloseHandler: func(e echo.Context) error {
			panic("closeHandler can not be nil")
		},
	}

	router.POST(connPath, func(e echo.Context) error {
		return t.ConnHandler(e)
	})
	router.DELETE(closePath, func(e echo.Context) error {
		return t.CloseHandler(e)
	})

	return t
}

// Listen to reverse listener.
func (t *Tunnel) Listen(l *revdial.Listener) error {
	return t.srv.Serve(l)
}

// Close closes the tunnel.
func (t *Tunnel) Close() error {
	if err := t.router.Close(); err != nil {
		return err
	}

	return t.srv.Close()
}
