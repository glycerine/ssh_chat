package ssh

import (
	"io/ioutil"
	"log"

	"code.google.com/p/go.crypto/ssh"
)

func Config() *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn *ssh.ServerConn, user, pass string) bool {
			return user == "test" && pass == "test123"
		},
	}
	config.NoClientAuth = true
	pemBytes, err := ioutil.ReadFile("privkey")
	if err != nil {
		log.Fatal("Failed to load private key file: ", err)
	}
	if err = config.SetRSAPrivateKey(pemBytes); err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}
	return config
}

func Open(config *ssh.ServerConfig) *ssh.Listener {
	conn, err := ssh.Listen("tcp", ":8765", config)
	if err != nil {
		log.Fatal("Failed to listen for connection")
	}
	return conn
}
