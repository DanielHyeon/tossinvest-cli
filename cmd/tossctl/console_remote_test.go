package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testCLISecret = "0123456789abcdef0123456789abcdef"

func TestConsoleOffersOnlyTheCompleteRemoteAccessFlagSet(t *testing.T) {
	cmd := findCommand(newRootCmd(), "console")
	if cmd == nil {
		t.Fatal("console is not registered")
	}
	for _, name := range []string{
		"port", "bind", "allowed-cidr", "public-url", "tls-cert", "tls-key",
		"remote-token-file", "trusted-network",
	} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("console has no --%s", name)
		}
	}
	for _, banned := range []string{
		"token", "session", "session-token", "csrf", "no-auth", "insecure",
		"approve", "auto-approve", "nonce", "confirm", "no-confirm",
	} {
		if cmd.Flags().Lookup(banned) != nil {
			t.Errorf("console offers unsafe --%s", banned)
		}
	}
}

func TestTrustedNetworkRemoteAccessDoesNotNeedATokenFile(t *testing.T) {
	got, err := remoteAccessOptions(&consoleOptions{
		bind:           "0.0.0.0",
		allowedCIDRs:   []string{"10.8.0.0/24"},
		publicURL:      "https://console.vpn.test:37085",
		tlsCert:        "/run/secrets/tls.crt",
		tlsKey:         "/run/secrets/tls.key",
		trustedNetwork: true,
	})
	if err != nil {
		t.Fatalf("trusted-network options: %v", err)
	}
	if !got.TrustedNetwork {
		t.Fatal("trusted-network selection was not carried into console options")
	}
	if got.AccessToken != "" {
		t.Fatal("trusted-network options unexpectedly contain an access token")
	}
}

func TestTrustedNetworkRejectsATokenFileConflict(t *testing.T) {
	_, err := remoteAccessOptions(&consoleOptions{
		bind:            "0.0.0.0",
		allowedCIDRs:    []string{"10.8.0.0/24"},
		publicURL:       "https://console.vpn.test:37085",
		tlsCert:         "/run/secrets/tls.crt",
		tlsKey:          "/run/secrets/tls.key",
		trustedNetwork:  true,
		remoteTokenFile: "/run/secrets/remote-token",
	})
	if err == nil {
		t.Fatal("trusted-network and remote-token-file were accepted together")
	}
}

func TestRemoteAccessTokenFileMustBePrivateAndLong(t *testing.T) {
	path := filepath.Join(t.TempDir(), "remote-token")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", 32)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	token, err := loadRemoteAccessToken(path)
	if err != nil {
		t.Fatalf("private token: %v", err)
	}
	if token != strings.Repeat("x", 32) {
		t.Fatalf("token = %q", token)
	}

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadRemoteAccessToken(path); err == nil {
		t.Fatal("world-readable token file was accepted")
	}

	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("too-short"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadRemoteAccessToken(path); err == nil {
		t.Fatal("short token was accepted")
	}
}

func TestRelaunchPreservesRemoteFlagsWithoutAddingASecret(t *testing.T) {
	args := []string{
		"tossctl", "--config-dir", "/cfg", "console",
		"--bind", "10.8.0.1", "--allowed-cidr", "10.8.0.0/24",
		"--public-url", "https://console.vpn.test:37085",
		"--tls-cert", "/run/secrets/tls.crt", "--tls-key", "/run/secrets/tls.key",
		"--remote-token-file", "/run/secrets/remote-token",
		"--port", "0",
	}
	got := argvWithPort(args, 37085)
	joined := strings.Join(got, " ")
	for _, want := range []string{
		"--bind 10.8.0.1", "--allowed-cidr 10.8.0.0/24",
		"--public-url https://console.vpn.test:37085",
		"--remote-token-file /run/secrets/remote-token", "--port 37085",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("relaunch argv missing %q: %s", want, joined)
		}
	}
	if strings.Contains(joined, testCLISecret) {
		t.Fatal("relaunch argv contains a credential value")
	}
}
