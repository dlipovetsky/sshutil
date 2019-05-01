package sshutil

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	gossh "golang.org/x/crypto/ssh"
)

// Session adds file system operations to golang.org/x/crypto/ssh.Session.
type Session struct {
	*gossh.Session
}

// ReadFile reads the remote file into a stream.
func (s *Session) ReadFile(dst io.Writer, filename string) error {
	cmd := fmt.Sprintf("cat %s", filename)
	s.Stdout = dst
	return s.run(cmd)
}

// ReadFileAs reads, as the user, the remote file into a stream.
func (s *Session) ReadFileAs(dst io.Writer, filename, user string) error {
	cmd := fmt.Sprintf("sudo --user=%s cat %s", user, filename)
	s.Stdout = dst
	return s.run(cmd)
}

// WriteFile writes the stream into a remote file.
func (s *Session) WriteFile(filename string, src io.Reader, perm os.FileMode) error {
	// Remove file first to ensure file is created with the correct owner.
	cmd := fmt.Sprintf("rm -f %[1]s && tee %[1]s > /dev/null && chmod %[2]d %[1]s", filename, perm)
	s.Stdin = src
	return s.run(cmd)
}

// WriteFileAs writes, as the user, the stream into a remote file.
func (s *Session) WriteFileAs(filename string, src io.Reader, perm os.FileMode, user string) error {
	// Remove file first to ensure file is created with the correct owner.
	cmd := fmt.Sprintf("sudo --user %[3]s rm -f %[1]s && sudo --user %[3]s tee %[1]s > /dev/null && sudo --user %[3]s chmod %[2]d %[1]s", filename, perm, user)
	s.Stdin = src
	return s.run(cmd)
}

// Mkdir creates a new directory with the specified name and permission
// bits (before umask).
func (s *Session) Mkdir(path string, perm os.FileMode) error {
	cmd := fmt.Sprintf("mkdir %[1]s && chmod %[2]d %[1]s", path, perm)
	return s.run(cmd)
}

// MkdirAs creates, as the chosen user, a new directory with the specified name
// and permission bits (before umask).
func (s *Session) MkdirAs(path string, perm os.FileMode, user string) error {
	cmd := fmt.Sprintf("sudo --user %[3]s mkdir %[1]s && sudo --user %[3]s chmod %[2]d %[1]s", path, perm, user)
	return s.run(cmd)
}

// Remove removes the named file or (empty) directory.
func (s *Session) Remove(name string) error {
	cmd := fmt.Sprintf("rm %s", name)
	return s.run(cmd)
}

// RemoveAs removes, as the user, the named file or (empty) directory.
func (s *Session) RemoveAs(name, user string) error {
	cmd := fmt.Sprintf("sudo --user %[2]s rm %[1]s", name, user)
	return s.run(cmd)
}

// RemoveAll removes the path and any children it contains. It removes
// everything it can but returns the first error it encounters. If the path does
// not exist, RemoveAll returns nil (no error).
func (s *Session) RemoveAll(path string) error {
	cmd := fmt.Sprintf("rm --recursive --force %s", path)
	return s.run(cmd)
}

// RemoveAllAs removes, as the user, the path and any children it contains. It
// removes everything it can but returns the first error it encounters. If the
// path does not exist, RemoveAll returns nil (no error).
func (s *Session) RemoveAllAs(path, user string) error {
	cmd := fmt.Sprintf("sudo --user %[2]s rm --recursive --force %[1]s", path, user)
	return s.run(cmd)
}

// Exists checks that the path exists.
func (s *Session) Exists(path string) (bool, error) {
	cmd := fmt.Sprintf("test -e %s", path)
	if err := s.run(cmd); err != nil {
		if hasExitStatus(err, 1) {
			// exit status 1 means the path does not exist
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistsAs checks, as the user, that the path exists.
func (s *Session) ExistsAs(path, user string) (bool, error) {
	cmd := fmt.Sprintf("sudo --user %[2]s test -e %[1]s", path, user)
	if err := s.run(cmd); err != nil {
		if hasExitStatus(err, 1) {
			// exit status 1 means the path does not exist
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Move moves the source path to the destination path, using the semantics of
// the `mv` command.
func (s *Session) Move(srcpath, dstpath string) error {
	cmd := fmt.Sprintf("mv %s %s", srcpath, dstpath)
	return s.run(cmd)
}

// MoveAs moves, as the user, the source path to the destination path, using the
// semantics of the `mv` command.
func (s *Session) MoveAs(srcpath, dstpath, user string) error {
	cmd := fmt.Sprintf("sudo --user %[3]s mv %[1]s %[2]s", srcpath, dstpath, user)
	return s.run(cmd)
}

// Copy copies the source path to the destination path, using the semantics of
// the `cp` command.
func (s *Session) Copy(srcpath, dstpath string) error {
	cmd := fmt.Sprintf("cp %s %s", srcpath, dstpath)
	return s.run(cmd)
}

// CopyAs moves, as the user, the source path to the destination path, using the
// semantics of the `cp` command.
func (s *Session) CopyAs(srcpath, dstpath, user string) error {
	cmd := fmt.Sprintf("sudo --user %[3]s cp %[1]s %[2]s", srcpath, dstpath, user)
	return s.run(cmd)
}

// run runs a remote command. The error will contain the stderr output, which
// often explains why the command failed. By comparison, the error returned by
// ssh.Run() contains only the exit code.
func (s *Session) run(cmd string) error {
	var errbuf bytes.Buffer
	s.Stderr = &errbuf
	if err := s.Run(cmd); err != nil {
		return errors.Wrap(err, string(errbuf.Bytes()))
	}
	return nil
}

func hasExitStatus(err error, status int) bool {
	if ee, ok := errors.Cause(err).(*gossh.ExitError); ok && ee.Waitmsg.ExitStatus() == status {
		return true
	}
	return false
}
