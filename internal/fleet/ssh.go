package fleet

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SSHConfig struct {
	User                 string
	Port                 int
	KeyFile              string
	Timeout              time.Duration
	StrictHostKeyChecking bool
}

type SSHPool struct {
	mu      sync.RWMutex
	clients map[string]*ssh.Client
	config  SSHConfig
}

type SSHResult struct {
	Hostname string
	Stdout   string
	Stderr   string
	Error    error
	Duration time.Duration
}

func NewSSHPool(config SSHConfig) *SSHPool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Port == 0 {
		config.Port = 22
	}

	return &SSHPool{
		clients: make(map[string]*ssh.Client),
		config:  config,
	}
}

func (p *SSHPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for hostname, client := range p.clients {
		client.Close()
		delete(p.clients, hostname)
	}
}

func (p *SSHPool) GetClient(host Host) (*ssh.Client, error) {
	p.mu.RLock()
	client, exists := p.clients[host.Hostname]
	p.mu.RUnlock()

	if exists {
		_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
		if err == nil {
			return client, nil
		}
		p.mu.Lock()
		delete(p.clients, host.Hostname)
		p.mu.Unlock()
	}

	sshConfig := p.resolveConfig(host)

	client, err := ssh.Dial("tcp", sshConfig.addr, sshConfig.config)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to %s: %w", host.Hostname, err)
	}

	p.mu.Lock()
	p.clients[host.Hostname] = client
	p.mu.Unlock()

	return client, nil
}

type resolvedSSHConfig struct {
	addr   string
	config *ssh.ClientConfig
}

func (p *SSHPool) resolveConfig(host Host) resolvedSSHConfig {
	port := p.config.Port
	if host.Port > 0 {
		port = host.Port
	}

	user := p.config.User
	if host.User != "" {
		user = host.User
	}

	addr := net.JoinHostPort(host.Address, fmt.Sprintf("%d", port))

	keyFile := p.config.KeyFile
	if host.KeyFile != "" {
		keyFile = host.KeyFile
	}

	auths := []ssh.AuthMethod{}

	if keyFile != "" {
		signer, err := loadKey(keyFile)
		if err == nil {
			auths = append(auths, signer)
		}
	}

	hostKeyCallback := ssh.InsecureIgnoreHostKey()
	if p.config.StrictHostKeyChecking {
		hkcb, err := knownhosts.New(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
		if err == nil {
			hostKeyCallback = hkcb
		}
	}

	sshConf := &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         p.config.Timeout,
	}

	return resolvedSSHConfig{addr: addr, config: sshConf}
}

func loadKey(keyFile string) (ssh.AuthMethod, error) {
	expanded := keyFile
	if len(keyFile) > 0 && keyFile[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		expanded = filepath.Join(home, keyFile[1:])
	}

	key, err := os.ReadFile(expanded)
	if err != nil {
		return nil, fmt.Errorf("cannot read key file %s: %w", expanded, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("cannot parse key file %s: %w", expanded, err)
	}

	return ssh.PublicKeys(signer), nil
}

func (p *SSHPool) Execute(host Host, command string) SSHResult {
	start := time.Now()

	client, err := p.GetClient(host)
	if err != nil {
		return SSHResult{
			Hostname: host.Hostname,
			Error:    err,
			Duration: time.Since(start),
		}
	}

	session, err := client.NewSession()
	if err != nil {
		return SSHResult{
			Hostname: host.Hostname,
			Error:    fmt.Errorf("cannot create session: %w", err),
			Duration: time.Since(start),
		}
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return SSHResult{
			Hostname: host.Hostname,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Error:    fmt.Errorf("command failed: %w", err),
			Duration: time.Since(start),
		}
	}

	return SSHResult{
		Hostname: host.Hostname,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(start),
	}
}

func (p *SSHPool) ExecuteWithTimeout(host Host, command string, timeout time.Duration) SSHResult {
	type result struct {
		res SSHResult
	}

	ch := make(chan result, 1)
	go func() {
		ch <- result{res: p.Execute(host, command)}
	}()

	select {
	case r := <-ch:
		return r.res
	case <-time.After(timeout):
		return SSHResult{
			Hostname: host.Hostname,
			Error:    fmt.Errorf("command timed out after %v", timeout),
			Duration: timeout,
		}
	}
}
