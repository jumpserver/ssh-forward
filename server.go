package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	client *ssh.Client

	dstAddr string
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

type Response struct {
	Status int
	Host   string
	Port   int
	Err    string
	Msg    string
}

func (r Response) String() string {
	headers := make([]string, 0, 5)
	headers = append(headers, fmt.Sprintf("%s:%d", StatusHeader, r.Status))
	if r.Host != "" {
		headers = append(headers, fmt.Sprintf("%s:%s", HostHeader, r.Host))
	}
	if r.Port != 0 {
		headers = append(headers, fmt.Sprintf("%s:%d", PortHeader, r.Port))
	}

	if r.Err != "" {
		headers = append(headers, fmt.Sprintf("%s:%s", ErrHeader, r.Err))
	}
	headers = append(headers, fmt.Sprintf("%s:%s", MsgHeader, r.Msg))
	var buf bytes.Buffer
	for i := range headers {
		buf.WriteString(headers[i])
		buf.WriteString(LineSeparator)
	}
	buf.WriteString(LineSeparator)
	return buf.String()
}

func (r Response) Return() {
	// 是否退出程序
	exited := r.Status != SuccessStatus
	_, _ = io.WriteString(os.Stdout, r.String())
	if exited {
		os.Exit(1)
	}
}

const (
	SuccessStatus = 200
	BadStatus     = 400
)

const (
	StatusHeader = "status"
	HostHeader   = "host"
	PortHeader   = "port"
	ErrHeader    = "error"
	MsgHeader    = "message"

	LineSeparator = "\r\n"

	MsgOk = "ok"
)

const (
	ErrGateWay = "ErrGateway"
	ErrParams  = "ErrParams"
	ErrListen  = "ErrListen"
)

func NewSuccessResponse(addr *net.TCPAddr) Response {
	return Response{
		Status: SuccessStatus,
		Host:   addr.IP.String(),
		Port:   addr.Port,
		Msg:    MsgOk,
	}
}

func NewErrResponse(err, msg string) Response {
	return Response{
		Status: BadStatus,
		Err:    err,
		Msg:    msg,
	}
}
