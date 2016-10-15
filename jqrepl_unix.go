// +build darwin dragonfly freebsd linux,!appengine netbsd openbsd

package jqrepl

import "os"

// ReopenTTY will open a new filehandle to the controlling terminal.
//
// Used when stdin has been redirected so that we can still get an interactive
// prompt to the user.
func ReopenTTY() (*os.File, error) {
	return os.Open("/dev/tty")
}
