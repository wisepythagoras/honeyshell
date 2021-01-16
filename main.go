package main

import (
	"flag"
	"fmt"
)

var logger *Logman

// https://github.com/karfield/ssh2go/
// http://api.libssh.org/master/group__libssh__session.html
// https://github.com/linuxdeepin/go-lib/blob/master/users/passwd/passwd.go

func main() {
	// Define the command line flags and their default values.
	username := flag.String("user", "", "Set the permissions to a certain user (ie 'nobody')")
	port := flag.Int("port", 22, "The port the deamon should run on")
	banner := flag.String("banner", "OpenSSH_7.4p1 Raspbian-10+deb9u3", "The banner for the SSH server")
	key := flag.String("key", "", "The RSA key to use")

	// Parse the command line arguments (flags).
	flag.Parse()

	// Require an RSA key.
	if *key == "" {
		fmt.Println("An RSA key is required. Use the '-key' flag")
		return
	}

	// Validate the port.
	if *port < 1 || *port > 65535 {
		fmt.Printf("Invalid port number %d\n", *port)
		return
	}

	// Start the logger.
	logger = GetLogmanInstance()
	logger.Println("Starting Honeyshell")

	// Create a new SSH server object.
	sshServer := &SSHServer{
		port:    *port,
		address: "0.0.0.0",
		key:     *key,
		banner:  *banner,
	}

	// Initialize the SSH server.
	if !sshServer.Init() {
		fmt.Println("Unable to start server")
		return
	}

	// Now, if there was a username passed from the command line arguments, try to switch
	// all of the permissions to that user.
	if *username != "" {
		fmt.Printf("Changing permissions to user '%s'\n", *username)

		// If this fails it means that either the program wasn't run with 'sudo.' or the user
		// doesn't have sufficient permissions.
		if !ChangePermissions(*username) {
			err := "Either the user does not exist or you don't have adequate permissions"
			fmt.Println(err)
			logger.Println(err)
			return
		}
	}

	// Run the loop that listens for new connections.
	sshServer.ListenLoop()

	// Close the SSH server.
	sshServer.Stop()

	logger.Println("Terminating process")
}

