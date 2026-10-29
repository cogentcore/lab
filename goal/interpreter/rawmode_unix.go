// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos

package interpreter

import "golang.org/x/sys/unix"

// makeRaw puts the given terminal into raw mode, as the line reader does
// for the terminal of the process, returning a function that restores it.
// Unlike golang.org/x/term.MakeRaw, it keeps the output processing (OPOST),
// as the line reader ends lines with a newline only.
func makeRaw(fd int) (func() error, error) {
	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return nil, err
	}
	old := *termios
	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, termios); err != nil {
		return nil, err
	}
	return func() error {
		return unix.IoctlSetTermios(fd, ioctlWriteTermios, &old)
	}, nil
}
