package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/coreos/go-systemd/activation"
)

func main() {
	echoValue := os.Getenv("ECHO_VALUE")
	Log("Starting; ECHO_VALUE=", echoValue)

	Log("LISTEN_PID", os.Getenv("LISTEN_PID"))
	Log("LISTEN_FDS", os.Getenv("LISTEN_FDS"))
	Log("LISTEN_FDNAMES", os.Getenv("LISTEN_FDNAMES"))

	listeners, err := activation.Listeners()
	if err != nil {
		panic(err)
	}

	if listeners == nil {
		panic("listeners == nil")
	}
	if len(listeners) != 1 {
		panic(fmt.Sprintf("Unexpected number of socket activation fds: %d", len(listeners)))
	}
	l := listeners[0]

	Log("Ready")
	for {
		conn, err := l.Accept()
		if err != nil {
			Log("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn, echoValue)
	}
}

func handleRequest(conn net.Conn, echoValue string) {
	Log("Handling connection")
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	for {
		Log("Reading command from connection")
		ln, err := r.ReadBytes('\n')
		if err != nil {
			Log("Error reading connection: ", err.Error())
			break
		}
		var result map[string]interface{}
		err = json.Unmarshal(ln, &result)
		if err != nil {
			Log("Error parsing message: ", err.Error())
			break
		}
		if result["cmd"] == "echo" {
			bytes, err := json.Marshal(echoValue)
			if err != nil {
				Log("Error marshalling echo response: ", err.Error())
				break
			}
			n, err := w.Write(bytes)
			if err != nil {
				Log("Error writing to socket; wrote ", n, " bytes; error: ", err.Error())
				break
			}
			err = w.WriteByte('\n')
			if err != nil {
				Log("Error writing endline to socket: ", err.Error())
				break
			}
			err = w.Flush()
			if err != nil {
				Log("Error flushing socket: ", err.Error())
				break
			}
		}
	}
	Log("Closing connection")
	conn.Close() // Ignoring Close() error
}

func Log(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}
