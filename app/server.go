package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type HttpHeader struct {
	Path string
	Method string
	UserAgent string
}

var basePath *string

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	if len(os.Args) == 3 {
		str := os.Args[2]
		basePath = &str
	}
	fmt.Println(os.Args)

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

		go routeRequest(conn)
	}
}

func routeRequest(conn net.Conn) {
	buffer := make([]byte, 1024)
	conn.Read(buffer);
	fmt.Println(string(buffer))
	header := parseHttpHeader(buffer)

	defer conn.Close()
	if header.Method == "GET" && header.Path == "/"{
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if header.Method == "GET" && strings.HasPrefix(header.Path, "/echo") {
		echoHandler(conn, &header)
	} else if header.Method == "GET" && header.Path == "/user-agent" {
		userAgentHandler(conn, &header)
	} else if header.Method == "GET" && strings.HasPrefix(header.Path, "/files/")  {
		fileHandler(conn, &header)
	} else {
		conn.Write([]byte("HTTP/1.1 404\r\n\r\n"))
	}
}

func fileHandler(conn net.Conn, header *HttpHeader) {
	if basePath == nil {
		fmt.Println("Base path not set")
		writeHeaders(conn, 500, map[string]string{})
		return
	}
	fmt.Println("Base path: ", *basePath)
	filePath := header.Path[7:]
	fmt.Println("File path: ", filePath)
	file, err := os.Open(*basePath + "/" + filePath)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 404\r\n\r\n"))
		fmt.Println("Error opening file: ", err)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		conn.Write([]byte("HTTP/1.1 500\r\n\r\n"))
		return
	}

	writeHeaders(conn, 200, map[string]string{"Content-Type": "application/octet-stream", "Content-Length": fmt.Sprintf("%d", stat.Size())})
	io.Copy(conn, file)
}

func userAgentHandler(conn net.Conn, header *HttpHeader) {
	writeHeaders(conn, 200, map[string]string{"Content-Type": "text/plain", "Content-Length": fmt.Sprintf("%d", len(header.UserAgent))})
	conn.Write([]byte(header.UserAgent))
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

	// create a map with the rest of the headers
	headers := make(map[string]string)
	for _, line := range lines[1:] {
		if line == "" {
			break
		}
		header := strings.SplitN(line, ":", 2)
		headers[header[0]] = strings.TrimSpace(header[1])
	}

	return HttpHeader{Path: firstLine[1], Method: firstLine[0], UserAgent: headers["User-Agent"]}
}

func writeHeaders(conn net.Conn, status int, headers map[string]string) {
	// Implement the function here
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d\r\n", status)))
	for key, value := range headers {
		conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
	}
	conn.Write([]byte("\r\n"))
}