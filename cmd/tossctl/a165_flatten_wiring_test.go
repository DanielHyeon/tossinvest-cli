package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

func TestA165FlattenWiringUsesTheBoundedNonDefinitivePolicy(t *testing.T) {
	policy := execgw.FlattenCancelRefusal()
	if policy == nil || *policy != (execgw.CancelRefusalPolicy{Retries: 2, MaxWait: 5 * time.Second, Window: 8 * time.Second}) {
		t.Fatalf("flatten policy: %+v", policy)
	}
	body, err := os.ReadFile("flatten.go")
	if err != nil {
		t.Fatalf("reading flatten.go: %v", err)
	}
	if !strings.Contains(string(body), "CancelRefusal: execgw.FlattenCancelRefusal()") {
		t.Fatal("buildFlattenWiring does not inject FlattenCancelRefusal")
	}
}
