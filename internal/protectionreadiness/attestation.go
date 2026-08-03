package protectionreadiness

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

type attestationBody struct {
	SchemaVersion      string           `json:"schema_version"`
	Serial             uint64           `json:"serial"`
	KeyID              string           `json:"key_id"`
	SignatureAlgorithm string           `json:"signature_algorithm"`
	AccountID          string           `json:"account_id"`
	ProfileID          string           `json:"profile_id"`
	Market             Market           `json:"market"`
	OrderType          string           `json:"order_type"`
	SessionScope       string           `json:"session_scope"`
	QuantityMin        uint64           `json:"quantity_min"`
	QuantityMax        uint64           `json:"quantity_max"`
	TriggerSource      string           `json:"trigger_source"`
	ReplaceSemantics   string           `json:"replace_semantics"`
	Broker             brokerCapability `json:"broker"`
	ToolDigest         string           `json:"tool_digest"`
	BuildDigest        string           `json:"build_digest"`
	EvidenceDigest     string           `json:"evidence_digest"`
	IssuedAt           string           `json:"issued_at"`
	ExpiresAt          string           `json:"expires_at"`
}

type attestationEnvelope struct {
	attestationBody
	Signature string `json:"signature"`
}

func canonicalAttestationBody(body attestationBody) ([]byte, error) {
	return json.Marshal(body)
}

type verifiedAttestation struct {
	body       attestationBody
	issuedAt   time.Time
	expiresAt  time.Time
	bodyDigest string
}

func verifyAttestation(policy pinnedTrustPolicy, state durableState, now time.Time, market Market, input marketAssessmentInput) (verifiedAttestation, RefusalCode) {
	if !validObservedFile(input.File, policy, market) {
		return verifiedAttestation{}, RefusalFileMetadata
	}
	envelope, code := decodeStrictEnvelope(input.File.bytes)
	if code != RefusalNone {
		return verifiedAttestation{}, code
	}
	body := envelope.attestationBody
	if body.SchemaVersion != SchemaVersionV1 {
		return verifiedAttestation{}, RefusalSchema
	}
	if body.SignatureAlgorithm != AlgorithmEd25519 || len(policy.allowedAlgorithms) != 1 || policy.allowedAlgorithms[0] != AlgorithmEd25519 {
		return verifiedAttestation{}, RefusalAlgorithm
	}
	if !validAttestationFields(body) {
		return verifiedAttestation{}, RefusalInvalid
	}
	key, ok := policy.key(body.KeyID)
	if !ok {
		return verifiedAttestation{}, RefusalUnknownKey
	}
	if !key.revokedAt.IsZero() && !now.Before(key.revokedAt) {
		return verifiedAttestation{}, RefusalRevokedKey
	}
	issuedAt, issueOK := parseCanonicalTime(body.IssuedAt)
	expiresAt, expiryOK := parseCanonicalTime(body.ExpiresAt)
	if !issueOK || !expiryOK || !expiresAt.After(issuedAt) {
		return verifiedAttestation{}, RefusalInvalid
	}
	if issuedAt.Before(key.acceptFrom) || !issuedAt.Before(key.overlapUntil) || now.Before(key.acceptFrom) || !now.Before(key.overlapUntil) {
		return verifiedAttestation{}, RefusalRotationWindow
	}
	if expiresAt.Sub(issuedAt) > policy.maximumLifetime {
		return verifiedAttestation{}, RefusalMaximumLifetime
	}
	if issuedAt.After(now) {
		return verifiedAttestation{}, RefusalIssuedInFuture
	}
	if !now.Before(expiresAt) {
		return verifiedAttestation{}, RefusalExpired
	}
	scope := serialScope{AccountID: body.AccountID, ProfileID: body.ProfileID, Market: body.Market}
	if body.Serial <= state.Serials[scope] {
		return verifiedAttestation{}, RefusalSerialRollback
	}
	canonical, err := canonicalAttestationBody(body)
	if err != nil {
		return verifiedAttestation{}, RefusalInvalid
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(envelope.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize || !ed25519.Verify(key.publicKey, canonical, signature) {
		return verifiedAttestation{}, RefusalSignature
	}
	if !validBrokerCapability(body.Broker) {
		return verifiedAttestation{}, RefusalBrokerCapabilityUnattested
	}
	if !scopeMatches(body, input.Scope) || body.Market != market {
		return verifiedAttestation{}, RefusalScopeMismatch
	}
	if !validSupervisorBinding(input.Supervisor, input.Scope) {
		return verifiedAttestation{}, RefusalSupervisorUnwired
	}
	digest := sha256.Sum256(canonical)
	return verifiedAttestation{body: body, issuedAt: issuedAt, expiresAt: expiresAt, bodyDigest: hexBytes(digest[:])}, RefusalNone
}

func validAttestationFields(body attestationBody) bool {
	return body.Serial > 0 && body.KeyID != "" && body.AccountID != "" && body.ProfileID != "" && validMarket(body.Market) &&
		body.OrderType != "" && body.SessionScope != "" && body.QuantityMin > 0 && body.QuantityMax >= body.QuantityMin && body.TriggerSource != "" &&
		(body.ReplaceSemantics == ReplaceAtomic || body.ReplaceSemantics == ReplaceContinuousCoverage) && validDigest(body.ToolDigest) && validDigest(body.BuildDigest) && validDigest(body.EvidenceDigest)
}

func scopeMatches(body attestationBody, scope runtimeScope) bool {
	return body.AccountID == scope.AccountID && body.ProfileID == scope.ProfileID && body.Market == scope.Market && body.OrderType == scope.OrderType &&
		body.SessionScope == scope.SessionScope && scope.Quantity >= body.QuantityMin && scope.Quantity <= body.QuantityMax && body.TriggerSource == scope.TriggerSource &&
		body.ReplaceSemantics == scope.ReplaceSemantics && body.Broker == scope.Broker && body.ToolDigest == scope.ToolDigest && body.BuildDigest == scope.BuildDigest &&
		body.EvidenceDigest == scope.EvidenceDigest
}

func parseCanonicalTime(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Location() != time.UTC || parsed.UTC().Format(time.RFC3339Nano) != value {
		return time.Time{}, false
	}
	return parsed, true
}

func decodeStrictEnvelope(data []byte) (attestationEnvelope, RefusalCode) {
	duplicate, err := containsDuplicateJSONKey(data)
	if err != nil {
		return attestationEnvelope{}, RefusalInvalid
	}
	if duplicate {
		return attestationEnvelope{}, RefusalDuplicateField
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var envelope attestationEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		if strings.Contains(err.Error(), "unknown field") {
			return attestationEnvelope{}, RefusalUnknownField
		}
		return attestationEnvelope{}, RefusalInvalid
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return attestationEnvelope{}, RefusalInvalid
	}
	canonical, err := json.Marshal(envelope)
	if err != nil || !bytes.Equal(canonical, data) {
		return attestationEnvelope{}, RefusalNonCanonical
	}
	return envelope, RefusalNone
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("protectionreadiness: trailing JSON value")
		}
		return err
	}
	return nil
}

func containsDuplicateJSONKey(data []byte) (bool, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	duplicate, err := scanJSONValue(decoder)
	if err != nil {
		return false, err
	}
	if duplicate {
		return true, nil
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return false, err
	}
	return duplicate, nil
}

func scanJSONValue(decoder *json.Decoder) (bool, error) {
	token, err := decoder.Token()
	if err != nil {
		return false, err
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return false, nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]bool)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return false, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return false, errors.New("protectionreadiness: object key is not a string")
			}
			if seen[key] {
				return true, nil
			}
			seen[key] = true
			duplicate, err := scanJSONValue(decoder)
			if err != nil || duplicate {
				return duplicate, err
			}
		}
		end, err := decoder.Token()
		return false, validateClosingDelimiter(end, '}', err)
	case '[':
		for decoder.More() {
			duplicate, err := scanJSONValue(decoder)
			if err != nil || duplicate {
				return duplicate, err
			}
		}
		end, err := decoder.Token()
		return false, validateClosingDelimiter(end, ']', err)
	default:
		return false, errors.New("protectionreadiness: unexpected JSON delimiter")
	}
}

func validateClosingDelimiter(token json.Token, want json.Delim, err error) error {
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != want {
		return errors.New("protectionreadiness: invalid JSON closing delimiter")
	}
	return nil
}

func hexBytes(data []byte) string {
	const alphabet = "0123456789abcdef"
	encoded := make([]byte, len(data)*2)
	for index, value := range data {
		encoded[index*2] = alphabet[value>>4]
		encoded[index*2+1] = alphabet[value&15]
	}
	return string(encoded)
}
