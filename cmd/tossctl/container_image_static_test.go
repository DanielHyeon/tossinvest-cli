package main

import (
	"os"
	"strings"
	"testing"
)

func TestContainerImagePinsEntrypointExecutableMode(t *testing.T) {
	body, err := os.ReadFile("../../Dockerfile")
	if err != nil {
		t.Fatal(err)
	}
	const executableCopy = "COPY --chmod=0755 deploy/container-entrypoint.sh /usr/local/bin/tossos-entrypoint"
	if !strings.Contains(string(body), executableCopy) {
		t.Fatalf("Dockerfile must pin the entrypoint mode with %q", executableCopy)
	}
}
