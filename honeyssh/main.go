package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
)

var (
	statsd     = kingpin.Flag("statsd", "URL to HoneySSH statds").Default("http://localhost:4000/attempt").String()
	alwaysDeny = kingpin.Flag("always-deny", "Allways deny SSH login attempts").Default("true").Bool()
	hostKey    = kingpin.Flag("key", "Path to SSH host key to use").Default("./id_rsa").String()
	listen     = kingpin.Flag("listen", "Listen address in host:port format").Default("localhost:4022").String()
)

func main() {
	kingpin.Parse()

	ch := make(chan Attempt, 100)

	go func() {
		for msg := range ch {
			if *statsd != "" {
				blob, err := json.Marshal(msg)
				if err != nil {
					logrus.Errorf("failed to marshal attempt: %s", err)
					continue
				}
				http.Post(*statsd, "application/json", bytes.NewBuffer(blob))
			}
		}
	}()

	logrus.Infof("Starting HoneySSH ....")
	sshServer(ch)
}
