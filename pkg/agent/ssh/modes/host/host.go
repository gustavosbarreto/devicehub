// Package host defines authentication and sessions handles to SSH when it is running in host mode.
//
// Host mode means that the SSH's server runs in the host machine, using the host "/etc/passwd", "/etc/shadow",
// redirecting the SSH's connection to the device sdin, stdout and stderr and other things needed to run the SSH's
// server in the host machine.
package host

type Mode struct {
	Authenticator
	Sessioner
}
