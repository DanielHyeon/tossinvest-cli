package releaseupdate

import (
	"encoding/hex"
	"testing"

	sigstoredata "github.com/sigstore/sigstore-go/pkg/testing/data"
)

func TestSigstorePolicyVerifiesCryptographyIdentityPredicateAndSourceCommit(t *testing.T) {
	entity := sigstoredata.Bundle(t, "sigstore.js@2.0.0-provenance.sigstore.json")
	digest, err := hex.DecodeString(
		"46d4e2f74c4877316640000a6fdf8a8b59f1e0847667973e9859f774dd31b8f" +
			"1e0937813b777fb66a2ac67d50540fe34640966eee9fc2ccca387082b4c85cd3c")
	if err != nil {
		t.Fatal(err)
	}
	trusted := sigstoredata.TrustedRoot(t, "public-good.json")
	got, err := verifySignedEntityWithTrustedRoot(entity, trusted, bundleExpectation{
		DigestAlgorithm: "sha512",
		Digest:          digest,
		Issuer:          "https://token.actions.githubusercontent.com",
		SAN:             "https://github.com/sigstore/sigstore-js/.github/workflows/release.yml@refs/heads/main",
		PredicateType:   slsaProvenanceV1,
		Repository:      "https://github.com/sigstore/sigstore-js",
		WorkflowPath:    ".github/workflows/release.yml",
		Ref:             "refs/heads/main",
	})
	if err != nil {
		t.Fatalf("verify bundle: %v", err)
	}
	if got.SourceCommit != "f0b49a04e5a62250e0f60fb128004a73110fe311" {
		t.Fatalf("source commit = %q", got.SourceCommit)
	}
}

func TestSigstorePolicyRejectsWrongIdentityDigestAndPredicate(t *testing.T) {
	entity := sigstoredata.Bundle(t, "sigstore.js@2.0.0-provenance.sigstore.json")
	trusted := sigstoredata.TrustedRoot(t, "public-good.json")
	base := bundleExpectation{
		DigestAlgorithm: "sha512",
		Digest:          make([]byte, 64),
		Issuer:          "https://token.actions.githubusercontent.com",
		SAN:             "https://github.com/sigstore/sigstore-js/.github/workflows/release.yml@refs/heads/main",
		PredicateType:   slsaProvenanceV1,
		Repository:      "https://github.com/sigstore/sigstore-js",
		WorkflowPath:    ".github/workflows/release.yml",
		Ref:             "refs/heads/main",
	}
	for _, tc := range []struct {
		name string
		edit func(*bundleExpectation)
	}{
		{"digest", func(e *bundleExpectation) {}},
		{"SAN", func(e *bundleExpectation) {
			e.SAN = "https://github.com/attacker/repo/.github/workflows/release.yml@refs/heads/main"
		}},
		{"predicate", func(e *bundleExpectation) { e.PredicateType = "https://example.invalid/predicate" }},
		{"repository", func(e *bundleExpectation) { e.Repository = "https://github.com/attacker/repo" }},
		{"ref", func(e *bundleExpectation) { e.Ref = "refs/tags/v9.9.9" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expect := base
			expect.Digest = append([]byte(nil), base.Digest...)
			tc.edit(&expect)
			if _, err := verifySignedEntityWithTrustedRoot(entity, trusted, expect); err == nil {
				t.Fatal("verification succeeded")
			}
		})
	}
}

func TestPinnedPublicGoodRootDigest(t *testing.T) {
	if err := verifyPinnedTUFRoot(); err != nil {
		t.Fatal(err)
	}
}
