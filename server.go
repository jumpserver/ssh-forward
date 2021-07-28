package main

import (
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	addr string

	client *ssh.Client

	dstAddr string
}

func (s *Server) ListenAndServe() {

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("listen addr %s err: %s\n", s.addr, err)
	}
	log.Printf("Listen at %s", s.addr)
	log.Println(s.Serve(ln))
}

func (s *Server) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("listen accept failed: %v", err)
		}

		go s.forward(conn)
	}
}

func (s *Server) forward(conn net.Conn) {
	defer conn.Close()

	proxyCon, err := s.client.Dial("tcp", s.dstAddr)
	if err != nil {
		log.Fatalf("ssh.Dial failed: %s\n", err)
	}
	go func() {
		defer proxyCon.Close()
		if _, err = io.Copy(proxyCon, conn); err != nil {
			log.Printf("io.Copy local-> proxy err: %s\n", err)
		}
	}()
	if _, err := io.Copy(conn, proxyCon); err != nil {
		log.Printf("io.Copy proxy -> local err: %s\n", err)
	}

}
