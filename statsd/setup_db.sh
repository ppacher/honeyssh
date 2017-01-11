#!/bin/bash

sqlite3 ./analysis.db <<EOT
CREATE TABLE IF NOT EXISTS sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    ip VARCHAR(256) NOT NULL,
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER  PRIMARY KEY AUTOINCREMENT NOT NULL,
    user VARCHAR(256) NOT NULL,
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS passwords (
    id INTEGER  PRIMARY KEY AUTOINCREMENT NOT NULL,
    password VARCHAR(256) NOT NULL,
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS versions (
    id INTEGER  PRIMARY KEY AUTOINCREMENT NOT NULL,
    version VARCHAR(256) NOT NULL,
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS applications (
    id INTEGER  PRIMARY KEY AUTOINCREMENT NOT NULL,
    app VARCHAR(256) NOT NULL,
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS logon_attempts (
    id INTEGER  PRIMARY KEY AUTOINCREMENT NOT NULL,
    source_id INTEGER,
    user_id INTEGER,
    pass_id INTEGER,
    version_id INTEGER,
    app_id INTEGER,

    time INTEGER,

    FOREIGN KEY(source_id) REFERENCES sources(id),
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(pass_id) REFERENCES passwords(id),
    FOREIGN KEY(version_id) REFERENCES versions(id),
    FOREIGN KEY(app_id) REFERENCES applications(id)
);

EOT
