package main

// #include <stdlib.h>
// #include <pwd.h>
// #include <sys/types.h>
// #include <unistd.h>
import "C"

import (
	"errors"
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

// ChangePermissions : Change the permissions to a specific user.
func ChangePermissions(name string) bool {
	// Get the passwd before moving on.
	passwdC, err := GetPasswd(name)

	if passwdC == nil {
		fmt.Printf("%s\n", err.Error())
		return false
	} else {
		logman.Println(name, "has gid:", uint32(passwdC.pw_gid), "and uid:", uint32(passwdC.pw_uid))

		// Set the group id before anything else.
		if int(C.setgid(passwdC.pw_gid)) == -1 {
			errStr := "Unable to set group id"

			fmt.Println(errStr)
			logman.Println(errStr)

			return false
		}

		// Set the user id after the group id.
		if int(C.setuid(passwdC.pw_uid)) == -1 {
			errStr := "Unable to set user id"

			fmt.Printf(errStr)
			logman.Println(errStr)

			return false
		}
	}

	logman.Println("User set to", name)

	return true
}

// SockAddrToIP : Return the IP address of a sockaddr
func SockAddrToIP(sock *syscall.Sockaddr) (ip net.IP, port int, success bool) {
	switch sock := (*sock).(type) {
	case *syscall.SockaddrInet4:
		return net.IP((&sock.Addr)[:]), sock.Port, true
	case *syscall.SockaddrInet6:
		return net.IP((&sock.Addr)[:]), sock.Port, true
	}

	return
}

// GetPasswd : Gets the passwd of a specific user.
func GetPasswd(name string) (*C.struct_passwd, error) {
	// Read the name as a C string.
	nameC := C.CString(name)

	// Then defer the pointer and call getpwnam.
	defer C.free(unsafe.Pointer(nameC))
	passwdC, err := C.getpwnam(nameC)

	if passwdC == nil {
		if err == nil {
			return nil, errors.New("unable to load the user")
		}

		return nil, err
	}

	// Finally return the passwd.
	return passwdC, nil
}
