package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type HttpHeader struct {
	Path string
	Method string
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		buffer := make([]byte, 1024)
		conn.Read(buffer);
		fmt.Println(string(buffer))
		header := parseHttpHeader(buffer)
		if header.Method == "GET" && header.Path == "/"{
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		} else if header.Method == "GET" && strings.HasPrefix(header.Path, "/echo") {
			echoHandler(conn, &header)
		} else {
			conn.Write([]byte("HTTP/1.1 404\r\n\r\n"))
		}
		conn.Close()
	}
}

func echoHandler(conn net.Conn, header *HttpHeader) {
	echoMessage := header.Path[6:]
	writeHeaders(conn, 200, map[string]string{"Content-Type": "text/plain", "Content-Length": fmt.Sprintf("%d", len(echoMessage))})
	conn.Write([]byte(echoMessage))
}

func parseHttpHeader(buffer []byte) HttpHeader {
	// Implement the function here
	lines := strings.Split(string(buffer), "\r\n")
	firstLine := strings.Split(lines[0], " ")
	return HttpHeader{Path: firstLine[1], Method: firstLine[0]}
}

func writeHeaders(conn net.Conn, status int, headers map[string]string) {
	// Implement the function here
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d\r\n", status)))
	for key, value := range headers {
		conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
	}
	conn.Write([]byte("\r\n"))
}