package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
)

func TestHTTPAPIMutationsStayAbsentUntilBothSecurityArtifactsAreConfigured(t *testing.T) {
	routes, err := httpAPIMutationRoutes(nil, nil, &httpAPIOptions{}, nil)
	if err != nil || routes != nil {
		t.Fatalf("default no-token routes = %#v err=%v", routes, err)
	}
	for _, opts := range []*httpAPIOptions{{securityDB: "security.db"}, {capabilityPublicKey: "public.key"}} {
		if _, err := httpAPIMutationRoutes(nil, nil, opts, nil); err == nil {
			t.Fatal("partial mutation security configuration was accepted")
		}
	}
}

func TestHTTPAPICapabilityPublicKeyLoaderAcceptsOnlyOneExactEd25519Key(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	secureDirectory := t.TempDir()
	if err := os.Chmod(secureDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(secureDirectory, "capability-public.key")
	if err := os.WriteFile(path, []byte(base64.RawStdEncoding.EncodeToString(publicKey)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadHTTPAPICapabilityPublicKey(path)
	if err != nil || !publicKey.Equal(loaded) {
		t.Fatalf("loaded key equal=%v err=%v", publicKey.Equal(loaded), err)
	}
	if err := os.WriteFile(path, []byte("not-a-key"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadHTTPAPICapabilityPublicKey(path); err == nil {
		t.Fatal("malformed public key accepted")
	}
	valid := []byte(base64.RawStdEncoding.EncodeToString(publicKey) + "\n")
	if err := os.WriteFile(path, valid, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := loadHTTPAPICapabilityPublicKey(path); err == nil {
		t.Fatal("group/world-writable public key accepted")
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "capability-public-link.key")
	if err := os.Symlink(path, link); err == nil {
		if _, err := loadHTTPAPICapabilityPublicKey(link); err == nil {
			t.Fatal("symlinked public key accepted")
		}
	}
}

func TestHTTPAPIApplicationsCannotApplyCurrentContextualProtectionField(t *testing.T) {
	ctx := context.Background()
	registry, err := optimization.CoreRegistry(ctx)
	if err != nil {
		t.Fatal(err)
	}
	store, err := optimization.Open(ctx, optimization.Options{
		Path: filepath.Join(t.TempDir(), optimization.DatabaseFileName), Registry: registry, Actor: "fixture",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	identity := httpapi.MutationIdentity{Actor: "mobile", Client: "phone", Mode: httpapi.AuthModeCapability}
	commander, err := store.ForActor(httpAPIOptimizationActor(identity))
	if err != nil {
		t.Fatal(err)
	}
	view, err := commander.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	field, ok := view.Registry.Field("exit.common-policy")
	if !ok || len(field.Descriptor.Options) == 0 {
		t.Fatal("current owner registry has no finite protection preset")
	}
	preview, err := commander.Preview(ctx, optimization.PreviewRequest{
		BaseVersion: view.Snapshot.Version, Category: optimization.CategoryExitProtection,
		Changes: map[string]string{"exit.common-policy": field.Descriptor.Options[0].ID},
		Source:  optimization.SourceServerPreset, Reason: optimization.ReasonServerPreset,
	})
	if err != nil {
		t.Fatal(err)
	}
	mutation := httpAPIOptimizationMutation{store: store, kind: httpAPIOptimizationApply}
	body := []byte(fmt.Sprintf(`{"capability":%q,"confirmed":true}`, preview.Capability))
	_, err = mutation.Execute(ctx, httpapi.AuthorizedMutation{
		Identity:      identity,
		CanonicalBody: body, IfMatch: quoteHTTPAPIVersion(view.Snapshot.Version),
	})
	if err == nil {
		t.Fatal("remote contextual protection application was accepted")
	}
	var commandError *httpapi.CommandError
	if !errors.As(err, &commandError) || commandError.StatusCode() != http.StatusForbidden ||
		commandError.ErrorCode() != "REMOTE_TRANSITION_REFUSED" {
		t.Fatalf("contextual application error=%T %v", err, err)
	}
	after, err := store.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if after.Snapshot.Version != view.Snapshot.Version {
		t.Fatalf("version changed from %d to %d", view.Snapshot.Version, after.Snapshot.Version)
	}
}

func TestHTTPAPIPreviewRejectsArbitraryInputAsBadRequest(t *testing.T) {
	_, err := executeHTTPAPIOptimizationPreview(context.Background(), nil,
		[]byte(`{"category":"exit-protection","baseVersion":1,"changes":[]}`), 1)
	var commandError *httpapi.CommandError
	if !errors.As(err, &commandError) || commandError.StatusCode() != http.StatusBadRequest ||
		commandError.ErrorCode() != "INVALID_SELECTION" {
		t.Fatalf("invalid preview error=%T %v", err, err)
	}
}

func TestHTTPAPIOptimizationActorScopeCannotCollideAcrossActorAndClient(t *testing.T) {
	left := httpAPIOptimizationActor(httpapi.MutationIdentity{Actor: "a@b", Client: "c", Mode: httpapi.AuthModeCapability})
	right := httpAPIOptimizationActor(httpapi.MutationIdentity{Actor: "a", Client: "b@c", Mode: httpapi.AuthModeCapability})
	if left == right {
		t.Fatalf("structured actor scopes collided: %q", left)
	}
}

func TestHTTPAPIRemoteApplicationTransitionRegistryStartsEmpty(t *testing.T) {
	if len(remoteApplicableOptimizationTransitions) != 0 {
		t.Fatalf("a051 unexpectedly exposes remote effective transitions: %v", remoteApplicableOptimizationTransitions)
	}
}
