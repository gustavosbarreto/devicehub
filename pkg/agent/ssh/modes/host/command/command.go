package command

type SFTPServerMode string

const (
	SFTPServerModeNative SFTPServerMode = "native"
	SFTPServerModeDocker SFTPServerMode = "docker"
)
