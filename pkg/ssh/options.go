package ssh

type Options func(*SSHManager)

func WithPath(path string) Options {
	return func(ssh *SSHManager) {
		ssh.path = path
	}
}

func WithPort(port int) Options {
	return func(ssh *SSHManager) {
		ssh.port = port
	}
}

func WithKeyName(name string) Options {
	return func(ssh *SSHManager) {
		ssh.keyName = name
	}
}
