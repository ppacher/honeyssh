package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"syscall"
	"unsafe"

	"github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func sshServer(ch chan Attempt) {
	config := ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			logrus.Infof("Logon attempt: host=%s version=%s user=%q pass=%q", c.RemoteAddr(), c.ClientVersion(), c.User(), string(pass))

			source, _, err := net.SplitHostPort(c.RemoteAddr().String())
			if err != nil {
				source = c.RemoteAddr().String()
			}

			ch <- Attempt{
				User:        c.User(),
				Password:    string(pass),
				Source:      source,
				Version:     string(c.ClientVersion()),
				Application: "ssh",
			}

			if *alwaysDeny {
				return nil, fmt.Errorf("rejected")
			}

			return nil, fmt.Errorf("rejected")
		},
	}

	privateBytes, err := ioutil.ReadFile(*hostKey)
	if err != nil {
		logrus.Fatal(err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		logrus.Fatal(err)
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Listening on %s", *listen)
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			logrus.Error(err)
			continue
		}

		sshConn, _, _, err := ssh.NewServerConn(tcpConn, &config)
		if err != nil {
			logrus.Error(err)
		}

		if sshConn != nil {
			sshConn.Close()
		}
		tcpConn.Close()

		/*
			go ssh.DiscardRequests(reqs)
			go handleChannels(chans)
		*/
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return

	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return

	}
	/*
		// Fire up bash for this session
		bash := exec.Command("bash")

		// Prepare teardown function
		close := func() {
			connection.Close()
			_, err := bash.Process.Wait()
			if err != nil {
				log.Printf("Failed to exit bash (%s)", err)

			}
			log.Printf("Session closed")

		}

		// Allocate a terminal for this channel
		log.Print("Creating pty...")
		bashf, err := pty.Start(bash)
		if err != nil {
			log.Printf("Could not start pty (%s)", err)
			close()
			return

		}
	*/
	//pipe session to bash and visa-versa
	//var once sync.Once
	go func() {
		//io.Copy(connection, bashf)
		//once.Do(close)

		scanner := bufio.NewScanner(connection)
		for scanner.Scan() {
			logrus.Infof("> %s", scanner.Text())
		}

		connection.Close()
	}()
	go func() {
		//io.Copy(bashf, connection)
		//once.Do(close)
	}()

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				// We only accept the default shell
				// (i.e. no command in the Payload)
				if len(req.Payload) == 0 {
					req.Reply(true, nil)

				}
				/*
					case "pty-req":
						termLen := req.Payload[3]
						w, h := parseDims(req.Payload[termLen+4:])
						SetWinsize(bashf.Fd(), w, h)
						// Responding true (OK) here will let the client
						// know we have a pty ready for input
						req.Reply(true, nil)
					case "window-change":
						w, h := parseDims(req.Payload)
						SetWinsize(bashf.Fd(), w, h)
				*/
			}

		}

	}()
}

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h

}

// ======================

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused

}

// SetWinsize sets the size of the given pty.
func SetWinsize(fd uintptr, w, h uint32) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))

}
