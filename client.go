package sshutil

import (
	"golang.org/x/crypto/ssh"
)

// Client implements a traditional SSH client that supports shells,
// subprocesses, TCP port/streamlocal forwarding and tunneled dialing,
// and file operations.
type Client struct {
	*ssh.Client
}

// NewSession opens a new Session for this client. (A session is a remote
// execution of a program, or file operation.)
func (c *Client) NewSession() (*Session, error) {
	s, err := c.Client.NewSession()
	if err != nil {
		return nil, err
	}
	return &Session{
		Session: s,
	}, nil
}
