// Package ngrok starts a free ngrok HTTP tunnel for a local address.
package ngrok

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	quickstartURL  = "https://ngrok.com/docs/share-localhost/quickstart"
	downloadURL    = "https://ngrok.com/download"
	signupURL      = "https://dashboard.ngrok.com/signup"
	startupTimeout = 15 * time.Second
)

var findExecutable = exec.LookPath

// Lookup verifies that ngrok is available before a local server is started.
func Lookup() (string, error) {
	path, err := findExecutable("ngrok")
	if err != nil {
		return "", setupError("--live requires ngrok")
	}
	return path, nil
}

// Tunnel is a running ngrok agent process.
type Tunnel struct {
	url      string
	cmd      *exec.Cmd
	finished <-chan struct{}
	mu       sync.Mutex
	waitErr  error
}

// URL is the public HTTPS URL assigned by ngrok.
func (t *Tunnel) URL() string { return t.url }

// Done is closed when the ngrok process stops.
func (t *Tunnel) Done() <-chan struct{} { return t.finished }

// Err returns the process result after Done has been closed.
func (t *Tunnel) Err() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.waitErr
}

// Close stops the agent and waits briefly for it to exit.
func (t *Tunnel) Close() error {
	if t == nil || t.cmd.Process == nil {
		return nil
	}
	if err := t.cmd.Process.Signal(os.Interrupt); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	select {
	case <-t.finished:
		return t.Err()
	case <-time.After(2 * time.Second):
		if err := t.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return err
		}
		<-t.finished
		return t.Err()
	}
}

// Start starts a free HTTP endpoint forwarding to target, such as 127.0.0.1:8080.
func Start(binary, target string) (*Tunnel, error) {
	cmd := exec.Command(binary, commandArgs(target)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("prepare ngrok output: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("prepare ngrok errors: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start ngrok: %w", err)
	}

	lines := make(chan string, 32)
	var readers sync.WaitGroup
	readers.Add(2)
	go readLines(stdout, lines, &readers)
	go readLines(stderr, lines, &readers)
	go func() {
		readers.Wait()
		close(lines)
	}()
	finished := make(chan struct{})
	tunnel := &Tunnel{cmd: cmd, finished: finished}
	go func() {
		tunnel.mu.Lock()
		tunnel.waitErr = cmd.Wait()
		tunnel.mu.Unlock()
		close(finished)
	}()

	var output bytes.Buffer
	timer := time.NewTimer(startupTimeout)
	defer timer.Stop()
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				<-finished
				return nil, startupError(tunnel.Err(), output.String())
			}
			appendOutput(&output, line)
			if url := publicURL(line); url != "" {
				tunnel.url = url
				go drain(lines)
				return tunnel, nil
			}
		case <-finished:
			return nil, startupError(tunnel.Err(), output.String())
		case <-timer.C:
			_ = cmd.Process.Kill()
			<-finished
			return nil, startupError(errors.New("timed out waiting for a public URL"), output.String())
		}
	}
}

func commandArgs(target string) []string {
	return []string{"http", target, "--log=stdout", "--log-format=json"}
}

func drain(lines <-chan string) {
	for range lines {
	}
}

func readLines(reader io.Reader, lines chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024), 64<<10)
	for scanner.Scan() {
		lines <- scanner.Text()
	}
}

func publicURL(line string) string {
	var entry struct {
		URL string `json:"url"`
	}
	if json.Unmarshal([]byte(line), &entry) != nil || !strings.HasPrefix(entry.URL, "https://") {
		return ""
	}
	return strings.TrimSuffix(entry.URL, "/")
}

func appendOutput(output *bytes.Buffer, line string) {
	const maxOutput = 4096
	if output.Len() >= maxOutput {
		return
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	remaining := maxOutput - output.Len()
	if len(line) > remaining {
		line = line[:remaining]
	}
	output.WriteString(line)
	output.WriteByte('\n')
}

func startupError(err error, output string) error {
	message := "ngrok could not start"
	if err != nil {
		message += ": " + err.Error()
	}
	if output != "" {
		message += "\nngrok output: " + strings.TrimSpace(output)
	}
	return setupError(message)
}

func setupError(message string) error {
	return fmt.Errorf("%s\n1. Create an account: %s\n2. Install ngrok: %s\n3. Configure it: ngrok config add-authtoken <YOUR_TOKEN>\nQuickstart: %s", message, signupURL, downloadURL, quickstartURL)
}
