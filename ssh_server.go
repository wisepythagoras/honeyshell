package main

// #include <stdlib.h>
// #include <pwd.h>
// #include <sys/types.h>
// #include <grp.h>
// #include <inttypes.h>
// #include <sys/types.h>
// #include <libssh/libssh.h>
// #include <libssh/server.h>
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

// SSHServer : This is the object that defines an SSH server.
type SSHServer struct {
	port    int
	address string
	key     string
	banner  string
	sshbind C.ssh_bind
}

// Init : Initialize the SSH server.
func (server *SSHServer) Init() bool {
	// Start the SSH server here.
	C.ssh_init()

	// Create a new SSH binder.
	server.sshbind = C.ssh_bind_new()

	// Convert some settings to unsafe pointers so that they can be used within our
	// C functions.
	address := unsafe.Pointer(C.CString(server.address))
	portC := unsafe.Pointer(&server.port)
	banner := unsafe.Pointer(C.CString(server.banner))
	rsaKey := unsafe.Pointer(C.CString(server.key))

	// Set the bind options.
	C.ssh_bind_options_set(server.sshbind, C.SSH_BIND_OPTIONS_BINDADDR, address)
	C.ssh_bind_options_set(server.sshbind, C.SSH_BIND_OPTIONS_BINDPORT, portC)
	C.ssh_bind_options_set(server.sshbind, C.SSH_BIND_OPTIONS_BANNER, banner)
	C.ssh_bind_options_set(server.sshbind, C.SSH_BIND_OPTIONS_RSAKEY, rsaKey)

	fmt.Println("Starting on port", server.port)
	logger.Println("Starting on port", server.port)

	// Try to start listening on the given port.
	if int(C.ssh_bind_listen(server.sshbind)) < 0 {
		fmt.Println("Error binding:", C.GoString(C.ssh_get_error(unsafe.Pointer(server.sshbind))))
		logger.Println("Error binding:", C.GoString(C.ssh_get_error(unsafe.Pointer(server.sshbind))))
		return false
	}

	return true
}

// HandleSSHAuth : Handles the authentication process.
func (server *SSHServer) HandleSSHAuth(session *C.ssh_session) bool {
	// Get the IP of the client that's connected.
	ip, port, _ := SockAddrToIP(GetSSHSockaddr(*session))

	// Get the status of the connection and report if it's closed.
	if C.ssh_get_status(*session) != 0 {
		err := C.GoString(C.ssh_get_error(unsafe.Pointer(*session)))
		logger.Println(ip.String() + " closed status: " + err)
		fmt.Println(ip.String() + " closed status: " + err)
	}

	fmt.Println(ip.String(), port, "connection request")
	logger.Println(ip.String(), port, "connection request")

	// Set how we want to allow peers to connect.
	C.ssh_set_auth_methods(*session, C.SSH_AUTH_METHOD_PASSWORD | C.SSH_AUTH_METHOD_PUBLICKEY);

	// Handle the key exchange.
	if C.ssh_handle_key_exchange(*session) != C.SSH_OK {
		// If there was an error, report it.
		sshErr := C.GoString(C.ssh_get_error(unsafe.Pointer(*session)))
		err := ip.String() + " Error exchanging keys " + sshErr

		logger.Println(err)
		fmt.Println(err)

		return false
	}

	// Get the client's SSH banner. This can give us some useful information as to what
	// software the attacker is running.
	clientBanner := C.GoString(C.ssh_get_clientbanner(*session))
	fmt.Println(ip.String(), port, "client connected with", clientBanner)
	logger.Println(ip.String(), port, "client connected with", clientBanner)

	for {
		// Receive the message from the client.
		message := C.ssh_message_get(*session)

		if message == nil {
			break
		}

		messageType := C.ssh_message_subtype(message)

		// If the attacker is submitting an authentication message, then we need to read
		// it and output the data that they entered.
		if messageType == C.SSH_AUTH_METHOD_PASSWORD {
			logger.Printf("%s %s %s\n",
				ip.String(),
				C.GoString(C.ssh_message_auth_user(message)),
				C.GoString(C.ssh_message_auth_password(message)))
			fmt.Printf("%s %s %s\n",
				ip.String(),
				C.GoString(C.ssh_message_auth_user(message)),
				C.GoString(C.ssh_message_auth_password(message)))
		} else if messageType == C.SSH_AUTH_METHOD_PUBLICKEY {
		    // pubKey := C.ssh_message_auth_pubkey(message)
		}

		// Reply with the default message and clear the pointer.
		C.ssh_message_reply_default(message)
		C.ssh_message_free(message)
	}

	fmt.Println(ip.String(), "connection terminated")
	logger.Println(ip.String(), "connection terminated")

	return true
}

// ListenLoop : Run the listener for our server.
func (server *SSHServer) ListenLoop() {
	// Now, this is the main loop where all the connections should be captured.
	for {
		// Create a new SSH Session manager for the new connection.
		session := C.ssh_new()

		// Try to accept the connection.
		if C.ssh_bind_accept(server.sshbind, session) == C.SSH_ERROR {
			msg := C.GoString(C.ssh_get_error(unsafe.Pointer(server.sshbind)))
			fmt.Println("Error accepting", msg)
			logger.Println("Error accepting", msg)
			continue
		}

		// Handle authentication in a goroutine so that the loop is freed up for a possible
		// concurrent connection.
		go func() {
			server.HandleSSHAuth(&session)
		}()
	}
}

// Stop : Stop the SSH server from running.
func (server *SSHServer) Stop() {
	C.ssh_finalize()
}

// GetSSHSockaddr : Returns the socket address of an SSH client
//                  (https://golang.org/pkg/syscall/#Sockaddr).
func GetSSHSockaddr(session C.ssh_session) *syscall.Sockaddr {
	sockFd := int(C.ssh_get_fd(session))
	sock, err := syscall.Getpeername(sockFd)

	if err != nil {
		fmt.Println(err.Error())
		logger.Println(err.Error())
		return nil
	}

	return &sock
}

