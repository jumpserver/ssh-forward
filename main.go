package main

import (
	"flag"
	"github.com/sevlyar/go-daemon"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type ForwardConfig struct {
	// SSH Server information
	proxyHost  string
	proxyPort  string
	proxyUser  string
	proxyPass  string
	remoteAddr string
}

func forward(localConn net.Conn, fc ForwardConfig) {
	// Setup sshClientConn (type *ssh.ClientConn)
	sshConfig := &ssh.ClientConfig{
		User: fc.proxyUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(fc.proxyPass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	serverAddr := fc.proxyHost + ":" + fc.proxyPort
	sshClientConn, err := ssh.Dial("tcp", serverAddr, sshConfig)
	log.Printf("Connect proxy server: %s", serverAddr)
	if err != nil {
		log.Fatalf("ssh.Dial failed: %s", err)
	}

	// Setup sshConn (type net.Conn)
	sshConn, err := sshClientConn.Dial("tcp", fc.remoteAddr)

	// Copy localConn.Reader to sshConn.Writer
	log.Printf("Connect remote server: %s", fc.remoteAddr)
	go func() {
		_, err = io.Copy(sshConn, localConn)
		if err != nil {
			log.Fatalf("io.Copy failed: %v", err)
		}
	}()

	// Copy sshConn.Reader to localConn.Writer
	go func() {
		_, err = io.Copy(localConn, sshConn)
		if err != nil {
			log.Fatalf("io.Copy failed: %v", err)
		}
	}()
}

func startForward(addr string, config ForwardConfig) {
	// Setup localListener (type net.Listener)
	localListener, err := net.Listen("tcp", addr)
	log.Printf("Listen at %s", addr)
	if err != nil {
		log.Fatalf("net.Listen failed: %v", err)
	}

	for {
		// Setup localConn (type net.Conn)
		localConn, err := localListener.Accept()
		if err != nil {
			log.Fatalf("listen.Accept failed: %v", err)
		}
		go forward(localConn, config)
	}
}

func startAsDaemon(addr string, config ForwardConfig) {
	ctx := &daemon.Context{
		PidFileName: "/tmp/" + addr + ".pid",
		PidFilePerm: 0644,
		LogFileName: "/tmp/" + addr + ".log",
		LogFilePerm: 0640,
		Umask:       027,
		WorkDir:     "./",
	}
	child, err := ctx.Reborn()
	if err != nil {
		log.Fatalf("run failed: %v", err)
	}
	if child != nil {
		return
	}
	defer ctx.Release()
	startForward(addr, config)
}

func main() {
	signal := flag.String("s", "start", "start | stop")
	asDaemon := flag.Bool("d", false, "As daemon")
	listenAddr := flag.String("listen", "12222", "Listen addr")
	proxyHost := flag.String("host", "127.0.0.1", "Proxy server host")
	proxyPort := flag.String("port", "22", "Proxy server port")
	proxyUser := flag.String("username", "root", "SSH username to connect")
	proxyPass := flag.String("password", "", "SSH password")
	remoteAddr := flag.String("remoteAddr", "1.1.1.1:3389", "Remote addr proxy connect to")
	flag.Parse()

	var addr string
	if strings.Contains(*listenAddr, ":") {
		addr = *listenAddr
	} else {
		addr = "127.0.0.1:" + *listenAddr
	}

	if *signal == "stop" {
		pidPath := "/tmp/" + addr + ".pid"
		logPath := "/tmp/" + addr + ".log"
		pid, err := ioutil.ReadFile(pidPath)
		if err != nil {
			log.Fatal("File not exist")
			return
		}
		pidInt, _ := strconv.Atoi(string(pid))
		err = syscall.Kill(pidInt, syscall.SIGTERM)
		if err != nil {
			log.Fatalf("Stop failed: %v", err)
		} else {
			os.Remove(pidPath)
			os.Remove(logPath)
		}
		return
	}

	config := ForwardConfig{
		proxyHost:  *proxyHost,
		proxyPort:  *proxyPort,
		proxyUser:  *proxyUser,
		proxyPass:  *proxyPass,
		remoteAddr: *remoteAddr,
	}

	if *asDaemon {
		startAsDaemon(addr, config)
	} else {
		startForward(addr, config)
	}
}
