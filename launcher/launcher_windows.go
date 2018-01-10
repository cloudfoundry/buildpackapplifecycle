package main

import (
	"fmt"
	"math"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32       = windows.NewLazySystemDLL("kernel32.dll")
	createProcessW = kernel32.NewProc("CreateProcessW")
)

func runProcess(dir, command string) {
	err := createProcessW.Find()
	handleErr("couldn't find func address", err)

	args, err := syscall.UTF16PtrFromString(command)
	handleErr("casting command failed", err)
	cwd, err := syscall.UTF16PtrFromString(dir)
	handleErr("casting cwd failed", err)

	p, _ := syscall.GetCurrentProcess()
	fd := make([]syscall.Handle, 3)
	for i, file := range []*os.File{os.Stdin, os.Stdout, os.Stderr} {
		err := syscall.DuplicateHandle(p, syscall.Handle(file.Fd()), p, &fd[i], 0, true, syscall.DUPLICATE_SAME_ACCESS)
		if err != nil {
			handleErr("DuplicateHandle failed", err)
		}
		defer syscall.CloseHandle(syscall.Handle(fd[i]))
	}
	si := new(syscall.StartupInfo)
	si.Cb = uint32(unsafe.Sizeof(*si))
	si.Flags = syscall.STARTF_USESTDHANDLES
	si.StdInput = fd[0]
	si.StdOutput = fd[1]
	si.StdErr = fd[2]
	pi := new(syscall.ProcessInformation)

	// Change the parent's working directory to the app dir so
	// CreateProcessW will search it when starting the child process
	err = os.Chdir(dir)
	handleErr("couldn't change working directory", err)

	// CreateProcessW docs
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms682425(v=vs.85).aspx

	// Process Creation flags
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
	r, _, e := syscall.Syscall12(createProcessW.Addr(), 10,
		uintptr(uint16(0)),            // appname
		uintptr(unsafe.Pointer(args)), // executable and args
		uintptr(unsafe.Pointer(nil)),  // process security attributes
		uintptr(unsafe.Pointer(nil)),  // thread security attributes
		uintptr(uint32(1)),            // inherit parent's handles
		uintptr(uint32(0)),            // creation flags
		uintptr(unsafe.Pointer(nil)),  // inherit parent's environment
		uintptr(unsafe.Pointer(cwd)),  // process working directory
		uintptr(unsafe.Pointer(si)),   // startup info
		uintptr(unsafe.Pointer(pi)),   // process info for the created process
		0, 0)

	if r == 0 {
		handleErr(fmt.Sprintf("CreateProcessW failed %s:%s", dir, command), e)
	}
	defer syscall.CloseHandle(syscall.Handle(pi.Thread))
	defer syscall.CloseHandle(syscall.Handle(pi.Process))

	_, err = syscall.WaitForSingleObject(pi.Process, math.MaxUint32)
	handleErr("WaitForSingleObject failed", err)

	var exitCode uint32
	err = syscall.GetExitCodeProcess(pi.Process, &exitCode)
	handleErr("GetExitCodeProcess failed", err)

	os.Exit(int(exitCode))
}

func handleErr(description string, err error) {
	if err != nil {
		fmt.Printf("%s: %s", description, err.Error())
		os.Exit(1)
	}
}
