[![Build Status](https://travis-ci.org/st3v/waitfor.svg?branch=master)](https://travis-ci.org/st3v/waitfor)
[![Coverage Status](https://coveralls.io/repos/st3v/waitfor/badge.svg?branch=master&service=github)](https://coveralls.io/github/st3v/waitfor?branch=master)

# waitfor

Command-line tool and Go library to wait for various conditions to become true. Inspired by Ansible's `wait_for`  [module](http://docs.ansible.com/ansible/wait_for_module.html). It currently supports port checks. More checks will be added in the future, e.g. checks for files, processes.

## Installation

Make sure Go is installed and setup correctly. To pull the library into the `$GOPATH` directory, build the binary, and put it into the `$GOBIN` directory, simply run:

```
go get github.com/st3v/waitfor/...
```

Assuming your `$PATH` contains `$GOBIN`, you can now run `waitfor` from anywhere on your machine.

```
waitfor --help
```

## Command-line Interface

### Wait for Host to Listen on Port

```
$ waitfor port --help
NAME:
   waitfor port - wait for host to listen on port

USAGE:
   waitfor port [command options] [arguments...]

OPTIONS:
   --closed, -c			wait for port to be closed
   --host, -h "127.0.0.1"	resolvable hostname or IP address
   --network, -n "tcp"		named network, ['tcp', 'tcp4', 'tcp6', 'udp', 'udp4', 'udp6', 'ip', 'ip4', 'ip6']
   --timeout, -t "5m0s"		maximum time to wait for
   --interval, -i "1s"		time in-between checks
   --verbose, -v		enable additional logging
```

For example, wait up to 1 minute for `localhost` to listen on port `8080` using the `tcp` protocol. Check port every 500 milliseconds.

```
waitfor port 8080 -h localhost -n tcp -t 1m -i 500ms
```

### Wait for Host to Stop Listening on Port

Use the `--closed` flag to wait for a port to be closed.

```
waitfor port 8080 -h localhost -n tcp -c
```
