package main

/*
#include "./honeyshell.h"
*/
import "C"

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"gorm.io/gorm"
)

// SSHServer : This is the object that defines an SSH server.
type SSHServer struct {
	db      *gorm.DB
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

	// If the database object is passed to the constructor / intializer of the SSHServer, Go returns
	// the following error. Figure it out.
	//
	// panic: runtime error: cgo argument has Go pointer to Go pointer
	// goroutine 1 [running]:
	// main.(*SSHServer).Init.func2(0xc000243180, 0xc000243180, 0x0)
	// 	/home/diogenis/Projects/honeyshell/ssh_server.go:84 +0x6e
	// main.(*SSHServer).Init(0xc000243180, 0xc000243180)
	// 	/home/diogenis/Projects/honeyshell/ssh_server.go:84 +0xe5
	// main.main()
	// 	/home/diogenis/Projects/honeyshell/main.go:55 +0x345

	log.Println("Starting on port", server.port)
	logman.Println("Starting on port", server.port)

	// Try to start listening on the given port.
	if int(C.ssh_bind_listen(server.sshbind)) < 0 {
		log.Println("Error binding:", C.GoString(C.ssh_get_error(unsafe.Pointer(server.sshbind))))
		logman.Println("Error binding:", C.GoString(C.ssh_get_error(unsafe.Pointer(server.sshbind))))
		return false
	}

	return true
}

// SetDB sets the database. For some reason this can't be passed in the constructor / initializer,
// because this error happens: "panic: runtime error: cgo argument has Go pointer to Go pointer".
func (server *SSHServer) SetDB(db *gorm.DB) {
	server.db = db
}

// HandleSSHAuth : Handles the authentication process.
func (server *SSHServer) HandleSSHAuth(session *C.ssh_session) bool {
	// Get the IP of the client that's connected.
	ip, port, _ := SockAddrToIP(GetSSHSockaddr(*session))
	authMethods := (C.int)(C.SSH_AUTH_METHOD_PASSWORD | C.SSH_AUTH_METHOD_PUBLICKEY)

	// Get the status of the connection and report if it's closed.
	if C.ssh_get_status(*session) != 0 {
		err := C.GoString(C.ssh_get_error(unsafe.Pointer(*session)))
		logman.Println(ip.String() + " closed status: " + err)
		log.Println(ip.String() + " closed status: " + err)
	}

	log.Println(ip.String(), port, "connection request")
	logman.Println(ip.String(), port, "connection request")

	password_queue := C.create_password_queue()
	fmt.Println(password_queue)

	go func() {
		for {
			// if C.is_password_queue_empty(&password_queue) == 1 {
			// 	continue
			// }

			// if msg := C.get_password_msg(&password_queue); msg != nil {
			// 	fmt.Println("Get message") //fmt.Println("->", msg)
			// }

			msg := C.wait_for_password(&password_queue)

			if msg != nil {
				fmt.Println("->", msg)
			}
		}
	}()

	// Set how we want to allow peers to connect.
	C.ssh_set_auth_methods(*session, authMethods)
	C.handle_auth(*session, &password_queue)

	// Handle the key exchange.
	if C.ssh_handle_key_exchange(*session) != C.SSH_OK {
		// If there was an error, report it.
		sshErr := C.GoString(C.ssh_get_error(unsafe.Pointer(*session)))
		err := ip.String() + " Error exchanging keys " + sshErr

		logman.Println(err)
		log.Println(err)

		return false
	}

	// Get the client's SSH banner. This can give us some useful information as to what
	// software the attacker is running.
	clientBanner := C.GoString(C.ssh_get_clientbanner(*session))
	log.Println(ip.String(), port, "client connected with", clientBanner)
	logman.Println(ip.String(), port, "client connected with", clientBanner)

	for {
		break
		// Receive the message from the client.
		message := C.ssh_message_get(*session)

		if message == nil {
			break
		}

		messageType := C.ssh_message_subtype(message)
		username := C.GoString(C.ssh_message_auth_user(message))

		// If the attacker is submitting an authentication message, then we need to read
		// it and output the data that they entered.
		if messageType == C.SSH_AUTH_METHOD_PASSWORD {
			// password := C.GoString(C.ssh_message_auth_password(message))

			// // Add the password to the database.
			// server.db.Create(&PasswordConnection{
			// 	IPAddress: ip.String(),
			// 	Username:  username,
			// 	Password:  password,
			// })

			// logman.Printf("%s %s pass:%s\n", ip.String(), username, password)
			// log.Printf("%s %s pass:%s\n", ip.String(), username, password)
		} else if messageType == C.SSH_AUTH_METHOD_PUBLICKEY {
			// Get the user's auth key. This may not be that helpful, but I think it may
			// be interesting to capture some keys from those who are not careful.
			authKey := C.ssh_message_auth_pubkey(message)

			// Get the key type (ssh-rsa, etc).
			keyType := C.GoString(C.get_ssh_key_type(authKey))

			// This will hold the public key in base64.
			var pubKey *C.char

			// Now get the public key blob.
			C.ssh_pki_export_pubkey_base64(authKey, &pubKey)

			// Now we get the SHA3-256 hash of the public key, because it may be too long
			// to display on the screen.
			pubKeyHash, err := GetSHA3256Hash(C.GoBytes(unsafe.Pointer(pubKey),
				C.int(len(C.GoString(pubKey)))))

			// Add the key to the database.
			server.db.Create(&KeyConnection{
				IPAddress: ip.String(),
				Username:  username,
				Key:       C.GoString(pubKey),
				KeyHash:   ByteArrayToHex(pubKeyHash),
			})

			if err == nil {
				logman.Printf("%s %s key:%s (%s)\n",
					ip.String(),
					username,
					ByteArrayToHex(pubKeyHash),
					keyType)
				log.Printf("%s %s key:%s (%s)\n",
					ip.String(),
					username,
					ByteArrayToHex(pubKeyHash),
					keyType)
			}
		} else {
			C.ssh_message_auth_set_methods(message, authMethods)
			C.ssh_message_reply_default(message)
			continue
		}

		// Reply with the default message and clear the pointer.
		C.ssh_message_auth_set_methods(message, authMethods)
		C.ssh_message_reply_default(message)
		C.ssh_message_free(message)
	}

	log.Println(ip.String(), "connection terminated")
	logman.Println(ip.String(), "connection terminated")

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
			log.Println("Error accepting", msg)
			logman.Println("Error accepting", msg)
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
//
//	(https://golang.org/pkg/syscall/#Sockaddr).
func GetSSHSockaddr(session C.ssh_session) *syscall.Sockaddr {
	sockFd := int(C.ssh_get_fd(session))
	sock, err := syscall.Getpeername(sockFd)

	if err != nil {
		log.Println(err.Error())
		logman.Println(err.Error())
		return nil
	}

	return &sock
}
