package weeklyvaluelane

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var ErrStrictSchema = errors.New("weekly value strict schema invalid")

func newDisclosureConfig(market Market, source DisclosureSource, schema, modelVersion, modelDigest, thresholdDigest string) DisclosureConfig {
	config := DisclosureConfig{Market: market, Source: source, SchemaVersion: strings.TrimSpace(schema), ModelVersion: strings.TrimSpace(modelVersion), ModelConfigDigest: strings.TrimSpace(modelDigest), ThresholdDigest: strings.TrimSpace(thresholdDigest)}
	config.seal = disclosureConfigSeal(config)
	return config
}

func DecodeKREvidence(body []byte) (DisclosureEvidence, error) {
	var evidence DisclosureEvidence
	if err := strictDecode(body, &evidence); err != nil {
		return DisclosureEvidence{}, fmt.Errorf("%w: %v", ErrStrictSchema, err)
	}
	evidence.seal = evidenceSnapshotSeal(evidence)
	return evidence, nil
}

func DecodeUSEvidence(body []byte) (DisclosureEvidence, error) {
	var evidence DisclosureEvidence
	if err := strictDecode(body, &evidence); err != nil {
		return DisclosureEvidence{}, fmt.Errorf("%w: %v", ErrStrictSchema, err)
	}
	evidence.seal = evidenceSnapshotSeal(evidence)
	return evidence, nil
}

func strictDecode(body []byte, target any) error {
	if len(body) == 0 || len(body) > 1<<20 {
		return errors.New("body outside strict bound")
	}
	if err := rejectDuplicateKeys(body); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func rejectDuplicateKeys(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := consumeValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func consumeValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("non-string object key")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate key %q", key)
			}
			seen[key] = struct{}{}
			if err := consumeValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := consumeValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return errors.New("unexpected delimiter")
	}
}

func EvaluateKREvidence(evidence DisclosureEvidence, config DisclosureConfig) EvidenceResult {
	return evaluateEvidence(evidence, config, MarketKR, SourceOpenDART, KRDisclosureSchemaV1, "KRW", uint32(0))
}

func EvaluateUSEvidence(evidence DisclosureEvidence, config DisclosureConfig) EvidenceResult {
	return evaluateEvidence(evidence, config, MarketUS, SourceEDGAR, USDisclosureSchemaV1, "USD", uint32(2))
}

func evaluateEvidence(evidence DisclosureEvidence, config DisclosureConfig, market Market, source DisclosureSource, schema, currency string, scale uint32) EvidenceResult {
	refuse := func(code RefusalCode) EvidenceResult { return EvidenceResult{Code: code} }
	if evidence.seal == ([32]byte{}) || evidence.seal != evidenceSnapshotSeal(evidence) {
		return refuse(RefusalStrictSchema)
	}
	if config.seal == ([32]byte{}) || config.seal != disclosureConfigSeal(config) || config.Market != market || config.Source != source || config.SchemaVersion != schema {
		return refuse(RefusalConfigMismatch)
	}
	if evidence.Market != market {
		return refuse(RefusalMarketMismatch)
	}
	if evidence.Source != source {
		return refuse(RefusalSourceMismatch)
	}
	if evidence.SchemaVersion != schema || evidence.ModelVersion != config.ModelVersion || evidence.ModelConfigDigest != config.ModelConfigDigest || evidence.ThresholdDigest != config.ThresholdDigest {
		return refuse(RefusalConfigMismatch)
	}
	identities := []string{evidence.Symbol, evidence.IssuerID, evidence.FilingID, evidence.ReportID, evidence.RevisionID, evidence.SupersededRevisionID, evidence.DilutionFactsDigest, evidence.ModelID, evidence.EvidenceDigest}
	for _, value := range identities {
		if !validBoundedIdentity(value) {
			return refuse(RefusalSchemaInvalid)
		}
	}
	if evidence.RevisionSequence == 0 || evidence.RevisionID == evidence.SupersededRevisionID || (evidence.RevisionSequence == 1 && evidence.SupersededRevisionID != "NONE") || (evidence.RevisionSequence > 1 && evidence.SupersededRevisionID == "NONE") {
		return refuse(RefusalRevisionInvalid)
	}
	for _, instant := range []time.Time{evidence.AsOf, evidence.ObservedAt, evidence.IngestedAt, evidence.CutoffAt, evidence.EvaluatedAt, evidence.FreshUntil, evidence.DilutionAsOf} {
		if instant.IsZero() {
			return refuse(RefusalPointInTime)
		}
	}
	if evidence.AsOf.After(evidence.ObservedAt) || evidence.ObservedAt.After(evidence.IngestedAt) || evidence.IngestedAt.After(evidence.CutoffAt) || evidence.CutoffAt.After(evidence.EvaluatedAt) {
		return refuse(RefusalPointInTime)
	}
	if evidence.EvaluatedAt.After(evidence.FreshUntil) {
		return refuse(RefusalEvidenceStale)
	}
	if evidence.DilutionAsOf.After(evidence.CutoffAt) || (evidence.DilutionStatus != "NONE" && evidence.DilutionStatus != "OBSERVED") {
		return refuse(RefusalPointInTime)
	}
	if evidence.Currency != currency || evidence.MonetaryUnit != "MINOR" || evidence.MonetaryScale != scale || evidence.SharesUnit != "SHARES" {
		return refuse(RefusalUnitInvalid)
	}
	if evidence.DilutedShares == 0 {
		return refuse(RefusalSchemaInvalid)
	}
	if len(evidence.FinancialInputs) == 0 || len(evidence.FinancialInputs) > 128 {
		return refuse(RefusalSchemaInvalid)
	}
	seen := map[string]bool{}
	for _, input := range evidence.FinancialInputs {
		if !validBoundedIdentity(input.Name) || seen[input.Name] || input.Unit != currency+"_MINOR" {
			return refuse(RefusalSchemaInvalid)
		}
		seen[input.Name] = true
		if _, ok := parseSigned(input.ValueMinor); !ok {
			return refuse(RefusalArithmeticInvalid)
		}
	}
	equity, equityOK := parseUnsigned(evidence.EquityValueMinor)
	fair, fairOK := parseUnsigned(evidence.FairValueMinor)
	if !equityOK || !fairOK || equity.Sign() <= 0 || fair.Sign() <= 0 {
		return refuse(RefusalArithmeticInvalid)
	}
	computed := new(big.Int).Quo(equity, new(big.Int).SetUint64(evidence.DilutedShares))
	if computed.BitLen() > maxArithmeticBits || computed.Cmp(fair) != 0 {
		return refuse(RefusalArithmeticMismatch)
	}
	digest := evidenceDecisionDigest(evidence)
	return EvidenceResult{Accepted: true, FairValueMinor: fair.String(), DecisionDigest: digest}
}

func disclosureConfigSeal(config DisclosureConfig) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{string(config.Market), string(config.Source), config.SchemaVersion, config.ModelVersion, config.ModelConfigDigest, config.ThresholdDigest}, "\x00")))
}

func evidenceDecisionDigest(evidence DisclosureEvidence) string {
	preimage := strings.Join([]string{"weekly-value-decision-v1", hex.EncodeToString(evidence.seal[:])}, "\x00")
	sum := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(sum[:])
}

func evidenceSnapshotSeal(evidence DisclosureEvidence) [32]byte {
	parts := []string{
		"weekly-value-evidence-v1", evidence.SchemaVersion, string(evidence.Market), string(evidence.Source), evidence.Symbol, evidence.IssuerID,
		evidence.FilingID, evidence.ReportID, evidence.RevisionID, evidence.SupersededRevisionID, strconv.FormatUint(evidence.RevisionSequence, 10),
		canonicalTime(evidence.AsOf), canonicalTime(evidence.ObservedAt), canonicalTime(evidence.IngestedAt), canonicalTime(evidence.CutoffAt),
		canonicalTime(evidence.EvaluatedAt), canonicalTime(evidence.FreshUntil), evidence.Currency, evidence.MonetaryUnit,
		strconv.FormatUint(uint64(evidence.MonetaryScale), 10), strconv.FormatUint(evidence.DilutedShares, 10), evidence.SharesUnit,
		evidence.DilutionStatus, evidence.DilutionFactsDigest, canonicalTime(evidence.DilutionAsOf),
	}
	for _, input := range evidence.FinancialInputs {
		parts = append(parts, input.Name, input.ValueMinor, input.Unit)
	}
	parts = append(parts, evidence.ModelID, evidence.ModelVersion, evidence.ModelConfigDigest, evidence.ThresholdDigest,
		evidence.EvidenceDigest, evidence.EquityValueMinor, evidence.FairValueMinor)
	return sha256.Sum256([]byte(strings.Join(parts, "\x00")))
}

func canonicalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
