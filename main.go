package main

import (
	"encoding/base64"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/sevlyar/go-daemon"
)

func main() {
	signal := flag.String("s", "start", "start | stop")
	asDaemon := flag.Bool("d", false, "As daemon")
	listenAddr := flag.String("listen", "12222", "Listen addr")
	proxyHost := flag.String("host", "127.0.0.1", "Proxy server host")
	proxyPort := flag.String("port", "22", "Proxy server port")
	proxyUser := flag.String("username", "root", "SSH username to connect")
	proxyPass := flag.String("password", "", "SSH password")
	proxyKey := flag.String("privateKey", "", "SSH private key path")
	proxyB64PrivateKey := flag.String("privateKey_b64", "", "SSH private key bs64 string")
	b64 := flag.Bool("b64", false, "Encoding pass")
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

	password := *proxyPass
	if *b64 {
		passwordBytes, err := base64.StdEncoding.DecodeString(password)
		if err != nil {
			log.Fatal("Decode password error")
			return
		}
		password = string(passwordBytes)
	}
	var privateKey string
	if *proxyB64PrivateKey != "" {
		passwordBytes, err := base64.StdEncoding.DecodeString(*proxyB64PrivateKey)
		if err != nil {
			log.Fatalf("Decode private key error %s", err)
			return
		}
		privateKey = string(passwordBytes)
	}
	if *proxyKey != "" {
		content, err := ioutil.ReadFile(*proxyKey)
		if err != nil {
			log.Fatalf("Read private key err: %s", err)
		}
		privateKey = string(content)
	}

	sshOptions := make([]SSHClientOption, 0, 5)

	sshOptions = append(sshOptions, SSHClientHost(*proxyHost))
	sshOptions = append(sshOptions, SSHClientPort(*proxyPort))
	sshOptions = append(sshOptions, SSHClientUsername(*proxyUser))
	sshOptions = append(sshOptions, SSHClientPassword(password))

	if privateKey != "" {
		sshOptions = append(sshOptions, SSHClientPassphrase(password))
		sshOptions = append(sshOptions, SSHClientPrivateKey(privateKey))
	}
	sshClient, err := NewSSHClient(sshOptions...)
	if err != nil {
		log.Fatalf("SSH Client err: %s\n", err)
	}

	srv := Server{
		addr:    addr,
		client:  sshClient,
		dstAddr: *remoteAddr,
	}

	if *asDaemon {
		startAsDaemon(&srv)
	} else {
		srv.ListenAndServe()
	}
}

func startAsDaemon(srv *Server) {
	ctx := &daemon.Context{
		PidFileName: "/tmp/" + srv.addr + ".pid",
		PidFilePerm: 0644,
		LogFileName: "/tmp/" + srv.addr + ".log",
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
	srv.ListenAndServe()
}