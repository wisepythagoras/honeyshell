package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/wisepythagoras/honeyshell/core"
	"github.com/wisepythagoras/honeyshell/plugin"
)

var logman *core.Logman

// https://github.com/karfield/ssh2go/
// http://api.libssh.org/master/group__libssh__session.html
// https://github.com/linuxdeepin/go-lib/blob/master/users/passwd/passwd.go

func main() {
	// Define the command line flags and their default values.
	username := flag.String("user", "", "Set the permissions to a certain user (ie 'nobody')")
	port := flag.Int("port", 22, "The port the deamon should run on")
	banner := flag.String("banner", "SSH-2.0-OpenSSH_7.4p1 Raspbian-10+deb9u3", "The banner for the SSH server")
	key := flag.String("key", "", "The RSA key to use")
	pluginsFolder := flag.String("plugins", "", "The path to the folder containing the plugins")
	vfsPath := flag.String("vfs", "", "The path to the VFS (virtual file system) JSON file")
	verbose := flag.Bool("verbose", false, "Print out debug messages")

	// Parse the command line arguments (flags).
	flag.Parse()

	// Require an RSA key.
	if *key == "" {
		log.Fatalln("An RSA key is required. Use the '-key' flag")
	}

	// Validate the port.
	if *port < 1 || *port > 65535 {
		log.Fatalf("Invalid port number %d\n", *port)
	}

	var pluginManager *plugin.PluginManager
	var vfs *plugin.VFS

	// Start the logman.
	logman = core.GetLogmanInstance()
	logman.Println("Starting Honeyshell")

	// Connect to the database.
	db, err := core.ConnectDB(*verbose)

	if err != nil {
		log.Fatalln(err)
	}

	if len(*vfsPath) > 0 {
		vfs, err = plugin.ReadVFSJSONFile(*vfsPath)

		if err != nil {
			log.Fatalln("Error:", err)
		}
	}

	if len(*pluginsFolder) > 0 {
		pluginManager = &plugin.PluginManager{
			DB:        db,
			PluginVFS: vfs,
		}

		if err := pluginManager.LoadPlugins(*pluginsFolder); err != nil {
			log.Fatalln("Error:", err)
		}
	}

	// Create a new SSH server object.
	sshServer := &core.SSHServer{
		Port:          *port,
		Address:       "0.0.0.0",
		Key:           *key,
		Banner:        *banner,
		PluginManager: pluginManager,
		Logger:        logman,
	}

	// Initialize the SSH server.
	if !sshServer.Init() {
		log.Fatalln("Unable to start server")
	}

	// Set the database onto the sshServer instance.
	sshServer.SetDB(db)

	// Now, if there was a username passed from the command line arguments, try to switch
	// all of the permissions to that user.
	if *username != "" {
		log.Printf("Changing permissions to user '%s'\n", *username)

		gid, uid, err := core.ChangePermissions(*username)

		// If this fails it means that either the program wasn't run with 'sudo.' or the user
		// doesn't have sufficient permissions.
		if err != nil {
			errStr := fmt.Sprintf("Error: %s", err)
			log.Println(errStr)
			logman.Println(errStr)
			return
		} else {
			log.Println("User set to", *username)
			logman.Println("User set to", *username)
			log.Println(*username, "has gid:", gid, "and uid:", uid)
			logman.Println(*username, "has gid:", gid, "and uid:", uid)
		}
	}

	// Run the loop that listens for new connections.
	sshServer.ListenLoop()

	// Close the SSH server.
	sshServer.Stop()

	logman.Println("Terminating process")
}
