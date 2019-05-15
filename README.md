# sshutil

This package adds some file operations to the standard go ssh library.

## Goals

- Perform a small set of file operations on hosts that do not have `scp` or `sftp`.
- Perform the file operations as user different from the login user.
- Compose well with the [https://godoc.org/golang.org/x/crypto/ssh](https://godoc.org/golang.org/x/crypto/ssh) package.

## Non-Goals

- Support the entire set of SFTP operations. Please see the [https://github.com/pkg/sftp](https://github.com/pkg/sftp) package.
- Support SCP. Please see the [https://github.com/kkirsche/gscp](https://github.com/kkirsche/gscp) package.
