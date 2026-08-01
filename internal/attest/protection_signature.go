package attest

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ProtectionTrustRootFileName      = "protection-attestation-trust-root.json"
	ProtectionVerifierPolicyFileName = "protection-verifier-policy.json"
	ProtectionTrustRootFormatVersion = 1
	ProtectionVerifierPolicyVersion  = 1
	ProtectionEnvelopeVersion        = 1
	ProtectionSignatureDomain        = "TossOS/protection-capability-attestation/ed25519/v1"
	ProtectionSignatureAlgorithm     = "Ed25519"
	ProtectionSignerRole             = "PROTECTION_ATTESTATION_SIGNER"

	maxProtectionTrustRootKeys = 16
	maxProtectionTrustRootSize = 1 << 20
)

var (
	ErrProtectionTrust     = errors.New("attest: protection trust root is absent or invalid")
	ErrProtectionSignature = errors.New("attest: protection attestation signature is invalid")
)

type ProtectionKeyStatus string

const (
	ProtectionKeyActive  ProtectionKeyStatus = "ACTIVE"
	ProtectionKeyRevoked ProtectionKeyStatus = "REVOKED"
)

// ProtectionTrustRoot is a verifier-only keyset. This package deliberately
// exposes no save, generation, download, or TOFU API for it.
type ProtectionTrustRoot struct {
	FormatVersion int                  `json:"format_version"`
	Generation    uint64               `json:"generation"`
	Keys          []ProtectionTrustKey `json:"keys"`
}

type protectionVerifierPolicy struct {
	FormatVersion   int    `json:"format_version"`
	Generation      uint64 `json:"generation"`
	AttestationPath string `json:"attestation_path"`
	TrustRootPath   string `json:"trust_root_path"`
	TrustRootOwner  uint32 `json:"trust_root_owner_uid"`
	TrustRootDigest string `json:"trust_root_digest"`
}

type protectionPolicySource struct {
	path  string
	owner fileIdentity
}

// ProtectionVerifier is a sealed verifier. Production code cannot inject a
// path, UID, digest or clock; until an audited canonical policy source is wired,
// the zero value fails closed and protection remains UNWIRED/OFF.
type ProtectionVerifier struct {
	source             protectionPolicySource
	clock              func() time.Time
	policyMu           sync.Mutex
	observedGeneration uint64
	observedRootDigest string
}

type ProtectionTrustKey struct {
	KeyID            string              `json:"key_id"`
	Role             string              `json:"role"`
	Algorithm        string              `json:"algorithm"`
	PublicKey        string              `json:"public_key"`
	NotBefore        time.Time           `json:"not_before"`
	NotAfter         time.Time           `json:"not_after"`
	Status           ProtectionKeyStatus `json:"status"`
	RevokedAt        *time.Time          `json:"revoked_at,omitempty"`
	RevocationReason string              `json:"revocation_reason,omitempty"`
}

type protectionSignedEnvelope struct {
	EnvelopeVersion int    `json:"envelope_version"`
	Domain          string `json:"domain"`
	Algorithm       string `json:"algorithm"`
	KeyID           string `json:"key_id"`
	Payload         string `json:"payload"`
	Signature       string `json:"signature"`
}

func (v *ProtectionVerifier) Verify(scope ProtectionScope, evidence map[string][]byte) (VerifiedProtectionCapability, error) {
	parsed, err := v.parse()
	if err != nil {
		return VerifiedProtectionCapability{}, err
	}
	return v.verifyParsed(parsed, scope, evidence, nil)
}

func (v *ProtectionVerifier) parse() (parsedProtectionCapability, error) {
	if v == nil || v.clock == nil || v.source.path == "" {
		return parsedProtectionCapability{}, fmt.Errorf("%w: canonical verifier policy source is not provisioned", ErrProtectionTrust)
	}
	ownerUID, ok := currentProtectionOwnerUID()
	if !ok {
		return parsedProtectionCapability{}, fmt.Errorf("%w: current owner cannot be determined", ErrProtectionFile)
	}
	policy, err := loadProtectionVerifierPolicy(v.source, nil)
	if err != nil {
		return parsedProtectionCapability{}, err
	}
	parsed, err := parseSignedProtectionCapability(v, policy, fileIdentity{UID: ownerUID})
	if err != nil {
		return parsedProtectionCapability{}, err
	}
	if err := v.acceptPolicy(policy); err != nil {
		return parsedProtectionCapability{}, err
	}
	return parsed, nil
}

func (v *ProtectionVerifier) verifyParsed(parsed parsedProtectionCapability, scope ProtectionScope, evidence map[string][]byte, afterEvidence func()) (VerifiedProtectionCapability, error) {
	if v == nil || parsed.verifier != v || v.clock == nil {
		return VerifiedProtectionCapability{}, fmt.Errorf("%w: parsed verifier identity is absent", ErrProtectionTrust)
	}
	matched, err := verifyProtectionMatrix(parsed, scope, evidence)
	if err != nil {
		return VerifiedProtectionCapability{}, err
	}
	if afterEvidence != nil {
		afterEvidence()
	}
	policy, err := loadProtectionVerifierPolicy(v.source, nil)
	if err != nil {
		return VerifiedProtectionCapability{}, err
	}
	if policy.Generation < parsed.policyGeneration || (policy.Generation == parsed.policyGeneration && policy.TrustRootDigest != parsed.rootDigest) {
		return VerifiedProtectionCapability{}, fmt.Errorf("%w: policy generation rolled back or digest changed without generation", ErrProtectionTrust)
	}
	root, err := loadProtectionTrustRoot(policy, nil)
	if err != nil {
		return VerifiedProtectionCapability{}, err
	}
	key, ok := root.key(parsed.envelope.KeyID)
	if !ok {
		return VerifiedProtectionCapability{}, fmt.Errorf("%w: unknown current key id %q", ErrProtectionTrust, parsed.envelope.KeyID)
	}
	if err := verifyProtectionEnvelopeSignature(parsed.envelope, key); err != nil {
		return VerifiedProtectionCapability{}, err
	}
	now := v.clock()
	if parsed.matrix.IssuedAt.After(now) {
		return VerifiedProtectionCapability{}, fmt.Errorf("%w: issued_at is in the future", ErrProtectionInvalid)
	}
	if !now.Before(parsed.matrix.ExpiresAt) {
		return VerifiedProtectionCapability{}, fmt.Errorf("%w: expired at %s", ErrProtectionExpired, parsed.matrix.ExpiresAt.UTC().Format(time.RFC3339))
	}
	if err := key.verifyWindow(now, parsed.matrix.IssuedAt, parsed.matrix.ExpiresAt); err != nil {
		return VerifiedProtectionCapability{}, err
	}
	if err := v.acceptPolicy(policy); err != nil {
		return VerifiedProtectionCapability{}, err
	}
	return verifiedProtectionCapability(scope, matched), nil
}

func (v *ProtectionVerifier) acceptPolicy(policy protectionVerifierPolicy) error {
	v.policyMu.Lock()
	defer v.policyMu.Unlock()

	if policy.Generation < v.observedGeneration ||
		(policy.Generation == v.observedGeneration && v.observedRootDigest != "" && policy.TrustRootDigest != v.observedRootDigest) {
		return fmt.Errorf("%w: policy generation rolled back or digest changed without generation", ErrProtectionTrust)
	}
	if policy.Generation > v.observedGeneration || v.observedRootDigest == "" {
		v.observedGeneration = policy.Generation
		v.observedRootDigest = policy.TrustRootDigest
	}
	return nil
}

func parseSignedProtectionCapability(verifier *ProtectionVerifier, policy protectionVerifierPolicy, owner fileIdentity) (parsedProtectionCapability, error) {
	data, err := readProtectionArtifact(policy.AttestationPath, ProtectionFileName, owner, 0o600, maxProtectionFileSize, checkProtectionParentInfo, checkProtectionFileInfo, ErrProtectionFile, nil)
	if err != nil {
		return parsedProtectionCapability{}, err
	}

	var envelope protectionSignedEnvelope
	if err := decodeCanonicalProtectionJSON(data, &envelope); err != nil {
		return parsedProtectionCapability{}, fmt.Errorf("%w: envelope: %v", ErrProtectionInvalid, err)
	}
	if envelope.EnvelopeVersion != ProtectionEnvelopeVersion || envelope.Domain != ProtectionSignatureDomain || envelope.Algorithm != ProtectionSignatureAlgorithm || !validProtectionKeyID(envelope.KeyID) {
		return parsedProtectionCapability{}, fmt.Errorf("%w: unsupported envelope header", ErrProtectionSignature)
	}
	payload, err := decodeCanonicalBase64URL(envelope.Payload, 1, maxProtectionFileSize)
	if err != nil {
		return parsedProtectionCapability{}, fmt.Errorf("%w: payload encoding: %v", ErrProtectionInvalid, err)
	}

	matrix, err := decodeProtectionMatrix(payload)
	if err != nil {
		return parsedProtectionCapability{}, err
	}
	canonicalPayload, err := json.Marshal(matrix)
	if err != nil || !bytes.Equal(payload, canonicalPayload) {
		return parsedProtectionCapability{}, fmt.Errorf("%w: payload is not canonical JSON", ErrProtectionInvalid)
	}
	if err := matrix.validate(); err != nil {
		return parsedProtectionCapability{}, err
	}

	root, err := loadProtectionTrustRoot(policy, nil)
	if err != nil {
		return parsedProtectionCapability{}, err
	}
	key, ok := root.key(envelope.KeyID)
	if !ok {
		return parsedProtectionCapability{}, fmt.Errorf("%w: unknown key id %q", ErrProtectionTrust, envelope.KeyID)
	}
	if err := verifyProtectionEnvelopeSignature(envelope, key); err != nil {
		return parsedProtectionCapability{}, err
	}
	return parsedProtectionCapability{matrix: matrix, envelope: envelope, policyGeneration: policy.Generation, rootDigest: policy.TrustRootDigest, verifier: verifier}, nil
}

func verifyProtectionEnvelopeSignature(envelope protectionSignedEnvelope, key ProtectionTrustKey) error {
	signature, err := decodeCanonicalBase64URL(envelope.Signature, ed25519.SignatureSize, ed25519.SignatureSize)
	if err != nil {
		return fmt.Errorf("%w: signature encoding: %v", ErrProtectionSignature, err)
	}
	publicKey, err := decodeCanonicalBase64URL(key.PublicKey, ed25519.PublicKeySize, ed25519.PublicKeySize)
	if err != nil {
		return fmt.Errorf("%w: key %q: %v", ErrProtectionTrust, key.KeyID, err)
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKey), protectionSigningMessage(envelope), signature) {
		return ErrProtectionSignature
	}
	return nil
}

func loadProtectionVerifierPolicy(source protectionPolicySource, afterRead func()) (protectionVerifierPolicy, error) {
	if source.path == "" || !filepath.IsAbs(source.path) || filepath.Clean(source.path) != source.path {
		return protectionVerifierPolicy{}, fmt.Errorf("%w: canonical policy source path is absent", ErrProtectionTrust)
	}
	data, err := readProtectionArtifact(source.path, ProtectionVerifierPolicyFileName, source.owner, 0o444, maxProtectionTrustRootSize, checkProtectionTrustRootParentInfo, checkProtectionTrustRootFileInfo, ErrProtectionTrust, afterRead)
	if err != nil {
		return protectionVerifierPolicy{}, err
	}
	var policy protectionVerifierPolicy
	if err := decodeCanonicalProtectionJSON(data, &policy); err != nil {
		return protectionVerifierPolicy{}, fmt.Errorf("%w: policy: %v", ErrProtectionTrust, err)
	}
	if err := policy.validate(source.path); err != nil {
		return protectionVerifierPolicy{}, err
	}
	return policy, nil
}

func loadProtectionTrustRoot(policy protectionVerifierPolicy, afterRead func()) (ProtectionTrustRoot, error) {
	data, err := readProtectionArtifact(policy.TrustRootPath, ProtectionTrustRootFileName, fileIdentity{UID: policy.TrustRootOwner}, 0o444, maxProtectionTrustRootSize, checkProtectionTrustRootParentInfo, checkProtectionTrustRootFileInfo, ErrProtectionTrust, afterRead)
	if err != nil {
		return ProtectionTrustRoot{}, err
	}
	if protectionDigest(data) != policy.TrustRootDigest {
		return ProtectionTrustRoot{}, fmt.Errorf("%w: trust-root digest does not match current canonical policy", ErrProtectionTrust)
	}
	var root ProtectionTrustRoot
	if err := decodeCanonicalProtectionJSON(data, &root); err != nil {
		return ProtectionTrustRoot{}, fmt.Errorf("%w: %v", ErrProtectionTrust, err)
	}
	if err := root.validate(); err != nil {
		return ProtectionTrustRoot{}, err
	}
	if root.Generation != policy.Generation {
		return ProtectionTrustRoot{}, fmt.Errorf("%w: trust-root generation does not match policy", ErrProtectionTrust)
	}
	return root, nil
}

func readProtectionArtifact(path, basename string, owner fileIdentity, mode os.FileMode, maximum int64, checkParent, check func(os.FileInfo, fileIdentity) error, sentinel error, afterRead func()) ([]byte, error) {
	if filepath.Base(path) != basename {
		return nil, fmt.Errorf("%w: basename must be %s", sentinel, basename)
	}
	parent := filepath.Dir(path)
	parentInfo, err := os.Lstat(parent)
	if err != nil {
		return nil, fmt.Errorf("%w: inspecting parent %s: %v", sentinel, parent, err)
	}
	if err := checkParent(parentInfo, owner); err != nil {
		return nil, fmt.Errorf("%w: parent %s: %v", sentinel, parent, err)
	}
	before, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("%w: inspecting %s: %v", sentinel, path, err)
	}
	if before.Mode().Perm() != mode {
		return nil, fmt.Errorf("%w: %s mode is %04o, want %04o", sentinel, path, before.Mode().Perm(), mode)
	}
	if err := check(before, owner); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", sentinel, path, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: opening %s: %v", sentinel, path, err)
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("%w: stat opened %s: %v", sentinel, path, err)
	}
	if !os.SameFile(before, opened) || opened.Mode().Perm() != mode {
		return nil, fmt.Errorf("%w: %s changed while it was opened", sentinel, path)
	}
	if err := check(opened, owner); err != nil {
		return nil, fmt.Errorf("%w: opened %s: %v", sentinel, path, err)
	}

	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, fmt.Errorf("%w: reading %s: %v", sentinel, path, err)
	}
	if int64(len(data)) > maximum {
		return nil, fmt.Errorf("%w: %s exceeds %d bytes", sentinel, path, maximum)
	}
	if afterRead != nil {
		afterRead()
	}
	openedAfterRead, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("%w: restat opened %s: %v", sentinel, path, err)
	}
	pathAfterRead, err := os.Lstat(path)
	if err != nil || !sameProtectionSnapshot(opened, openedAfterRead) || !sameProtectionSnapshot(opened, pathAfterRead) {
		return nil, fmt.Errorf("%w: %s changed while it was read", sentinel, path)
	}
	parentAfterRead, err := os.Lstat(parent)
	if err != nil || !sameProtectionSnapshot(parentInfo, parentAfterRead) {
		return nil, fmt.Errorf("%w: parent %s changed while artifact was read", sentinel, parent)
	}
	if openedAfterRead.Mode().Perm() != mode || pathAfterRead.Mode().Perm() != mode {
		return nil, fmt.Errorf("%w: %s mode changed while it was read", sentinel, path)
	}
	if err := check(openedAfterRead, owner); err != nil {
		return nil, fmt.Errorf("%w: restat opened %s: %v", sentinel, path, err)
	}
	if err := check(pathAfterRead, owner); err != nil {
		return nil, fmt.Errorf("%w: restat %s: %v", sentinel, path, err)
	}
	if err := checkParent(parentAfterRead, owner); err != nil {
		return nil, fmt.Errorf("%w: restat parent %s: %v", sentinel, parent, err)
	}
	return data, nil
}

func checkProtectionTrustRootParentInfo(info os.FileInfo, owner fileIdentity) error {
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("not a real directory")
	}
	mode := info.Mode().Perm()
	if mode&0o022 != 0 || mode&0o001 == 0 {
		return fmt.Errorf("mode is %04o, require other-traverse and no group/other write", mode)
	}
	uid, ok := fileOwnerUID(info)
	if !ok || uid != owner.UID {
		return fmt.Errorf("owner uid is %d, want %d", uid, owner.UID)
	}
	return nil
}

func checkProtectionTrustRootFileInfo(info os.FileInfo, owner fileIdentity) error {
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("not a regular file")
	}
	if info.Mode().Perm() != 0o444 {
		return fmt.Errorf("mode is %04o, want 0444", info.Mode().Perm())
	}
	uid, ok := fileOwnerUID(info)
	if !ok {
		return errors.New("owner cannot be determined")
	}
	if uid != owner.UID {
		return fmt.Errorf("owner uid is %d, want %d", uid, owner.UID)
	}
	links, ok := fileLinkCount(info)
	if !ok || links != 1 {
		return fmt.Errorf("hard-link count is %d, want 1", links)
	}
	return nil
}

func decodeCanonicalProtectionJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("trailing JSON value")
		}
		return err
	}
	canonical, err := json.Marshal(target)
	if err != nil {
		return err
	}
	if !bytes.Equal(data, canonical) {
		return errors.New("JSON is not in its single canonical encoding")
	}
	return nil
}

func decodeCanonicalBase64URL(value string, minimum, maximum int) ([]byte, error) {
	if value == "" || strings.Contains(value, "=") {
		return nil, errors.New("empty or padded base64url")
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) < minimum || len(decoded) > maximum || base64.RawURLEncoding.EncodeToString(decoded) != value {
		return nil, errors.New("noncanonical base64url or invalid length")
	}
	return decoded, nil
}

func (r ProtectionTrustRoot) validate() error {
	if r.FormatVersion != ProtectionTrustRootFormatVersion {
		return fmt.Errorf("%w: trust-root format_version is %d", ErrProtectionTrust, r.FormatVersion)
	}
	if r.Generation == 0 {
		return fmt.Errorf("%w: trust-root generation is absent", ErrProtectionTrust)
	}
	if len(r.Keys) == 0 || len(r.Keys) > maxProtectionTrustRootKeys {
		return fmt.Errorf("%w: trust root has %d keys", ErrProtectionTrust, len(r.Keys))
	}
	seenID := make(map[string]bool, len(r.Keys))
	seenPublic := make(map[string]bool, len(r.Keys))
	previousID := ""
	for i := range r.Keys {
		key := r.Keys[i]
		if err := key.validate(); err != nil {
			return fmt.Errorf("%w: key[%d]: %v", ErrProtectionTrust, i, err)
		}
		if seenID[key.KeyID] || seenPublic[key.PublicKey] || (i > 0 && key.KeyID <= previousID) {
			return fmt.Errorf("%w: duplicate or unsorted key id/public key at key[%d]", ErrProtectionTrust, i)
		}
		seenID[key.KeyID], seenPublic[key.PublicKey] = true, true
		previousID = key.KeyID
	}
	return nil
}

func (r ProtectionTrustRoot) key(keyID string) (ProtectionTrustKey, bool) {
	for i := range r.Keys {
		if r.Keys[i].KeyID == keyID {
			return r.Keys[i], true
		}
	}
	return ProtectionTrustKey{}, false
}

func (k ProtectionTrustKey) validate() error {
	if !validProtectionKeyID(k.KeyID) {
		return errors.New("invalid key_id")
	}
	if k.Role != ProtectionSignerRole || k.Algorithm != ProtectionSignatureAlgorithm {
		return errors.New("unknown signer role or algorithm")
	}
	if _, err := decodeCanonicalBase64URL(k.PublicKey, ed25519.PublicKeySize, ed25519.PublicKeySize); err != nil {
		return fmt.Errorf("public key: %v", err)
	}
	if k.NotBefore.IsZero() || k.NotAfter.IsZero() || !k.NotAfter.After(k.NotBefore) {
		return errors.New("invalid key validity window")
	}
	if k.NotBefore.Location() != time.UTC || k.NotAfter.Location() != time.UTC || (k.RevokedAt != nil && k.RevokedAt.Location() != time.UTC) {
		return errors.New("key timestamps must use exact UTC")
	}
	switch k.Status {
	case ProtectionKeyActive:
		if k.RevokedAt != nil || k.RevocationReason != "" {
			return errors.New("active key contains revocation metadata")
		}
	case ProtectionKeyRevoked:
		if k.RevokedAt == nil || k.RevokedAt.IsZero() || strings.TrimSpace(k.RevocationReason) == "" || k.RevocationReason != strings.TrimSpace(k.RevocationReason) || k.RevokedAt.Before(k.NotBefore) || k.RevokedAt.After(k.NotAfter) {
			return errors.New("revoked key lacks canonical in-window revocation metadata")
		}
	default:
		return errors.New("unknown key status")
	}
	return nil
}

func (k ProtectionTrustKey) verifyWindow(now, issuedAt, expiresAt time.Time) error {
	if k.Status != ProtectionKeyActive {
		return fmt.Errorf("%w: key %q is hard-revoked", ErrProtectionTrust, k.KeyID)
	}
	if now.Before(k.NotBefore) || !now.Before(k.NotAfter) {
		return fmt.Errorf("%w: key %q is not active at verification time", ErrProtectionTrust, k.KeyID)
	}
	if issuedAt.Before(k.NotBefore) || expiresAt.After(k.NotAfter) {
		return fmt.Errorf("%w: attestation window escapes key %q validity", ErrProtectionTrust, k.KeyID)
	}
	return nil
}

func validProtectionKeyID(value string) bool {
	if len(value) < 1 || len(value) > 64 || value != strings.TrimSpace(value) {
		return false
	}
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '.' || ch == '_' || ch == '-') {
			return false
		}
	}
	return true
}

func (p protectionVerifierPolicy) validate(policyPath string) error {
	if p.FormatVersion != ProtectionVerifierPolicyVersion || p.Generation == 0 {
		return fmt.Errorf("%w: verifier policy version/generation is invalid", ErrProtectionTrust)
	}
	if p.AttestationPath == "" || !filepath.IsAbs(p.AttestationPath) || filepath.Clean(p.AttestationPath) != p.AttestationPath || filepath.Base(p.AttestationPath) != ProtectionFileName {
		return fmt.Errorf("%w: attestation path is not canonical", ErrProtectionTrust)
	}
	if p.TrustRootPath == "" || !filepath.IsAbs(p.TrustRootPath) || filepath.Clean(p.TrustRootPath) != p.TrustRootPath || filepath.Base(p.TrustRootPath) != ProtectionTrustRootFileName {
		return fmt.Errorf("%w: trust-root path is not canonical", ErrProtectionTrust)
	}
	if !validSHA256(p.TrustRootDigest) {
		return fmt.Errorf("%w: trust-root digest is not pinned", ErrProtectionTrust)
	}
	parents := map[string]bool{}
	for _, path := range []string{policyPath, p.AttestationPath, p.TrustRootPath} {
		parent := filepath.Clean(filepath.Dir(path))
		if parents[parent] {
			return fmt.Errorf("%w: policy, trust root and attestation require separate directories", ErrProtectionTrust)
		}
		parents[parent] = true
	}
	return nil
}

// protectionSigningMessage uses a fixed external domain prefix and
// length-delimited fields. The envelope's domain, algorithm, key ID and exact
// canonical payload are all authenticated, with no algorithm negotiation.
func protectionSigningMessage(envelope protectionSignedEnvelope) []byte {
	message := make([]byte, 0, len(envelope.Payload)+256)
	message = append(message, ProtectionSignatureDomain...)
	message = append(message, 0)
	for _, field := range []string{
		strconv.Itoa(envelope.EnvelopeVersion),
		envelope.Domain,
		envelope.Algorithm,
		envelope.KeyID,
		envelope.Payload,
	} {
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(field)))
		message = append(message, length[:]...)
		message = append(message, field...)
	}
	return message
}
