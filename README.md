# honeyssh

`honeyssh` is a Honey-Pod for SSH. It automatically logs username and password
tries during brute-force attacks.

This repository contains two applications, `honeyssh` itself as well as `statsd`
which provides a simple HTTP API to query logon attempts.

## Installation and Usage

Installation of `honeyssh` requires a working `GoLang` environment:

```bash
go install github.com/nethack42/honeyssh
```

When started without a configuration file or commandline parameters, `honeyssh`
will start in user/password logging mode. Incoming authentication requests are
logged and afterwards rejected with "wrong password".

```bash
$ ./setup_db.sh
$ nohup ./statsd &
```

```bash
sudo ./honeyssh --listen 0.0.0.0:22
...
INFO[57] Logon attempt: host=112.99.218.173:51445 version=SSH-2.0-sshlib-0.1 user="root" pass="system"
INFO[57] Logon attempt: host=112.99.218.173:51445 version=SSH-2.0-sshlib-0.1 user="root" pass="raspi"
INFO[57] Logon attempt: host=112.99.218.173:51445 version=SSH-2.0-sshlib-0.1 user="root" pass="ubnt"
INFO[57] Logon attempt: host=112.99.218.173:51445 version=SSH-2.0-sshlib-0.1 user="root" pass="00000000"
...
```


Query logon attempts:

```bash
$ curl http://localhost:4000/stats
{
    "recent_ips": {
        "153.99.182.12": 102,
        "185.110.132.202": 1,
        "185.29.9.169": 3
    },
    "recent_usernames": {
        "globalflash": 3,
        "root": 102,
        "test": 1
    },
    "recent_passwords": {
        "!QAZxsw2#EDC": 1,
        "!QAZzaq1": 1,
        "!qa2ws3ed": 1,
        "00": 1,
        "0o9i8u7y": 1,
        "100200": 1,
        "10203040": 1,
        "110120": 1,
        "1111111111": 1,
        "121314": 1,
        "123456789a": 1,
        "123456789a123": 1,
        "123456a?": 1,
        "1234abc": 1,
        "123qwe,.": 1,
        "159159": 1,
        "1qaz2wsx#EDC": 1,
        "1qaz3edc": 1,
        "1qazse4": 1,
        "3.1415": 1,
        "4444444": 1,
        "@dmin": 1,
        "Nopass@elong.com": 1,
        "P@ssword123456": 1,
        "Pass@1234": 1,
        "abc.123": 1,
    }
}
```

