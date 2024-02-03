package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type HttpHeader struct {
	Path string
	Method string
	UserAgent string
	ContentLength int
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
	defer conn.Close()

	headerString := readUntilNewline(conn)
	// data, err := io.ReadAll(conn)
	// if err != nil {
	// 	fmt.Println("Error reading from connection: ", err)
	// 	writeHeaders(conn, 500, map[string]string{})
	// 	return
	
	// }
	// fmt.Printf("Data: %q\n", string(data))

	header := readHttpHeader([]byte(headerString))

	// _, fileData, _ := strings.Cut(string(data), "\r\n\r\n")

	
	if header.Method == "GET" && header.Path == "/"{
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if header.Method == "GET" && strings.HasPrefix(header.Path, "/echo") {
		echoHandler(conn, &header)
	} else if header.Method == "GET" && header.Path == "/user-agent" {
		userAgentHandler(conn, &header)
	} else if header.Method == "GET" && strings.HasPrefix(header.Path, "/files/")  {
		fileHandler(conn, &header)
	} else if header.Method == "POST" && strings.HasPrefix(header.Path, "/files/")  {
		handleFileUpload(conn, &header)
	} else {
		conn.Write([]byte("HTTP/1.1 404\r\n\r\n"))
	}
}

func handleFileUpload(conn net.Conn, header *HttpHeader) {
	// not a proper implementation, we aren't dealing with multipart/form-data or anything like that

	if basePath == nil {
		fmt.Println("Base path not set")
		writeHeaders(conn, 500, map[string]string{})
		return
	}
	fmt.Println("Base path: ", *basePath)
	filePath := header.Path[7:]
	fmt.Println("File path: ", filePath)
	file, err := os.OpenFile(*basePath + "/" + filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		writeHeaders(conn, 500, map[string]string{})
		fmt.Println("Error opening file: ", err)
		return
	}
	defer file.Close()


	buffer := make([]byte, header.ContentLength)
	n, err := io.ReadFull(conn, buffer)
	fmt.Println("Read ", n, " bytes")
	file.Write(buffer)
	if err == io.EOF {
		fmt.Println("EOF - no data read")
	} else if err != nil {
		writeHeaders(conn, 500, map[string]string{})
		fmt.Println("Error reading from connection: ", err)
		return
	}
	writeHeaders(conn, 201, map[string]string{})
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

func readHttpHeader(conn []byte) HttpHeader {
	reader := bufio.NewReader(bytes.NewReader(conn))
	builder := strings.Builder{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection: ", err)
			break
		}
		builder.WriteString(line)
		if line == "\r\n" {
			break
		}
	}
	fmt.Println("Header: ", builder.String())
	return parseHttpHeader(builder.String())
}


func parseHttpHeader(headerString string) HttpHeader {
	// Implement the function here
	lines := strings.Split(headerString, "\r\n")
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

	contentLength, err := strconv.Atoi(headers["Content-Length"])
	if err != nil {
		contentLength = 0
	}

	return HttpHeader{Path: firstLine[1], Method: firstLine[0], UserAgent: headers["User-Agent"], ContentLength: contentLength}
}

func writeHeaders(conn net.Conn, status int, headers map[string]string) {
	// Implement the function here
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d\r\n", status)))
	for key, value := range headers {
		conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
	}
	conn.Write([]byte("\r\n"))
}

func readUntilNewline(conn net.Conn) string {
	builder := strings.Builder{}
	for {
		buffer := make([]byte, 1)
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from connection: ", err)
			break
		}
		builder.Write(buffer)

		if strings.HasSuffix(builder.String(), "\r\n\r\n") {
			break
		}
	}
	return builder.String()
}