package core

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/wisepythagoras/honeyshell/plugin"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"gorm.io/gorm"
)

// SSHServer This is the object that defines an SSH server.
type SSHServer struct {
	Logger        *Logman
	db            *gorm.DB
	Port          int
	Address       string
	Key           string
	Banner        string
	config        *ssh.ServerConfig
	listener      net.Listener
	PluginManager *plugin.PluginManager
}

// Init Initializes the SSH server.
func (server *SSHServer) Init() bool {
	server.config = &ssh.ServerConfig{
		PasswordCallback:  server.passwordChecker,
		PublicKeyCallback: server.publicKeyChecker,
		ServerVersion:     server.Banner,
		AuthLogCallback:   server.authLogHandler,
	}

	// Now read the server's private key.
	privateKeyBytes, err := os.ReadFile(server.Key)

	if err != nil {
		log.Println("Failed to load private key", err)
		server.Logger.Println("Failed to load private key", err)
		return false
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)

	if err != nil {
		log.Println("Failed to parse private key", err)
		server.Logger.Println("Failed to parse private key", err)
		return false
	}

	server.config.AddHostKey(privateKey)

	// Listen on the provided port.
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", server.Port))
	server.listener = listener

	if err != nil {
		log.Println("Unable to listen on the provided port", err)
		server.Logger.Println("Unable to listen on the provided port", err)
	}

	log.Println("Starting on port", server.Port)
	server.Logger.Println("Starting on port", server.Port)

	return true
}

// authLogHandler is meant to just display when a user connects.
func (server *SSHServer) authLogHandler(c ssh.ConnMetadata, method string, err error) {
	if method == "none" {
		ip := c.RemoteAddr()
		clientBanner := string(c.ClientVersion())

		log.Println(ip.String(), "client connected with", clientBanner, "for user", c.User())
		server.Logger.Println(ip.String(), "client connected with", clientBanner, "for user", c.User())
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

	server.Logger.Printf("%s %s pass:%s\n", ip.String(), username, password)
	log.Printf("%s %s pass:%s\n", ip.String(), username, password)

	// If a plugin manager was passed in and initialized, then call all of the plugins that offer a password login
	// interception function.
	if server.PluginManager != nil {
		for _, pl := range server.PluginManager.GetPasswordIntercepts() {
			if shouldLogin := pl.CallPasswordInterceptor(username, password, &ipObj); shouldLogin {
				return nil, nil
			}
		}
	}

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

	server.Logger.Printf("%s %s key:%s (%s)\n",
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

// HandleSSHAuth Handles the authentication process, as well as any individual session.
func (server *SSHServer) HandleSSHAuth(connection *net.Conn) bool {
	conn, chans, reqs, err := ssh.NewServerConn(*connection, server.config)

	if err != nil {
		log.Println("Error during handshake", err)
		server.Logger.Println("Error during handshake", err)
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

		go func(in <-chan *ssh.Request) {
			for req := range in {
				if req.Type == "shell" || req.Type == "pty" {
					req.Reply(true, nil)
				}
			}
		}(requests)

		user := &plugin.User{
			Username: conn.User(),
			Group:    conn.User(),
		}

		sessionVFS := *server.PluginManager.PluginVFS
		sessionVFS.User = user

		sessionTerm := term.NewTerminal(channel, "$ ")
		session := &plugin.Session{
			VFS:     &sessionVFS,
			Term:    sessionTerm,
			Manager: server.PluginManager,
			User:    user,
		}
		sessionTerm.AutoCompleteCallback = session.AutoCompleteCallback

		// Change over to the home directory so that the session starts from there.
		session.Chdir(server.PluginManager.PluginVFS.Home)

		if server.PluginManager.LoginMessageFn != nil {
			loginMessage := server.PluginManager.LoginMessageFn(session)
			session.TermWrite(loginMessage)
		}

		// Set the initial prompt.
		sessionTerm.SetPrompt(server.PluginManager.PromptPlugin(session))

		go func() {
			defer channel.Close()

			for {
				line, err := sessionTerm.ReadLine()

				if err != nil {
					break
				}

				if strings.Trim(line, " ") == "" {
					continue
				}

				parts := strings.SplitN(line, " ", 2)
				cmd := parts[0]
				args := &plugin.CmdArgs{}

				if strings.HasPrefix(cmd, ".") {
					pwd := session.GetPWD()
					cmd = filepath.Join(pwd, cmd)
				}

				if len(parts) > 1 {
					args.RawArgs = parts[1]
					args.Parse()
				}

				if commandFn, ok := server.PluginManager.GetCommand(cmd); ok {
					commandFn(args, session)
				} else {
					out := ""

					if strings.HasPrefix(cmd, "/") {
						// TODO: In this case we want to handle what happens if a directory is found:
						//     bash: /dir/here: Is a directory
						// Or if a file is not executable:
						//     bash: ./path/to/file: Permission denied
						out = fmt.Sprintf("%s: No such file or directory\n", line)
					} else {
						out = fmt.Sprintf("%s: command not found\n", line)
					}

					sessionTerm.Write([]byte(out))
					server.Logger.Println("[client] $", line)
					log.Println("[client] $", line)
				}

				sessionTerm.SetPrompt(server.PluginManager.PromptPlugin(session))
			}
		}()
	}

	return true
}

// ListenLoop Run the listener for our server.
func (server *SSHServer) ListenLoop() {
	// Now, this is the main loop where all the connections should be captured.
	for {
		connection, err := server.listener.Accept()

		if err != nil {
			log.Println("Error on accepting connection", err)
			server.Logger.Println("Error on accepting connection", err)
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
