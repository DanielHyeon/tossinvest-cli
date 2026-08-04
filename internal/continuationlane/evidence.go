package continuationlane

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"
)

const maxContinuationThresholdPPM int64 = 1_000_000

const (
	UnitNotionalMinor = "NOTIONAL_MINOR"
	UnitShares        = "SHARES"
	UnitQuoteMinor    = "QUOTE_MINOR"
)

type EvidenceEnvelope struct {
	SchemaVersion  string `json:"schema_version"`
	Market         Market `json:"market"`
	Symbol         string `json:"symbol"`
	SourceRecordID string `json:"source_record_id"`
	SourceDigest   string `json:"source_digest"`
	EffectiveAt    string `json:"effective_at"`
	ObservedAt     string `json:"observed_at"`
	IngestedAt     string `json:"ingested_at"`
	EvaluatedAt    string `json:"evaluated_at"`
	FreshUntil     string `json:"fresh_until"`
	ThresholdSetID string `json:"threshold_set_id"`
	ConfigDigest   string `json:"config_digest"`
}

type KREvidence struct {
	Envelope              EvidenceEnvelope `json:"envelope"`
	NotionalUnit          string           `json:"notional_unit"`
	NetFlowNotionalMinor  string           `json:"net_flow_notional_minor"`
	TurnoverNotionalMinor string           `json:"turnover_notional_minor"`
	FlowPressurePPM       int64            `json:"flow_pressure_ppm"`
}

type USEvidence struct {
	Envelope                  EvidenceEnvelope `json:"envelope"`
	VolumeUnit                string           `json:"volume_unit"`
	PriceUnit                 string           `json:"price_unit"`
	ParticipatingVolumeShares string           `json:"participating_volume_shares"`
	BaselineVolumeShares      string           `json:"baseline_volume_shares"`
	ReferencePriceMinor       string           `json:"reference_price_minor"`
	LastPriceMinor            string           `json:"last_price_minor"`
	ParticipationPPM          int64            `json:"participation_ppm"`
	PriceChangePPM            int64            `json:"price_change_ppm"`
}

type KRFlowConfig struct {
	SchemaVersion          string
	ThresholdSetID         string
	Digest                 string
	MinimumFlowPressurePPM int64
	seal                   [32]byte
}

type USParticipationConfig struct {
	SchemaVersion           string
	ThresholdSetID          string
	Digest                  string
	MinimumParticipationPPM int64
	MinimumPriceChangePPM   int64
	seal                    [32]byte
}

func NewKRFlowConfig(thresholdSetID, digest string, minimumFlowPressurePPM int64) (KRFlowConfig, error) {
	config := KRFlowConfig{SchemaVersion: KRFlowSchemaV1, ThresholdSetID: strings.TrimSpace(thresholdSetID), Digest: strings.TrimSpace(digest), MinimumFlowPressurePPM: minimumFlowPressurePPM}
	if config.ThresholdSetID == "" || config.Digest == "" || minimumFlowPressurePPM < 0 || minimumFlowPressurePPM > maxContinuationThresholdPPM {
		return KRFlowConfig{}, errors.New("continuation lanes: invalid KR threshold config")
	}
	config.seal = krConfigSeal(config)
	return config, nil
}

func NewUSParticipationConfig(thresholdSetID, digest string, minimumParticipationPPM, minimumPriceChangePPM int64) (USParticipationConfig, error) {
	config := USParticipationConfig{SchemaVersion: USParticipationSchemaV1, ThresholdSetID: strings.TrimSpace(thresholdSetID), Digest: strings.TrimSpace(digest), MinimumParticipationPPM: minimumParticipationPPM, MinimumPriceChangePPM: minimumPriceChangePPM}
	if config.ThresholdSetID == "" || config.Digest == "" || minimumParticipationPPM < 0 || minimumParticipationPPM > maxContinuationThresholdPPM || minimumPriceChangePPM < 0 || minimumPriceChangePPM > maxContinuationThresholdPPM {
		return USParticipationConfig{}, errors.New("continuation lanes: invalid US threshold config")
	}
	config.seal = usConfigSeal(config)
	return config, nil
}

type SignalResult struct {
	Accepted         bool
	Code             RefusalCode
	FlowPressurePPM  int64
	ParticipationPPM int64
	PriceChangePPM   int64
}

func DecodeKREvidence(data []byte) (KREvidence, error) {
	var evidence KREvidence
	return evidence, decodeStrict(data, &evidence)
}

func DecodeUSEvidence(data []byte) (USEvidence, error) {
	var evidence USEvidence
	return evidence, decodeStrict(data, &evidence)
}

func decodeStrict(data []byte, target any) error {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("continuation lane evidence contains multiple values")
		}
		return err
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	return scanJSONValue(decoder)
}

func scanJSONValue(decoder *json.Decoder) error {
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
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("continuation lane evidence contains a non-string object key")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("continuation lane evidence contains duplicate JSON key %q", key)
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	default:
		return errors.New("continuation lane evidence contains an invalid JSON delimiter")
	}
	_, err = decoder.Token()
	return err
}

func EvaluateKRFlow(evidence KREvidence, config KRFlowConfig) SignalResult {
	if config.seal != krConfigSeal(config) || config.SchemaVersion != KRFlowSchemaV1 {
		return SignalResult{Code: RefusalConfigInvalid}
	}
	if code := validateEnvelope(evidence.Envelope, KRFlowSchemaV1, MarketKR, config.SchemaVersion, config.ThresholdSetID, config.Digest); code != RefusalNone {
		return SignalResult{Code: code}
	}
	if evidence.NotionalUnit != UnitNotionalMinor {
		return SignalResult{Code: RefusalUnitMismatch}
	}
	metric, err := signedPPM(evidence.NetFlowNotionalMinor, evidence.TurnoverNotionalMinor)
	if err != nil {
		return SignalResult{Code: codeForArithmetic(err)}
	}
	if metric != evidence.FlowPressurePPM {
		return SignalResult{Code: RefusalArithmeticMismatch}
	}
	if metric < config.MinimumFlowPressurePPM {
		return SignalResult{Code: RefusalThresholdNotMet, FlowPressurePPM: metric}
	}
	return SignalResult{Accepted: true, FlowPressurePPM: metric}
}

func EvaluateUSParticipation(evidence USEvidence, config USParticipationConfig) SignalResult {
	if config.seal != usConfigSeal(config) || config.SchemaVersion != USParticipationSchemaV1 {
		return SignalResult{Code: RefusalConfigInvalid}
	}
	if code := validateEnvelope(evidence.Envelope, USParticipationSchemaV1, MarketUS, config.SchemaVersion, config.ThresholdSetID, config.Digest); code != RefusalNone {
		return SignalResult{Code: code}
	}
	if evidence.VolumeUnit != UnitShares || evidence.PriceUnit != UnitQuoteMinor {
		return SignalResult{Code: RefusalUnitMismatch}
	}
	participation, err := unsignedPPM(evidence.ParticipatingVolumeShares, evidence.BaselineVolumeShares)
	if err != nil {
		return SignalResult{Code: codeForArithmetic(err)}
	}
	reference, err := parseUnsigned(evidence.ReferencePriceMinor)
	if err != nil || reference.Sign() <= 0 {
		return SignalResult{Code: codeForArithmetic(err)}
	}
	last, err := parseUnsigned(evidence.LastPriceMinor)
	if err != nil || last.Sign() <= 0 {
		return SignalResult{Code: codeForArithmetic(err)}
	}
	delta := new(big.Int).Sub(last, reference)
	priceChange, err := signedPPM(delta.String(), reference.String())
	if err != nil {
		return SignalResult{Code: codeForArithmetic(err)}
	}
	if participation != evidence.ParticipationPPM || priceChange != evidence.PriceChangePPM {
		return SignalResult{Code: RefusalArithmeticMismatch}
	}
	result := SignalResult{ParticipationPPM: participation, PriceChangePPM: priceChange}
	if participation < config.MinimumParticipationPPM || priceChange < config.MinimumPriceChangePPM {
		result.Code = RefusalThresholdNotMet
		return result
	}
	result.Accepted = true
	return result
}

func krConfigSeal(config KRFlowConfig) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{config.SchemaVersion, config.ThresholdSetID, config.Digest,
		new(big.Int).SetInt64(config.MinimumFlowPressurePPM).String()}, "\x00")))
}

func usConfigSeal(config USParticipationConfig) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{config.SchemaVersion, config.ThresholdSetID, config.Digest,
		new(big.Int).SetInt64(config.MinimumParticipationPPM).String(), new(big.Int).SetInt64(config.MinimumPriceChangePPM).String()}, "\x00")))
}

func validateEnvelope(envelope EvidenceEnvelope, schema string, market Market, configSchema, thresholdSet, digest string) RefusalCode {
	if configSchema != schema || envelope.SchemaVersion != schema {
		return RefusalSchemaMismatch
	}
	if envelope.Market != market {
		return RefusalMarketMismatch
	}
	if strings.TrimSpace(envelope.Symbol) == "" || strings.TrimSpace(envelope.SourceRecordID) == "" || strings.TrimSpace(envelope.SourceDigest) == "" {
		return RefusalEvidenceInvalid
	}
	if thresholdSet == "" || digest == "" || envelope.ThresholdSetID != thresholdSet || envelope.ConfigDigest != digest {
		return RefusalDigestMismatch
	}
	times := make([]time.Time, 5)
	for i, raw := range []string{envelope.EffectiveAt, envelope.ObservedAt, envelope.IngestedAt, envelope.EvaluatedAt, envelope.FreshUntil} {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return RefusalInvalidTime
		}
		times[i] = parsed
	}
	if times[3].After(times[4]) {
		return RefusalStaleEvidence
	}
	for i := 1; i < len(times); i++ {
		if times[i-1].After(times[i]) {
			return RefusalInvalidTime
		}
	}
	return RefusalNone
}
