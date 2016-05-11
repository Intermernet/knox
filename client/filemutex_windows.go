// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package client

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

const (
	INVALID_FILE_HANDLE     = ^syscall.Handle(0)
	LOCKFILE_EXCLUSIVE_LOCK = 2
)

func lockFileEx(h syscall.Handle, flags, reserved, locklow, lockhigh uint32, ol *syscall.Overlapped) (err error) {
	r1, _, e1 := syscall.Syscall6(procLockFileEx.Addr(), 6, uintptr(h), uintptr(flags), uintptr(reserved), uintptr(locklow), uintptr(lockhigh), uintptr(unsafe.Pointer(ol)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func unlockFileEx(h syscall.Handle, reserved, locklow, lockhigh uint32, ol *syscall.Overlapped) (err error) {
	r1, _, e1 := syscall.Syscall6(procUnlockFileEx.Addr(), 5, uintptr(h), uintptr(reserved), uintptr(locklow), uintptr(lockhigh), uintptr(unsafe.Pointer(ol)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// FileMutex is similar to sync.RWMutex, but also synchronizes across processes.
// This implementation is based on flock syscall.
type FileMutex struct {
	mu sync.RWMutex
	fd syscall.Handle
}

func MakeFileMutex(filename string) (*FileMutex, error) {
	if filename == "" {
		return &FileMutex{fd: INVALID_FILE_HANDLE}, fmt.Errorf("invalid file handle")
	}
	fd, err := syscall.CreateFile(&(syscall.StringToUTF16(filename)[0]), syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, nil, syscall.OPEN_ALWAYS, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, err
	}
	return &FileMutex{fd: fd}, nil
}

func (m *FileMutex) Lock() error {
	m.mu.Lock()
	if m.fd != INVALID_FILE_HANDLE {
		var ol syscall.Overlapped
		if err := lockFileEx(m.fd, LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &ol); err != nil {
			return err
		}
	}
	return nil
}

func (m *FileMutex) Unlock() error {
	if m.fd != INVALID_FILE_HANDLE {
		var ol syscall.Overlapped
		if err := unlockFileEx(m.fd, 0, 1, 0, &ol); err != nil {
			return err
		}
	}
	m.mu.Unlock()
	return nil
}
