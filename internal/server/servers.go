package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"_http_protocol_1.1/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	done     chan struct{}
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		done:     make(chan struct{}),
	}

	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	if s.closed.Swap(true) {
		return nil
	}

	err := s.listener.Close()
	<-s.done
	return err
}

func (s *Server) listen() {
	defer close(s.done)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Println(err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()

	body := "Hello World!\n"
	h := response.GetDefaultHeaders(len(body))

	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Println(err)
		return
	}
	if err := response.WriteHeaders(conn, h); err != nil {
		log.Println(err)
		return
	}
	if _, err := conn.Write([]byte(body)); err != nil {
		log.Println(err)
	}
}
