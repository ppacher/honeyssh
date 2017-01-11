package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/macaron.v1"
)

type Statistics struct {
	RecentIPs       map[string]int `json:"recent_ips"`
	RecentUsers     map[string]int `json:"recent_usernames"`
	RecentPasswords map[string]int `json:"recent_passwords"`
	RecentVersions  map[string]int `json:"recent_versions"`

	TopIPs       map[string]int `json:"top_ips"`
	TopUsers     map[string]int `json:"top_users"`
	TopPasswords map[string]int `json:"top_passwords"`
	TopVersions  map[string]int `json:"top_versions"`
}

type Attempt struct {
	Application string `json:"app"`      // i.e. ssh, ftp, rdp
	User        string `json:"user"`     // i.e. root, admin, Administrator@example.com
	Password    string `json:"password"` // i.e. toor
	Version     string `json:"version"`  // i.e. SSH-2.0-PUTTY, ...
	Source      string `json:"source"`   // i.e. 1.2.3.4
}

func main() {
	srv := macaron.Classic()

	ch := make(chan Attempt, 100)

	d, err := sql.Open("sqlite3", "./analysis.db")
	if err != nil {
		logrus.Fatal(err)
	}

	db := &db{db: d}

	srv.Group("", func() {
		srv.Post("/attempt", func(ctx *macaron.Context) {
			body, err := ctx.Req.Body().Bytes()
			if err != nil {
				logrus.Warnf("Cannot read body: %s", err)
				ctx.JSON(400, "Bad request")
				return
			}

			var attempt Attempt

			if err := json.Unmarshal(body, &attempt); err != nil {
				logrus.Warnf("Failed to parse body: %s", err)
				ctx.JSON(400, "Bad request")
				return
			}

			go func() {
				ch <- attempt
			}()
		})

		srv.Group("/stats", func() {
			srv.Get("", func(ctx *macaron.Context) {
				stats, err := db.GetStats()
				if err != nil {
					ctx.JSON(500, err.Error())
					return
				}

				ctx.JSON(200, stats)
			})
		})

	}, macaron.Renderer())

	go func(ch <-chan Attempt) {

		for test := range ch {
			if test.User == "" {
				logrus.Warn("No username for authentication attempt")
				continue
			}

			if test.Password == "" {
				logrus.Warn("No password for authentication attempt")
				continue
			}

			if test.Source == "" || test.Version == "" {
				logrus.Warn("No source/version for authentication attempt")
			}

			if err := db.handleAttempt(test); err != nil {
				logrus.Warnf("Failed to store logon-attempt: %s", err)
			}
		}
	}(ch)

	http.ListenAndServe("127.0.0.1:4000", srv)
}

type db struct {
	lock sync.Mutex
	db   *sql.DB
}

func (db *db) handleAttempt(a Attempt) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	source, err := db.insertSource(a.Source)
	if err != nil {
		return fmt.Errorf("source: %s", err)
	}
	user, err := db.insertUser(a.User)
	if err != nil {
		return fmt.Errorf("user: %s", err)
	}
	pass, err := db.insertPassword(a.Password)
	if err != nil {
		return fmt.Errorf("pass: %s", err)
	}
	app, err := db.insertApp(a.Application)
	if err != nil {
		return fmt.Errorf("app: %s", err)
	}
	version, err := db.insertVersion(a.Version)
	if err != nil {
		return fmt.Errorf("version: %s", err)
	}

	err = db.insertLogonAttempt(source, user, pass, version, app)
	if err != nil {
		return fmt.Errorf("attempt: %s", err)
	}

	return nil
}

func (db *db) insertSource(ip string) (int64, error) {
	var id int64 = -1

	err := db.db.QueryRow("SELECT id FROM sources WHERE ip=?", ip).Scan(&id)

	now := time.Now().Unix()

	if id == -1 {
		res, err := db.db.Exec("INSERT INTO sources(ip, first_seen, last_seen) VALUES(?, ?, ?)", ip, now, now)
		if err != nil {
			return 0, err
		}

		id, err = res.LastInsertId()
		return id, err
	}

	_, err = db.db.Exec("UPDATE sources SET last_seen=? WHERE id=?", now, id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (db *db) insertUser(user string) (int64, error) {

	var id int64 = -1

	err := db.db.QueryRow("SELECT user FROM users WHERE user=?", user).Scan(&id)

	now := time.Now().Unix()

	if id == -1 {
		res, err := db.db.Exec("INSERT INTO users(user, first_seen, last_seen) VALUES(?, ?, ?)", user, now, now)
		if err != nil {
			return 0, err
		}

		id, err = res.LastInsertId()
		return id, err
	}

	_, err = db.db.Exec("UPDATE users SET last_seen=? WHERE id=?", now, id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (db *db) insertPassword(password string) (int64, error) {
	var id int64 = -1

	err := db.db.QueryRow("SELECT id FROM passwords WHERE password=?", password).Scan(&id)

	now := time.Now().Unix()

	if id == -1 {
		res, err := db.db.Exec("INSERT INTO passwords(password, first_seen, last_seen) VALUES(?, ?, ?)", password, now, now)
		if err != nil {
			return 0, err
		}

		id, err = res.LastInsertId()
		return id, err
	}

	_, err = db.db.Exec("UPDATE passwords SET last_seen=? WHERE id=?", now, id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (db *db) insertVersion(version string) (int64, error) {
	var id int64 = -1

	err := db.db.QueryRow("SELECT id FROM versions WHERE version=?", version).Scan(&id)

	now := time.Now().Unix()

	if id == -1 {
		res, err := db.db.Exec("INSERT INTO versions(version, first_seen, last_seen) VALUES(?, ?, ?)", version, now, now)
		if err != nil {
			return 0, err
		}

		id, err = res.LastInsertId()
		return id, err
	}

	_, err = db.db.Exec("UPDATE versions SET last_seen=? WHERE id=?", now, id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (db *db) insertApp(app string) (int64, error) {
	var id int64 = -1

	err := db.db.QueryRow("SELECT id FROM applications WHERE app=?", app).Scan(&id)

	now := time.Now().Unix()

	if id == -1 {
		res, err := db.db.Exec("INSERT INTO applications(app, first_seen, last_seen) VALUES(?, ?, ?)", app, now, now)
		if err != nil {
			return 0, err
		}

		id, err = res.LastInsertId()
		return id, err
	}

	_, err = db.db.Exec("UPDATE applications SET last_seen=? WHERE id=?", now, id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (db *db) insertLogonAttempt(source, user, password, version, app int64) error {
	now := time.Now().Unix()
	_, err := db.db.Exec("INSERT INTO logon_attempts(source_id, user_id, pass_id, version_id, app_id, time) VALUES(?, ?, ?, ?, ?, ?)", source, user, password, version, app, now)

	if err != nil {
		return err
	}

	return nil
}

func (db *db) GetStats() (*Statistics, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	var err error

	var res Statistics
	res.RecentIPs, err = getRecent(db.db, "SELECT sources.ip, COUNT(logon_attempts.source_id) FROM sources LEFT JOIN logon_attempts ON sources.id = logon_attempts.source_id GROUP BY sources.ip")
	if err != nil {
		return nil, fmt.Errorf("RecentIPs: %s", err)
	}

	res.RecentUsers, err = getRecent(db.db, "SELECT users.user, COUNT(logon_attempts.user_id) FROM users LEFT JOIN logon_attempts ON users.id = logon_attempts.user_id GROUP BY users.user")
	if err != nil {
		return nil, fmt.Errorf("RecentUsers: %s", err)
	}

	res.RecentPasswords, err = getRecent(db.db, "SELECT passwords.password, COUNT(logon_attempts.pass_id) FROM passwords LEFT JOIN logon_attempts ON passwords.id = logon_attempts.pass_id GROUP BY passwords.password")
	if err != nil {
		return nil, fmt.Errorf("RecentPasswords: %s", err)
	}

	res.RecentVersions, err = getRecent(db.db, "SELECT versions.version, COUNT(logon_attempts.version_id) FROM versions LEFT JOIN logon_attempts ON versions.id = logon_attempts.version_id GROUP BY versions.version")
	if err != nil {
		return nil, fmt.Errorf("RecentVersions: %s", err)
	}

	return &res, nil
}

func getRecent(db *sql.DB, query string) (map[string]int, error) {
	rows, err := db.Query(query)
	if err != nil {
		logrus.Error("Failed to retrieve statistics")
		return nil, err
	}

	res := make(map[string]int)

	for rows.Next() {
		var str string
		var count int

		if err := rows.Scan(&str, &count); err != nil {
			return nil, err
		}

		res[str] = count
	}

	return res, nil
}
