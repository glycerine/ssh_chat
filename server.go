package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"

	sshServer "github.com/kdorland/ssh_chat/ssh"
	"github.com/kdorland/ssh_chat/text"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
)

var (
	path = ""
)

type Client struct {
	conn     net.Conn
	ch       chan text.Message
	username string
	term     *terminal.Terminal
}

type Socket interface {
	Write(p []byte) (n int, err error)
	ReadLine() (line string, err error)
}

func installQuitHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println(text.Yellow("Voyage ended"))
		os.Exit(1)
	}()
}

func main() {
	installQuitHandler()
	path, _ = filepath.Abs(".")

	// ssh
	config := sshServer.Config()
	conn := sshServer.Open(config)

	// Channel for messages that are passed around
	msgchan := make(chan text.Message)
	// Channel for clients that are added
	addchan := make(chan Client)
	// Channel for connections that are removed
	rmchan := make(chan net.Conn)

	go handleMessages(msgchan, addchan, rmchan)

	fmt.Println(text.Clear)
	fmt.Println(text.Yellow("Voyage started"))

	for {
		log.Println("Accepting incoming connections")
		sConn, err := conn.Accept()
		if err != nil {
			log.Println("Failed to accept incoming connection")
			continue
		}
		if err := sConn.Handshake(); err != nil {
			log.Println("Failed to handshake")
			continue
		}
		go handleSshConnection(sConn, msgchan, addchan, rmchan)
	}
}

func handleMessages(msgchan <-chan text.Message, addchan <-chan Client,
	rmchan <-chan net.Conn) {

	clients := make(map[net.Conn]chan<- text.Message)

	for {
		select {
		case msg := <-msgchan:
			log.Printf("New message: %s\n", msg)
			for _, ch := range clients {
				go func(mch chan<- text.Message) {
					mch <- msg
				}(ch)
			}
		case client := <-addchan:
			log.Printf("New client: %v\n", client.conn.RemoteAddr())
			clients[client.conn] = client.ch
			for conn, _ := range clients {
				go func(conn net.Conn) {
					conn.Write([]byte("")) // Hack to detect disconnect, empty packet write
				}(conn)
			}

		case conn := <-rmchan:
			log.Printf("Client disconnects: %v\n", conn.RemoteAddr())
			delete(clients, conn)
		}
	}
}

func handleSshConnection(sConn *ssh.ServerConn,
	msgchan chan<- text.Message, addchan chan<- Client, rmchan chan<- net.Conn) {

	defer sConn.Close()
	for {
		ch, err := sConn.Accept()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("handleServerConn Accept:", err)
			break
		}
		if ch.ChannelType() != "session" {
			ch.Reject(ssh.UnknownChannelType, "unknown channel type")
			break
		}

		log.Println("Client version:", string(sConn.ClientVersion))

		// Create terminal
		term := terminal.NewTerminal(ch, "")
		serverTerm := &ssh.ServerTerminal{
			Term:    term,
			Channel: ch,
		}
		ch.Accept()

		go handleConnection(sConn, serverTerm, term, msgchan, addchan, rmchan)
	}
}

func handleConnection(conn net.Conn, socket Socket, term *terminal.Terminal,
	msgchan chan<- text.Message, addchan chan<- Client, rmchan chan<- net.Conn) {

	// Setup channels
	toClient := make(chan text.Message)
	fromClient := make(chan text.Message)
	client := Client{conn, toClient, "", term}
	addchan <- client

	// Clean up stuff
	defer func() {
		rmchan <- conn
		conn.Close()
	}()

	go handleClientRead(socket, &client, fromClient)

	handleClient(socket, &client, fromClient, msgchan)
}

func handleClientRead(socket Socket, client *Client, fromClient chan<- text.Message) {
	defer close(fromClient)
	Writeln(socket, "Welcome to the Chat Server!\n")

	nickname := ""
	for ok := false; !ok; {
		nickname, ok = getNickname(socket)
	}
	client.username = nickname
	client.term.SetPrompt("say: ")

	t := text.LightGreen("New user ") + text.Yellow(nickname) +
		text.LightGreen(" has joined the chat.")
	fromClient <- text.Message{Msg: t, Sender: nickname}

	for {
		line, err := socket.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("handleClient readLine err:", err)
			continue
		}
        // TODO: Needs to be blocking so response can be synchronized
		fromClient <- text.Message{
			Msg:     string(line),
			Sender:  nickname,
			MsgType: "chat",
		}
	}

	t = text.LightGreen("User ") + text.Yellow(nickname) +
		text.LightGreen(" left the chat.")
	fromClient <- text.Message{Msg: t, Sender: nickname}
}

func handleClient(socket Socket, client *Client,
	fromClient chan text.Message, msgchan chan<- text.Message) {

	for {
		select {
		case msg, ok := <-fromClient: // from client socket
			if !ok {
				return
			}
			msgchan <- msg
		case msg := <-client.ch: // write to client socket

			m := text.FormatChatMsg(msg, client.username)
			_, err := Writeln(socket, m)
			if err != nil {
				return
			}
		}
	}
}

func getNickname(socket Socket) (string, bool) {
	Write(socket, "What is your nick? ")
	nick, err := socket.ReadLine()
	if err != nil {
		return "", false
	}

	nickname := string(nick)
	if nickname == "" {
		return "", false
	}

	Writeln(socket, "Welcome, "+text.Yellow(nickname)+"!")
	return nickname, true
}

func Write(stream io.Writer, msg string) (int, error) {
	return stream.Write([]byte(msg))
}

func Writeln(stream io.Writer, msg string) (int, error) {
	return stream.Write([]byte(msg + "\r\n"))
}
