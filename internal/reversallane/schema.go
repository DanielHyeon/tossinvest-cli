package reversallane

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"
)

const (
	maxStructuralWindow = 24 * time.Hour
	maxMetricPPM        = uint64(100_000_000)
)

func DecodeKREvidence(body []byte) (KREvidence, error) {
	var evidence KREvidence
	if err := strictDecode(body, &evidence); err != nil {
		return KREvidence{}, fmt.Errorf("%w: %v", ErrStrictSchema, err)
	}
	return evidence, nil
}

func DecodeUSEvidence(body []byte) (USEvidence, error) {
	var evidence USEvidence
	if err := strictDecode(body, &evidence); err != nil {
		return USEvidence{}, fmt.Errorf("%w: %v", ErrStrictSchema, err)
	}
	return evidence, nil
}

func strictDecode(body []byte, target any) error {
	if len(body) == 0 || len(body) > 1<<20 {
		return errors.New("schema body outside bound")
	}
	if err := rejectDuplicateJSONKeys(body); err != nil {
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

func rejectDuplicateJSONKeys(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := consumeJSONValue(decoder); err != nil {
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

func consumeJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("duplicate JSON key %q", key)
			}
			seen[key] = struct{}{}
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := consumeJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return errors.New("unexpected JSON delimiter")
	}
}

func EvaluateKRMetric(evidence KREvidence, config KRConfig) MetricResult {
	if strings.TrimSpace(config.Version) == "" || config.MinimumAbsorptionPPM == 0 || config.MinimumAbsorptionPPM > maxMetricPPM || config.StructuralWindow <= 0 || config.StructuralWindow > maxStructuralWindow {
		return MetricResult{Refusal: RefusalConfigMismatch}
	}
	if refusal := validateCommon(evidence.CommonEnvelope, MarketKR, config.SchemaVersion, config.ConfigDigest, config.ThresholdSet, config.StructuralWindow); refusal != "" {
		return MetricResult{Refusal: refusal}
	}
	if evidence.Units != (EvidenceUnits{Notional: "KRW_MINOR"}) {
		return MetricResult{Refusal: RefusalUnitInvalid}
	}
	ppm, ok := checkedPPM(evidence.AbsorbedNotionalMinor, evidence.AggressiveSellNotionalMinor)
	if !ok || ppm > maxMetricPPM || ppm != evidence.AbsorptionPPM {
		return MetricResult{Refusal: RefusalArithmeticInvalid}
	}
	if ppm < config.MinimumAbsorptionPPM {
		return MetricResult{Refusal: RefusalThresholdNotMet, AbsorptionPPM: ppm}
	}
	return MetricResult{Accepted: true, AbsorptionPPM: ppm}
}

func EvaluateUSMetric(evidence USEvidence, config USConfig) MetricResult {
	if strings.TrimSpace(config.Version) == "" || config.MinimumDrawdownPPM == 0 || config.MinimumDrawdownPPM > maxMetricPPM || config.MinimumRelativeVolumePPM == 0 || config.MinimumRelativeVolumePPM > maxMetricPPM || config.StructuralWindow <= 0 || config.StructuralWindow > maxStructuralWindow {
		return MetricResult{Refusal: RefusalConfigMismatch}
	}
	if refusal := validateCommon(evidence.CommonEnvelope, MarketUS, config.SchemaVersion, config.ConfigDigest, config.ThresholdSet, config.StructuralWindow); refusal != "" {
		return MetricResult{Refusal: refusal}
	}
	if evidence.Units != (EvidenceUnits{Price: "USD_MINOR", Volume: "SHARES"}) {
		return MetricResult{Refusal: RefusalUnitInvalid}
	}
	if evidence.ReferencePriceMinor == 0 || evidence.DislocationLowPriceMinor > evidence.ReferencePriceMinor || evidence.BaselineVolumeShares == 0 {
		return MetricResult{Refusal: RefusalArithmeticInvalid}
	}
	drawdown, ok := checkedPPM(evidence.ReferencePriceMinor-evidence.DislocationLowPriceMinor, evidence.ReferencePriceMinor)
	if !ok {
		return MetricResult{Refusal: RefusalArithmeticInvalid}
	}
	relative, ok := checkedPPM(evidence.DislocationVolumeShares, evidence.BaselineVolumeShares)
	if !ok || relative > maxMetricPPM || drawdown != evidence.DrawdownPPM || relative != evidence.RelativeVolumePPM {
		return MetricResult{Refusal: RefusalArithmeticInvalid}
	}
	result := MetricResult{DrawdownPPM: drawdown, RelativeVolumePPM: relative}
	if drawdown < config.MinimumDrawdownPPM || relative < config.MinimumRelativeVolumePPM {
		result.Refusal = RefusalThresholdNotMet
		return result
	}
	result.Accepted = true
	return result
}

func validateCommon(envelope CommonEnvelope, market Market, schemaVersion, configDigest, thresholdSet string, window time.Duration) RefusalCode {
	if envelope.Market != market {
		return RefusalScopeMismatch
	}
	if strings.TrimSpace(envelope.AccountRef) == "" || strings.TrimSpace(envelope.Symbol) == "" || envelope.PositionGeneration == 0 || strings.TrimSpace(envelope.SourceRecordID) == "" || strings.TrimSpace(envelope.SourceDigest) == "" {
		return RefusalStrictSchema
	}
	for _, value := range []string{envelope.SchemaVersion, envelope.AccountRef, envelope.Symbol, envelope.SourceRecordID, envelope.SourceDigest, envelope.ThresholdSet, envelope.ConfigDigest} {
		if len(value) > maxIdentityBytes {
			return RefusalStrictSchema
		}
	}
	if schemaVersion == "" || configDigest == "" || thresholdSet == "" || window < 0 || envelope.StructuralWindowNS < 0 || envelope.SchemaVersion != schemaVersion || envelope.ConfigDigest != configDigest || envelope.ThresholdSet != thresholdSet || envelope.StructuralWindowNS != int64(window) {
		return RefusalConfigMismatch
	}
	times := []time.Time{envelope.EffectiveAt, envelope.ObservedAt, envelope.IngestedAt, envelope.EvaluatedAt, envelope.FreshUntil}
	for _, instant := range times {
		if instant.IsZero() {
			return RefusalTimestampInvalid
		}
	}
	if envelope.EffectiveAt.After(envelope.ObservedAt) || envelope.ObservedAt.After(envelope.IngestedAt) || envelope.IngestedAt.After(envelope.EvaluatedAt) {
		return RefusalTimestampInvalid
	}
	if envelope.EvaluatedAt.After(envelope.FreshUntil) {
		return RefusalEvidenceStale
	}
	return ""
}

func checkedPPM(numerator, denominator uint64) (uint64, bool) {
	if denominator == 0 {
		return 0, false
	}
	product := new(big.Int).Mul(new(big.Int).SetUint64(numerator), big.NewInt(1_000_000))
	result := new(big.Int).Quo(product, new(big.Int).SetUint64(denominator))
	if !result.IsUint64() {
		return 0, false
	}
	return result.Uint64(), true
}
