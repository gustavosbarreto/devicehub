// Package mode defines the interfaces used by the server to determine how to handle authentication and sessions.
package modes

import gliderssh "github.com/gliderlabs/ssh"

// Mode defines the SSH's server mode type.
type Mode interface {
	Authenticator
	Sessioner
}

// Authenticator defines the authentication methods used by the SSH's server.
//
//go:generate mockery --name=Authenticator --filename=authenticator.go
type Authenticator interface {
	// Password must be implemented to deal with password authentication.
	Password(ctx gliderssh.Context, user string, password string) bool
	// PublicKey must be implemented to deal with public key authentication.
	PublicKey(ctx gliderssh.Context, user string, key gliderssh.PublicKey) bool
}

// Sessioner defines the session methods used by the SSH's server to deal wihth determining the type of session.
//
//go:generate mockery --name=Sessioner --filename=sessioner.go
type Sessioner interface {
	Subsystemer
	// Shell must be implemented to deal with shell session.
	Shell(session gliderssh.Session) error
	// Heredoc must be implemented to deal with heredoc session.
	//
	// heredoc is special block of code that contains multi-line strings that will be redirected to a stdin of a shell.
	// It request a shell, but doesn't allocate a pty.
	//
	// An example of heredoc is:
	//  cat <<EOF
	//      test123
	//  EOF
	Heredoc(session gliderssh.Session) error
	// Exec must be implemented to deal with exec session.
	Exec(session gliderssh.Session) error
}

// Subsystemer defines the subsystem methods used by the SSH's server to deal with determining the type of subsystem.
//
// Subsystemer is a subset of the [Sessioner] interface.
type Subsystemer interface {
	// SFTP must be implemented to deal with SFTP session.
	SFTP(session gliderssh.Session) error
}
