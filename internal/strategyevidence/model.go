package strategyevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

type EvidenceKind string

const (
	KindTradability     EvidenceKind = "tradability"
	KindDisclosureRisk  EvidenceKind = "disclosure_risk"
	KindDisclosure      EvidenceKind = "disclosure"
	KindKRNetFlow       EvidenceKind = "kr_net_flow"
	KindUSParticipation EvidenceKind = "us_participation"
)

type SourceAuthority string

const (
	AuthorityOpenDART SourceAuthority = "opendart"
	AuthorityKRX      SourceAuthority = "krx"
	AuthoritySEC      SourceAuthority = "sec_edgar"
)

type Availability string

const (
	AvailabilityAvailable   Availability = "AVAILABLE"
	AvailabilityUnavailable Availability = "UNAVAILABLE"
)

type Confidence string

const (
	ConfidenceVerified   Confidence = "VERIFIED"
	ConfidenceUnverified Confidence = "UNVERIFIED"
)

type RefusalCode string

const (
	RefusalUnsupportedEvidence RefusalCode = "EVIDENCE_UNSUPPORTED"
	RefusalIdentityMismatch    RefusalCode = "EVIDENCE_IDENTITY_MISMATCH"
	RefusalSourceUnavailable   RefusalCode = "SOURCE_UNAVAILABLE"
	RefusalCurrencyUnresolved  RefusalCode = "EVIDENCE_CURRENCY_UNRESOLVED"
	RefusalUnitUnresolved      RefusalCode = "EVIDENCE_UNIT_UNRESOLVED"
	RefusalTimestampInvalid    RefusalCode = "EVIDENCE_TIMESTAMP_INVALID"
	RefusalPayloadInvalid      RefusalCode = "EVIDENCE_PAYLOAD_INVALID"
	RefusalEvidenceMissing     RefusalCode = "EVIDENCE_MISSING"
	RefusalEvidenceStale       RefusalCode = "EVIDENCE_STALE"
	RefusalEvidenceConflict    RefusalCode = "EVIDENCE_CONFLICT"
	RefusalFatalEvidence       RefusalCode = "FATAL_EVIDENCE"
)

type ValidationError struct {
	Code   RefusalCode
	Field  string
	Detail string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("strategy evidence: %s (%s): %s", e.Code, e.Field, e.Detail)
}

func invalid(code RefusalCode, field, detail string) error {
	return &ValidationError{Code: code, Field: field, Detail: detail}
}

// Header is the complete immutable provenance preimage supplied to NewEnvelope.
// NewEnvelope normalizes it and Envelope only returns copies.
type Header struct {
	EvidenceID                 string
	Market                     marketclock.Market
	Symbol                     string
	IssuerIdentity             string
	IssuerMappingVersion       string
	Kind                       EvidenceKind
	SchemaVersion              string
	Authority                  SourceAuthority
	SourceRecordID             string
	RevisionIdentity           string
	SupersedesRevisionIdentity string
	MarketEffectiveDate        string
	SourceEventAt              time.Time
	SourceAvailableAt          time.Time
	ObservedAt                 time.Time
	IngestedAt                 time.Time
	Currency                   string
	Unit                       string
	Availability               Availability
	Confidence                 Confidence
}

// Envelope is immutable after construction. Its payload and header are never
// exposed by reference, which prevents a caller from changing a stored digest.
type Envelope struct {
	header  Header
	payload []byte
	digest  string
}

// NewEnvelope validates provenance and canonicalizes a normalized evidence
// payload. Payloads are deliberately restricted to a top-level JSON object so
// every schema version has named fields and cannot silently change its root
// shape to an array or scalar.
func NewEnvelope(header Header, payload []byte) (Envelope, error) {
	h, err := normalizeHeader(header)
	if err != nil {
		return Envelope{}, err
	}
	canonical, err := canonicalJSON(payload)
	if err != nil {
		return Envelope{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	if err := validateTypedPayload(h.Kind, canonical); err != nil {
		return Envelope{}, invalid(RefusalPayloadInvalid, "payload", err.Error())
	}
	sum := sha256.Sum256(canonical)
	return Envelope{header: h, payload: canonical, digest: hex.EncodeToString(sum[:])}, nil
}

func (e Envelope) Header() Header { return e.header }

func (e Envelope) CanonicalPayload() []byte {
	return bytes.Clone(e.payload)
}

func (e Envelope) PayloadDigest() string { return e.digest }

func normalizeHeader(in Header) (Header, error) {
	out := in
	out.EvidenceID = strings.TrimSpace(in.EvidenceID)
	out.Symbol = strings.ToUpper(strings.TrimSpace(in.Symbol))
	out.IssuerIdentity = strings.TrimSpace(in.IssuerIdentity)
	out.IssuerMappingVersion = strings.TrimSpace(in.IssuerMappingVersion)
	out.SchemaVersion = strings.TrimSpace(in.SchemaVersion)
	out.SourceRecordID = strings.TrimSpace(in.SourceRecordID)
	out.RevisionIdentity = strings.TrimSpace(in.RevisionIdentity)
	out.SupersedesRevisionIdentity = strings.TrimSpace(in.SupersedesRevisionIdentity)
	out.MarketEffectiveDate = strings.TrimSpace(in.MarketEffectiveDate)
	out.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	out.Unit = strings.TrimSpace(in.Unit)
	switch {
	case out.Market != marketclock.MarketKR && out.Market != marketclock.MarketUS:
		return Header{}, invalid(RefusalIdentityMismatch, "market", "only kr and us are supported")
	case out.EvidenceID == "" || out.Symbol == "" || out.IssuerIdentity == "" || out.IssuerMappingVersion == "":
		return Header{}, invalid(RefusalIdentityMismatch, "identity", "evidence, symbol, issuer identity and mapping version are required")
	case out.SchemaVersion == "" || out.SourceRecordID == "" || out.RevisionIdentity == "":
		return Header{}, invalid(RefusalIdentityMismatch, "source_identity", "schema, source record and revision are required")
	case out.SupersedesRevisionIdentity == out.RevisionIdentity && out.RevisionIdentity != "":
		return Header{}, invalid(RefusalIdentityMismatch, "supersedes", "a revision cannot supersede itself")
	case !authorityValid(out.Authority):
		return Header{}, invalid(RefusalSourceUnavailable, "authority", "unknown source authority")
	case !authoritySupportsMarket(out.Authority, out.Market):
		return Header{}, invalid(RefusalIdentityMismatch, "authority", "source authority does not serve evidence market")
	case !kindSupportsMarket(out.Kind, out.Market):
		return Header{}, invalid(RefusalUnsupportedEvidence, "kind", "evidence kind does not exist for market")
	case out.SourceAvailableAt.IsZero():
		return Header{}, invalid(RefusalSourceUnavailable, "source_available_at", "source availability is required")
	case out.SourceEventAt.IsZero() || out.ObservedAt.IsZero() || out.IngestedAt.IsZero():
		return Header{}, invalid(RefusalTimestampInvalid, "timestamps", "source event, observed and ingested clocks are required")
	case out.SourceAvailableAt.Before(out.SourceEventAt) || out.ObservedAt.Before(out.SourceAvailableAt) || out.IngestedAt.Before(out.ObservedAt):
		return Header{}, invalid(RefusalTimestampInvalid, "timestamps", "source_event_at <= source_available_at <= observed_at <= ingested_at is required")
	case out.Currency == "":
		return Header{}, invalid(RefusalCurrencyUnresolved, "currency", "currency is required")
	case out.Unit == "":
		return Header{}, invalid(RefusalPayloadInvalid, "unit", "unit is required")
	case out.Availability != AvailabilityAvailable && out.Availability != AvailabilityUnavailable:
		return Header{}, invalid(RefusalPayloadInvalid, "availability", "unknown availability")
	case out.Confidence != ConfidenceVerified && out.Confidence != ConfidenceUnverified:
		return Header{}, invalid(RefusalPayloadInvalid, "confidence", "unknown confidence")
	}
	if _, err := time.Parse("2006-01-02", out.MarketEffectiveDate); err != nil {
		return Header{}, invalid(RefusalTimestampInvalid, "market_effective_date", "expected YYYY-MM-DD")
	}
	out.SourceEventAt = out.SourceEventAt.UTC()
	out.SourceAvailableAt = out.SourceAvailableAt.UTC()
	out.ObservedAt = out.ObservedAt.UTC()
	out.IngestedAt = out.IngestedAt.UTC()
	return out, nil
}

func authorityValid(authority SourceAuthority) bool {
	switch authority {
	case AuthorityOpenDART, AuthorityKRX, AuthoritySEC:
		return true
	default:
		return false
	}
}

func authoritySupportsMarket(authority SourceAuthority, market marketclock.Market) bool {
	switch authority {
	case AuthorityOpenDART, AuthorityKRX:
		return market == marketclock.MarketKR
	case AuthoritySEC:
		return market == marketclock.MarketUS
	default:
		return false
	}
}

func kindSupportsMarket(kind EvidenceKind, market marketclock.Market) bool {
	switch kind {
	case KindTradability, KindDisclosureRisk, KindDisclosure:
		return true
	case KindKRNetFlow:
		return market == marketclock.MarketKR
	case KindUSParticipation:
		return market == marketclock.MarketUS
	default:
		return false
	}
}

func canonicalJSON(input []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.UseNumber()
	value, err := decodeCanonicalValue(decoder)
	if err != nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok || object == nil {
		return nil, errors.New("top-level JSON object is required")
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("multiple JSON values")
		}
		return nil, err
	}
	return json.Marshal(object)
}

type canonicalNumber string

const (
	maxCanonicalNumberBytes = 1024
	maxCanonicalExponent    = int64(1_000_000)
)

func (n canonicalNumber) MarshalJSON() ([]byte, error) { return []byte(n), nil }

func decodeCanonicalValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch value := token.(type) {
	case json.Delim:
		switch value {
		case '{':
			object := make(map[string]any)
			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				if keyErr != nil {
					return nil, keyErr
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, errors.New("JSON object key is not a string")
				}
				if _, duplicate := object[key]; duplicate {
					return nil, fmt.Errorf("duplicate JSON key %q", key)
				}
				child, childErr := decodeCanonicalValue(decoder)
				if childErr != nil {
					return nil, childErr
				}
				object[key] = child
			}
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}
			return object, nil
		case '[':
			var array []any
			for decoder.More() {
				child, childErr := decodeCanonicalValue(decoder)
				if childErr != nil {
					return nil, childErr
				}
				array = append(array, child)
			}
			if _, err := decoder.Token(); err != nil {
				return nil, err
			}
			return array, nil
		default:
			return nil, errors.New("unexpected JSON delimiter")
		}
	case json.Number:
		return normalizeJSONNumber(string(value))
	default:
		return value, nil
	}
}

func normalizeJSONNumber(value string) (canonicalNumber, error) {
	if len(value) == 0 || len(value) > maxCanonicalNumberBytes {
		return "", errors.New("JSON number exceeds canonicalization bound")
	}
	index := 0
	negative := false
	if value[index] == '-' {
		negative = true
		index++
		if index == len(value) {
			return "", errors.New("invalid JSON number")
		}
	}
	integerStart := index
	for index < len(value) && value[index] >= '0' && value[index] <= '9' {
		index++
	}
	if integerStart == index || (value[integerStart] == '0' && index-integerStart > 1) {
		return "", errors.New("invalid JSON integer")
	}
	integerDigits := value[integerStart:index]
	fractionDigits := ""
	if index < len(value) && value[index] == '.' {
		index++
		fractionStart := index
		for index < len(value) && value[index] >= '0' && value[index] <= '9' {
			index++
		}
		if fractionStart == index {
			return "", errors.New("invalid JSON fraction")
		}
		fractionDigits = value[fractionStart:index]
	}
	exponent := int64(0)
	if index < len(value) && (value[index] == 'e' || value[index] == 'E') {
		index++
		exponentNegative := false
		if index < len(value) && (value[index] == '+' || value[index] == '-') {
			exponentNegative = value[index] == '-'
			index++
		}
		exponentStart := index
		for index < len(value) && value[index] >= '0' && value[index] <= '9' {
			digit := int64(value[index] - '0')
			if exponent > maxCanonicalExponent+maxCanonicalNumberBytes {
				return "", errors.New("JSON number exponent exceeds canonicalization bound")
			}
			exponent = exponent*10 + digit
			index++
		}
		if exponentStart == index {
			return "", errors.New("invalid JSON exponent")
		}
		if exponentNegative {
			exponent = -exponent
		}
	}
	if index != len(value) {
		return "", errors.New("invalid JSON number suffix")
	}
	if exponent < -maxCanonicalExponent || exponent > maxCanonicalExponent {
		return "", errors.New("JSON number exponent exceeds canonicalization bound")
	}
	digits := strings.TrimLeft(integerDigits+fractionDigits, "0")
	if digits == "" {
		return canonicalNumber("0"), nil
	}
	trailingZeros := len(digits) - len(strings.TrimRight(digits, "0"))
	if trailingZeros > 0 {
		digits = digits[:len(digits)-trailingZeros]
	}
	exponent += int64(trailingZeros - len(fractionDigits))
	if exponent < -maxCanonicalExponent || exponent > maxCanonicalExponent {
		return "", errors.New("JSON number exponent exceeds canonicalization bound")
	}
	if negative {
		digits = "-" + digits
	}
	if exponent == 0 {
		return canonicalNumber(digits), nil
	}
	return canonicalNumber(digits + "e" + strconv.FormatInt(exponent, 10)), nil
}

func validateTypedPayload(kind EvidenceKind, canonical []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil {
		return err
	}
	if err := rejectSecretFields(object); err != nil {
		return err
	}
	types := map[string]string{"blocked": "bool", "code": "string", "score_ppm": "number", "flow": "number", "value": "number", "market": "string"}
	for field, expected := range types {
		value, exists := object[field]
		if !exists {
			continue
		}
		valid := false
		switch expected {
		case "bool":
			_, valid = value.(bool)
		case "string":
			_, valid = value.(string)
		case "number":
			_, valid = value.(json.Number)
		}
		if !valid {
			return fmt.Errorf("%s payload field %q must be %s", kind, field, expected)
		}
	}
	return nil
}

func rejectSecretFields(value any) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), " ", "_"))
			if strings.Contains(normalized, "credential") || strings.Contains(normalized, "secret") || strings.Contains(normalized, "password") || strings.Contains(normalized, "token") || strings.Contains(normalized, "api_key") || normalized == "crtfc_key" || normalized == "authorization" {
				return fmt.Errorf("secret-like payload field %q is forbidden", key)
			}
			if err := rejectSecretFields(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range typed {
			if err := rejectSecretFields(child); err != nil {
				return err
			}
		}
	}
	return nil
}

// stamp uses a fixed-width UTC representation. SQLite compares these values as
// TEXT, so RFC3339Nano's omitted trailing fraction would give the wrong order
// for two instants inside the same second.
func stamp(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000000000Z") }

func parseStamp(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }
