package main

import (
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClientOption func(conf *SSHClientOptions)

type SSHClientOptions struct {
	Host         string
	Port         string
	Username     string
	Password     string
	PrivateKey   string
	Passphrase   string
	keyboardAuth ssh.KeyboardInteractiveChallenge
	PrivateAuth  ssh.Signer
}

func (cfg *SSHClientOptions) AuthMethods() []ssh.AuthMethod {
	authMethods := make([]ssh.AuthMethod, 0, 3)
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}
	if cfg.keyboardAuth == nil {
		cfg.keyboardAuth = func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			return []string{cfg.Password}, nil
		}
	}
	authMethods = append(authMethods, ssh.KeyboardInteractive(cfg.keyboardAuth))

	if cfg.PrivateKey != "" {
		var (
			signer ssh.Signer
			err    error
		)
		if cfg.Passphrase != "" {
			// 先使用 passphrase 解析 PrivateKey
			if signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(cfg.PrivateKey),
				[]byte(cfg.Passphrase)); err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
		if err != nil || cfg.Passphrase == "" {
			// 1. 如果之前使用解析失败，则去掉 passphrase，则尝试直接解析 PrivateKey 防止错误的passphrase
			// 2. 如果没有 Passphrase 则直接解析 PrivateKey
			if signer, err = ssh.ParsePrivateKey([]byte(cfg.PrivateKey)); err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}
	if cfg.PrivateAuth != nil {
		authMethods = append(authMethods, ssh.PublicKeys(cfg.PrivateAuth))
	}

	return authMethods
}

func SSHClientUsername(username string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Username = username
	}
}

func SSHClientPassword(password string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Password = password
	}
}

func SSHClientPrivateKey(privateKey string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.PrivateKey = privateKey
	}
}

func SSHClientPassphrase(passphrase string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Passphrase = passphrase
	}
}

func SSHClientHost(host string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Host = host
	}
}

func SSHClientPort(port string) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.Port = port
	}
}

func SSHClientPrivateAuth(privateAuth ssh.Signer) SSHClientOption {
	return func(args *SSHClientOptions) {
		args.PrivateAuth = privateAuth
	}
}

func SSHClientKeyboardAuth(keyboardAuth ssh.KeyboardInteractiveChallenge) SSHClientOption {
	return func(conf *SSHClientOptions) {
		conf.keyboardAuth = keyboardAuth
	}
}

func NewSSHClient(opts ...SSHClientOption) (*ssh.Client, error) {
	cfg := &SSHClientOptions{
		Host: "127.0.0.1",
		Port: "22",
	}
	for _, setter := range opts {
		setter(cfg)
	}
	return NewSSHClientWithCfg(cfg)
}

func NewSSHClientWithCfg(cfg *SSHClientOptions) (*ssh.Client, error) {
	sshCfg := ssh.ClientConfig{
		User:            cfg.Username,
		Auth:            cfg.AuthMethods(),
		Timeout:         5 * time.Minute,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	destAddr := net.JoinHostPort(cfg.Host, cfg.Port)
	sshClient, err := ssh.Dial("tcp", destAddr, &sshCfg)
	if err != nil {
		return nil, err
	}
	return sshClient, nil
}
