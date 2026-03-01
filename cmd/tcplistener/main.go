// implement the http server functions
package main

import (
	"fmt"
	"log"
	"net"

	"_http_protocol_1.1/internal/request"
)

//func getLinesChannel(f io.ReadCloser) <-chan string {
//	lines := make(chan string)
//
//	go func() {
//		defer close(lines)
//		defer f.Close()
//
//		currentLine := ""
//
//		for {
//			buffer := make([]byte, 8)
//			n, err := f.Read(buffer)
//			if err == io.EOF {
//				if currentLine != "" {
//					lines <- currentLine
//				}
//				return
//			}
//			if err != nil {
//				log.Println(err)
//				return
//			}
//			if n == 0 {
//				continue
//			}
//
//			parts := strings.Split(string(buffer[:n]), "\n")
//			for i, part := range parts {
//				if i == len(parts)-1 {
//					currentLine += part
//					break
//				}
//				currentLine += part
//				lines <- currentLine
//				currentLine = ""
//			}
//		}
//	}()
//
//	return lines
//}

// explanations
// •
// lines := make(chan string) creates an unbuffered channel of lines.
// •
// The function starts a goroutine so reading/parsing happens asynchronously.
// •
// The function immediately returns the receive-only channel (<-chan string) to the caller.
// •
// Inside the goroutine, file bytes are read in chunks (8 bytes each), accumulated into currentLine, and split on \n.
// •
// Every complete line is sent with lines <- currentLine.
// •
// If EOF is reached and currentLine still has content, it sends that final partial line once.
// •
// defer close(lines) closes the channel when done, so consumers can for line := range ch safely terminate.
// •
// defer f.Close() ensures the reader is closed when goroutine exits.
// Concurrency behavior:
// •
// Because the channel is unbuffered, each lines <- ... blocks until a receiver reads.
// •
// That gives natural backpressure: producer won’t outrun consumer.
// One caveat:
// •
// If input uses \r\n, emitted lines may include trailing \r. You may want strings.TrimSuffix(line, "\r") before sending.
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
		//for line := range getLinesChannel(conn) {
		//	fmt.Printf("read: %s\n", line)
		//}
		//fmt.Println("connection closed")
		reqLine, err := request.RequestFromReader(conn)
		fmt.Printf("Request line: \n- Method:%s\n- Target:%s\n- Version: %s\n", reqLine.RequestLine.Method, reqLine.RequestLine.RequestTarget, reqLine.RequestLine.HttpVersion)

	}
}
