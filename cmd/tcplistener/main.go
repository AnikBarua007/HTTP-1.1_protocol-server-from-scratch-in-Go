// implement the http server functions
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer f.Close()

		currentLine := ""

		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err == io.EOF {
				if currentLine != "" {
					lines <- currentLine
				}
				return
			}
			if err != nil {
				log.Println(err)
				return
			}
			if n == 0 {
				continue
			}

			parts := strings.Split(string(buffer[:n]), "\n")
			for i, part := range parts {
				if i == len(parts)-1 {
					currentLine += part
					break
				}
				currentLine += part
				lines <- currentLine
				currentLine = ""
			}
		}
	}()

	return lines
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println("connection accepted")
		for line := range getLinesChannel(conn) {
			fmt.Printf("read: %s\n", line)
		}
		fmt.Println("connection closed")
	}
}
