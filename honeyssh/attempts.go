package main

type Attempt struct {
	Application string `json:"app"`      // i.e. ssh, ftp, rdp
	User        string `json:"user"`     // i.e. root, admin, Administrator@example.com
	Password    string `json:"password"` // i.e. toor
	Version     string `json:"version"`  // i.e. SSH-2.0-PUTTY, ...
	Source      string `json:"source"`   // i.e. 1.2.3.4
}
