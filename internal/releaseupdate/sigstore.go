package releaseupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/version"
	in_toto "github.com/in-toto/attestation/go/v1"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/verify"
)

const (
	actionsIssuer       = "https://token.actions.githubusercontent.com"
	slsaProvenanceV1    = "https://slsa.dev/provenance/v1"
	statementV1         = "https://in-toto.io/Statement/v1"
	pinnedTUFRootSHA256 = "a0dfcc5d51c1ce4a66b541a3fff0afa97225ccc40456b21140bb2e2f113122e2"
)

type productionVerifier struct {
	cacheDir string
}

type bundleExpectation struct {
	DigestAlgorithm string
	Digest          []byte
	AssetName       string
	Issuer          string
	SAN             string
	PredicateType   string
	Repository      string
	WorkflowPath    string
	Ref             string
}

func newProductionVerifier(cacheDir string) (*productionVerifier, error) {
	if err := verifyPinnedTUFRoot(); err != nil {
		return nil, err
	}
	if err := ensurePrivateCache(cacheDir); err != nil {
		return nil, err
	}
	return &productionVerifier{cacheDir: cacheDir}, nil
}

func (v *productionVerifier) Verify(
	ctx context.Context,
	bundleJSON []byte,
	digest, tag, asset string,
) (Provenance, error) {
	if !stableTag.MatchString(tag) {
		return Provenance{}, fmt.Errorf("releaseupdate: invalid attested tag %q", tag)
	}
	digestBytes, err := hex.DecodeString(digest)
	if err != nil || len(digestBytes) != sha256.Size {
		return Provenance{}, errors.New("releaseupdate: invalid SHA-256 digest for attestation")
	}
	if err := ensurePrivateCache(v.cacheDir); err != nil {
		return Provenance{}, err
	}
	fetcher, err := newTUFFetcher(ctx, nil, productionTUFURLAllowed)
	if err != nil {
		return Provenance{}, err
	}
	options := tuf.DefaultOptions().
		WithContext(ctx).
		WithCachePath(v.cacheDir).
		WithFetcher(fetcher)
	trusted, err := root.FetchTrustedRootWithOptions(options)
	if err != nil {
		return Provenance{}, fmt.Errorf("releaseupdate: refreshing pinned-root Sigstore trust: %w", err)
	}
	ref := "refs/tags/" + tag
	san := "https://github.com/" + versionRepo() + "/.github/workflows/release.yml@" + ref
	return verifyBundleWithTrustedRoot(bundleJSON, trusted, bundleExpectation{
		DigestAlgorithm: "sha256",
		Digest:          digestBytes,
		AssetName:       asset,
		Issuer:          actionsIssuer,
		SAN:             san,
		PredicateType:   slsaProvenanceV1,
		Repository:      "https://github.com/" + versionRepo(),
		WorkflowPath:    ".github/workflows/release.yml",
		Ref:             ref,
	})
}

func verifyBundleWithTrustedRoot(
	bundleJSON []byte,
	trusted root.TrustedMaterial,
	expect bundleExpectation,
) (Provenance, error) {
	entity := &bundle.Bundle{}
	if err := entity.UnmarshalJSON(bundleJSON); err != nil {
		return Provenance{}, fmt.Errorf("releaseupdate: decoding Sigstore bundle: %w", err)
	}
	return verifySignedEntityWithTrustedRoot(entity, trusted, expect)
}

func verifySignedEntityWithTrustedRoot(
	entity verify.SignedEntity,
	trusted root.TrustedMaterial,
	expect bundleExpectation,
) (Provenance, error) {
	identity, err := verify.NewShortCertificateIdentity(expect.Issuer, "", expect.SAN, "")
	if err != nil {
		return Provenance{}, fmt.Errorf("releaseupdate: constructing signer identity: %w", err)
	}
	verifier, err := verify.NewVerifier(
		trusted,
		verify.WithTransparencyLog(1),
		verify.WithIntegratedTimestamps(1),
		verify.WithSignedCertificateTimestamps(1),
	)
	if err != nil {
		return Provenance{}, fmt.Errorf("releaseupdate: constructing Sigstore verifier: %w", err)
	}
	result, err := verifier.Verify(entity, verify.NewPolicy(
		verify.WithArtifactDigest(expect.DigestAlgorithm, expect.Digest),
		verify.WithCertificateIdentity(identity),
	))
	if err != nil {
		return Provenance{}, fmt.Errorf("releaseupdate: Sigstore cryptographic verification failed: %w", err)
	}
	if result.Statement == nil || result.Statement.GetType() != statementV1 {
		return Provenance{}, errors.New("releaseupdate: attestation is not an in-toto v1 statement")
	}
	if result.Statement.GetPredicateType() != expect.PredicateType {
		return Provenance{}, fmt.Errorf(
			"releaseupdate: predicate type %q, want %q",
			result.Statement.GetPredicateType(), expect.PredicateType)
	}
	if len(result.VerifiedTimestamps) == 0 {
		return Provenance{}, errors.New("releaseupdate: no verified transparency timestamp")
	}
	if !subjectMatches(result.Statement.GetSubject(), expect) {
		return Provenance{}, errors.New("releaseupdate: statement subject does not bind the selected asset digest")
	}

	predicate := result.Statement.GetPredicate()
	if predicate == nil {
		return Provenance{}, errors.New("releaseupdate: SLSA predicate is missing")
	}
	rootMap := predicate.AsMap()
	workflow, ok := nestedMap(rootMap, "buildDefinition", "externalParameters", "workflow")
	if !ok ||
		stringValue(workflow["repository"]) != expect.Repository ||
		stringValue(workflow["path"]) != expect.WorkflowPath ||
		stringValue(workflow["ref"]) != expect.Ref {
		return Provenance{}, errors.New("releaseupdate: SLSA workflow repository/path/ref mismatch")
	}
	buildDefinition, ok := nestedMap(rootMap, "buildDefinition")
	if !ok {
		return Provenance{}, errors.New("releaseupdate: SLSA build definition is missing")
	}
	dependencies, ok := buildDefinition["resolvedDependencies"].([]any)
	if !ok {
		return Provenance{}, errors.New("releaseupdate: SLSA resolved dependencies are missing")
	}
	expectedURI := "git+" + expect.Repository + "@" + expect.Ref
	commit := ""
	for _, raw := range dependencies {
		dependency, ok := raw.(map[string]any)
		if !ok || stringValue(dependency["uri"]) != expectedURI {
			continue
		}
		digests, ok := dependency["digest"].(map[string]any)
		if ok {
			commit = strings.ToLower(stringValue(digests["gitCommit"]))
		}
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(commit) {
		return Provenance{}, errors.New("releaseupdate: SLSA source commit is missing or invalid")
	}
	return Provenance{WorkflowIdentity: expect.SAN, SourceCommit: commit}, nil
}

func subjectMatches(subjects []*in_toto.ResourceDescriptor, expect bundleExpectation) bool {
	wantDigest := strings.ToLower(hex.EncodeToString(expect.Digest))
	for _, subject := range subjects {
		if subject == nil ||
			strings.ToLower(subject.GetDigest()[expect.DigestAlgorithm]) != wantDigest {
			continue
		}
		if expect.AssetName == "" || path.Base(subject.GetName()) == expect.AssetName {
			return true
		}
	}
	return false
}

func nestedMap(root map[string]any, keys ...string) (map[string]any, bool) {
	current := root
	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func ensurePrivateCache(cacheDir string) error {
	cacheDir = filepath.Clean(strings.TrimSpace(cacheDir))
	if cacheDir == "." || !filepath.IsAbs(cacheDir) {
		return errors.New("releaseupdate: Sigstore cache path must be absolute")
	}
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return fmt.Errorf("releaseupdate: creating Sigstore cache: %w", err)
	}
	info, err := os.Lstat(cacheDir)
	if err != nil {
		return fmt.Errorf("releaseupdate: inspecting Sigstore cache: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("releaseupdate: Sigstore cache is not a non-symlink directory")
	}
	if err := os.Chmod(cacheDir, 0o700); err != nil {
		return fmt.Errorf("releaseupdate: protecting Sigstore cache: %w", err)
	}
	return nil
}

func verifyPinnedTUFRoot() error {
	sum := sha256.Sum256(tuf.DefaultRoot())
	got := hex.EncodeToString(sum[:])
	if got != pinnedTUFRootSHA256 {
		return fmt.Errorf(
			"releaseupdate: embedded Sigstore TUF root digest %s, want pinned %s",
			got, pinnedTUFRootSHA256)
	}
	return nil
}

func versionRepo() string {
	return version.Repo
}
