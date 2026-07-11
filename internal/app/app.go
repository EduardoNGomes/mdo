package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/egomes/mdo/internal/browser"
	"github.com/egomes/mdo/internal/document"
	webassets "github.com/egomes/mdo/internal/web"
)

const (
	maxMarkdownSize = 16 << 20
	shutdownTimeout = 20 * time.Second
)

var errUsage = errors.New("uso: mdo <arquivo.md>")

func Run(args []string) error {
	if len(args) != 1 {
		return errUsage
	}

	path, source, err := readMarkdown(args[0])
	if err != nil {
		return err
	}

	content, err := document.Render(source)
	if err != nil {
		return fmt.Errorf("renderizar markdown: %w", err)
	}

	token, err := randomToken()
	if err != nil {
		return fmt.Errorf("gerar endereço seguro: %w", err)
	}

	page, err := webassets.Page(webassets.PageData{
		Title:   filepath.Base(path),
		Path:    path,
		Content: content,
		Token:   token,
	})
	if err != nil {
		return fmt.Errorf("montar página: %w", err)
	}

	return serveAndOpen(path, token, page)
}

func readMarkdown(input string) (string, []byte, error) {
	if filepath.Ext(input) != ".md" {
		return "", nil, fmt.Errorf("somente arquivos com extensão .md são aceitos")
	}

	path, err := filepath.Abs(input)
	if err != nil {
		return "", nil, fmt.Errorf("resolver caminho: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, fmt.Errorf("abrir %q: %w", input, err)
	}
	if !info.Mode().IsRegular() {
		return "", nil, fmt.Errorf("%q não é um arquivo regular", input)
	}
	if info.Size() > maxMarkdownSize {
		return "", nil, fmt.Errorf("arquivo excede o limite de %d MiB", maxMarkdownSize>>20)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", nil, fmt.Errorf("abrir %q: %w", input, err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxMarkdownSize+1))
	if err != nil {
		return "", nil, fmt.Errorf("ler %q: %w", input, err)
	}
	if len(data) > maxMarkdownSize {
		return "", nil, fmt.Errorf("arquivo excede o limite de %d MiB", maxMarkdownSize>>20)
	}
	return path, data, nil
}

func serveAndOpen(markdownPath, token string, page []byte) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("abrir porta local: %w", err)
	}

	basePath := "/" + token + "/"
	readyPath := basePath + "ready"
	ready := make(chan struct{})
	var readyOnce sync.Once
	assets := http.StripPrefix(basePath, http.FileServer(http.Dir(filepath.Dir(markdownPath))))

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+basePath, func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = io.Copy(w, bytes.NewReader(page))
	})
	mux.HandleFunc("POST "+readyPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-MDO-Token") != token {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		readyOnce.Do(func() { close(ready) })
	})
	mux.Handle(basePath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w)
		assets.ServeHTTP(w, r)
	}))

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()

	url := "http://" + listener.Addr().String() + basePath
	if err := browser.Open(url); err != nil {
		_ = server.Close()
		return fmt.Errorf("abrir navegador (%s): %w", url, err)
	}

	select {
	case <-ready:
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("servidor local: %w", err)
		}
	case <-time.After(shutdownTimeout):
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		_ = server.Close()
	}
	return nil
}

func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", strings.Join([]string{
		"default-src 'none'",
		"script-src 'unsafe-inline'",
		"style-src 'unsafe-inline'",
		"img-src 'self' data: https: http:",
		"font-src data:",
		"connect-src 'self'",
	}, "; "))
}

func randomToken() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
