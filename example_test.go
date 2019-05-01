package sshutil_test

import (
	"bytes"
	"fmt"
	"io"
	"log"

	gliderssh "github.com/gliderlabs/ssh"
)

func ExampleSession_ReadFile() {
	srv := &gliderssh.Server{
		Handler: func(s gliderssh.Session) {
			io.WriteString(s, "sample file")
		},
	}
	s, cleanup, err := newTestSession(srv, nil)
	if err != nil {
		log.Fatalf("error creating session: %s", err)
	}
	defer cleanup()
	var dst bytes.Buffer
	err = s.ReadFile(&dst, "foo")
	if err != nil {
		log.Fatalf("error reading file: %s", err)
	}
	fmt.Println(dst.String())
	// Output: sample file
}
