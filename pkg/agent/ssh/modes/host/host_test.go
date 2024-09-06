package host

import (
	"context"
	"net"
	"sync"

	gliderssh "github.com/gliderlabs/ssh"
)

type testSSHContext struct {
	context.Context
	*sync.Mutex

	user          string
	sessionID     string
	clientVersion string
	serverVersion string
	remoteAddr    net.Addr
	localAddr     net.Addr
	permissions   *gliderssh.Permissions
}

func (ctx *testSSHContext) SetValue(key, value interface{}) {
	ctx.Context = context.WithValue(ctx.Context, key, value)
}

func (ctx *testSSHContext) User() string {
	return ctx.user
}

func (ctx *testSSHContext) SessionID() string {
	return ctx.sessionID
}

func (ctx *testSSHContext) ClientVersion() string {
	return ctx.clientVersion
}

func (ctx *testSSHContext) ServerVersion() string {
	return ctx.serverVersion
}

func (ctx *testSSHContext) RemoteAddr() net.Addr {
	return ctx.remoteAddr
}

func (ctx *testSSHContext) LocalAddr() net.Addr {
	return ctx.localAddr
}

func (ctx *testSSHContext) Permissions() *gliderssh.Permissions {
	return ctx.permissions
}
