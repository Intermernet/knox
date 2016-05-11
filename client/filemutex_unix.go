// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd

package client

import (
	"sync"
	"syscall"
)

const mkdirPerm = 0750

// FileMutex is similar to sync.RWMutex, but also synchronizes across processes.
// This implementation is based on flock syscall.
type FileMutex struct {
	mu sync.RWMutex
	fd int
}

func MakeFileMutex(filename string) (*FileMutex, error) {
	if filename == "" {
		return &FileMutex{fd: -1}, nil
	}
	fd, err := syscall.Open(filename, syscall.O_CREAT|syscall.O_RDONLY, mkdirPerm)
	if err != nil {
		return nil, err
	}
	return &FileMutex{fd: fd}, nil
}

func (m *FileMutex) Lock() error {
	m.mu.Lock()
	if m.fd != -1 {
		if err := syscall.Flock(m.fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			return err
		}
	}
	return nil
}

func (m *FileMutex) Unlock() error {
	if m.fd != -1 {
		if err := syscall.Flock(m.fd, syscall.LOCK_UN|syscall.LOCK_NB); err != nil {
			return err
		}
	}
	m.mu.Unlock()
	return nil
}
