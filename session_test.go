package sshutil_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/dlipovetsky/sshutil"
	gliderssh "github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("failed to listen on a port: %v", err))
		}
	}
	return l
}

func newClientSession(addr string, config *gossh.ClientConfig) (*gossh.Session, *gossh.Client, error) {
	if config == nil {
		config = &gossh.ClientConfig{
			User: "testuser",
			Auth: []gossh.AuthMethod{
				gossh.Password("testpass"),
			},
		}
	}
	if config.HostKeyCallback == nil {
		config.HostKeyCallback = gossh.InsecureIgnoreHostKey()
	}
	client, err := gossh.Dial("tcp", addr, config)
	if err != nil {
		return nil, nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, nil, err
	}
	return session, client, err
}

func newTestSession(srv *gliderssh.Server, cfg *gossh.ClientConfig) (*sshutil.Session, func(), error) {
	l := newLocalListener()
	go func() {
		err := srv.Serve(l)
		if err != gliderssh.ErrServerClosed {
			return
		}
	}()
	s, c, err := newClientSession(l.Addr().String(), cfg)
	if err != nil {
		return nil, nil, err
	}
	return &sshutil.Session{Session: s}, func() {
		s.Close()
		c.Close()
		srv.Close()
	}, nil
}

func TestReadFile(t *testing.T) {
	testCases := []struct {
		Name          string
		Handler       func(s gliderssh.Session)
		WantContents  string
		WantErrString string
	}{
		{
			Name: "should read file",
			Handler: func(s gliderssh.Session) {
				io.WriteString(s, "sample file")
			},
			WantContents: "sample file",
		},
		{
			Name: "should fail if file not found",
			Handler: func(s gliderssh.Session) {
				io.WriteString(s.Stderr(), "cat: foo: No such file or directory")
				s.Exit(1)
			},
			WantErrString: "cat: foo: No such file or directory: Process exited with status 1",
		},
	}

	for _, tc := range testCases {
		srv := &gliderssh.Server{
			Handler: tc.Handler,
		}
		s, cleanup, err := newTestSession(srv, nil)
		if err != nil {
			t.Fatalf("error creating session: %s", err)
		}
		defer cleanup()
		var dst bytes.Buffer
		err = s.ReadFile(&dst, "foo")
		if tc.WantErrString != "" {
			if err == nil {
				t.Fatalf("%s failed: wanted error %s, got no error", tc.Name, tc.WantErrString)
			}
			if tc.WantErrString != err.Error() {
				t.Fatalf("%s failed: wanted error %s, got %v", tc.Name, tc.WantErrString, err.Error())
			}
			continue
		}
		gotContents := dst.String()
		if tc.WantContents != dst.String() {
			t.Fatalf("%s failed: wanted %s, got %s", tc.Name, tc.WantContents, gotContents)
		}
	}
}
