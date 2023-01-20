package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/wisepythagoras/honeyshell/plugin"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"gorm.io/gorm"
)

// SSHServer : This is the object that defines an SSH server.
type SSHServer struct {
	db            *gorm.DB
	port          int
	address       string
	key           string
	banner        string
	config        *ssh.ServerConfig
	listener      net.Listener
	pluginManager *plugin.PluginManager
}

// Init Initialize the SSH server.
func (server *SSHServer) Init() bool {
	server.config = &ssh.ServerConfig{
		PasswordCallback:  server.passwordChecker,
		PublicKeyCallback: server.publicKeyChecker,
		ServerVersion:     server.banner,
		AuthLogCallback:   server.authLogHandler,
	}

	// Now read the server's private key.
	privateKeyBytes, err := os.ReadFile(server.key)

	if err != nil {
		log.Println("Failed to load private key", err)
		logman.Println("Failed to load private key", err)
		return false
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)

	if err != nil {
		log.Println("Failed to parse private key", err)
		logman.Println("Failed to parse private key", err)
		return false
	}

	server.config.AddHostKey(privateKey)

	// Listen on the provided port.
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", server.port))
	server.listener = listener

	if err != nil {
		log.Println("Unable to listen on the provided port", err)
		logman.Println("Unable to listen on the provided port", err)
	}

	log.Println("Starting on port", server.port)
	logman.Println("Starting on port", server.port)

	return true
}

// authLogHandler is meant to just display when a user connects.
func (server *SSHServer) authLogHandler(c ssh.ConnMetadata, method string, err error) {
	if method == "none" {
		ip := c.RemoteAddr()
		clientBanner := string(c.ClientVersion())

		log.Println(ip.String(), "client connected with", clientBanner, "for user", c.User())
		logman.Println(ip.String(), "client connected with", clientBanner, "for user", c.User())
	}
}

// passwordChecker will take the connection metadata and password that was used and log it along with
// other needed information to both the log file/stdout and database.
func (server *SSHServer) passwordChecker(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	ip := c.RemoteAddr()
	ipStr, _, _ := net.SplitHostPort(ip.String())
	ipObj := net.ParseIP(ipStr)
	username := c.User()
	password := string(pass)

	// Add the password to the database.
	server.db.Create(&PasswordConnection{
		IPAddress: ipStr,
		Username:  username,
		Password:  password,
	}).Commit()

	logman.Printf("%s %s pass:%s\n", ip.String(), username, password)
	log.Printf("%s %s pass:%s\n", ip.String(), username, password)

	// If a plugin manager was passed in and initialized, then call all of the plugins that offer a password login
	// interception function.
	if server.pluginManager != nil {
		for _, pl := range server.pluginManager.GetPasswordIntercepts() {
			if shouldLogin := pl.CallPasswordInterceptor(username, password, &ipObj); shouldLogin {
				return nil, nil
			}
		}
	}

	// This is where and how I'd ideally add logic to mock logins and trap bad bots.
	// if c.User() == "admin" && string(pass) == "admin" {
	// 	return nil, nil
	// }

	return nil, fmt.Errorf("incorrect password for %q", c.User())
}

// publicKeyChecker handles any attempt to send a public key. This could be especially helpful when monitoring
// the keys of your organization. If you see a strange IP using it, it's been compromised.
func (server *SSHServer) publicKeyChecker(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	ip := c.RemoteAddr()
	username := c.User()

	// Now we get the SHA3-256 hash of the public key, because it may be too long
	// to display on the screen.
	pubKeyHash, _ := GetSHA3256Hash(pubKey.Marshal())

	logman.Printf("%s %s key:%s (%s)\n",
		ip.String(),
		username,
		ByteArrayToHex(pubKeyHash),
		pubKey.Type())
	log.Printf("%s %s key:%s (%s)\n",
		ip.String(),
		username,
		ByteArrayToHex(pubKeyHash),
		pubKey.Type())

	// Add the key to the database.
	server.db.Create(&KeyConnection{
		IPAddress: ip.String(),
		Username:  username,
		Key:       string(pubKey.Marshal()),
		KeyHash:   ByteArrayToHex(pubKeyHash),
	}).Commit()

	return nil, fmt.Errorf("unknown public key for %q", c.User())
}

// SetDB sets the database. For some reason this can't be passed in the constructor / initializer,
// because this error happens: "panic: runtime error: cgo argument has Go pointer to Go pointer".
func (server *SSHServer) SetDB(db *gorm.DB) {
	server.db = db
}

// HandleSSHAuth Handles the authentication process.
func (server *SSHServer) HandleSSHAuth(connection *net.Conn) bool {
	conn, chans, reqs, err := ssh.NewServerConn(*connection, server.config)

	if err != nil {
		log.Println("Error during handshake", err)
		logman.Println("Error during handshake", err)
		return false
	}

	// At this point the user would be logged in.

	log.Printf("User %s logged in via %s\n", conn.User(), string(conn.ClientVersion()))

	// When I figure out what to do in this stage, this will have to be handled by something like the
	// code here: https://github.com/gogs/gogs/blob/main/internal/ssh/ssh.go
	// Ideally the honypot would log all payloads.
	go ssh.DiscardRequests(reqs)

	for c := range chans {
		if c.ChannelType() != "session" {
			c.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := c.Accept()

		if err != nil {
			log.Fatalf("Could not accept channel: %v", err)
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				if req.Type == "shell" || req.Type == "pty" {
					req.Reply(true, nil)
				}
			}
		}(requests)

		term := terminal.NewTerminal(channel, "$ ")

		go func() {
			defer channel.Close()
			for {
				line, err := term.ReadLine()

				if err != nil {
					break
				}

				if line == "pass" {
					pass, _ := term.ReadPassword("pass: ")
					fmt.Println(pass)
				} else {
					fmt.Println("->", line)
				}
			}
		}()
	}

	return true
}

// ListenLoop : Run the listener for our server.
func (server *SSHServer) ListenLoop() {
	// Now, this is the main loop where all the connections should be captured.
	for {
		connection, err := server.listener.Accept()

		if err != nil {
			log.Println("Error on accepting connection", err)
			logman.Println("Error on accepting connection", err)
		}

		// Handle authentication in a goroutine so that the loop is freed up for a possible
		// concurrent connection.
		go func() {
			server.HandleSSHAuth(&connection)
		}()
	}
}

// Stop will stop the SSH server from running.
func (server *SSHServer) Stop() {
	server.listener.Close()
}
