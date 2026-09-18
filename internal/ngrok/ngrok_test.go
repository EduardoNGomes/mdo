package ngrok

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestPublicURL(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{`{"url":"https://example.ngrok-free.app"}`, "https://example.ngrok-free.app"},
		{`{"url":"https://example.ngrok-free.app/"}`, "https://example.ngrok-free.app"},
		{`{"url":"http://example.ngrok-free.app"}`, ""},
		{`not json`, ""},
	}
	for _, tt := range tests {
		if got := publicURL(tt.line); got != tt.want {
			t.Errorf("publicURL(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestCommandArgsUseOnlyFreeHTTPTunnelOptions(t *testing.T) {
	want := []string{"http", "127.0.0.1:8080", "--log=stdout", "--log-format=json"}
	if got := commandArgs("127.0.0.1:8080"); !reflect.DeepEqual(got, want) {
		t.Errorf("commandArgs() = %q, want %q", got, want)
	}
}

func TestStartupErrorIncludesConfigurationHelp(t *testing.T) {
	err := startupError(nil, "authtoken is required")
	message := err.Error()
	for _, want := range []string{
		"ngrok could not start",
		signupURL,
		downloadURL,
		"ngrok config add-authtoken <YOUR_TOKEN>",
		quickstartURL,
	} {
		if !strings.Contains(message, want) {
			t.Errorf("startupError() = %q, want %q", message, want)
		}
	}
}

func TestLookupExplainsHowToSetUpNgrok(t *testing.T) {
	original := findExecutable
	t.Cleanup(func() { findExecutable = original })
	findExecutable = func(string) (string, error) { return "", errors.New("not found") }

	_, err := Lookup()
	if err == nil {
		t.Fatal("Lookup() error = nil, want setup instructions")
	}
	for _, want := range []string{signupURL, downloadURL, quickstartURL, "ngrok config add-authtoken <YOUR_TOKEN>"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Lookup() error = %q, want %q", err, want)
		}
	}
}
