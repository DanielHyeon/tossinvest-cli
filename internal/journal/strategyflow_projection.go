package journal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

const (
	strategyflowRiskBindingSchemaVersionV2 = "journal-strategyflow-risk-binding:v2"
	strategyflowRiskBindingSchemaVersion   = "journal-strategyflow-risk-binding:v3"
)

type StrategyflowLineageProjectionRequest struct {
	Result                   strategyflow.Result
	RiskIntent               RiskIntent
	ActivationManifestDigest string
	CreatedAt                time.Time
}

type strategyflowRiskBindingPayload struct {
	SchemaVersion             string          `json:"schema_version"`
	Strategyflow              json.RawMessage `json:"strategyflow"`
	StrategyflowPayloadDigest string          `json:"strategyflow_payload_digest"`
	RiskIntent                RiskIntent      `json:"risk_intent"`
	ActivationManifestDigest  string          `json:"activation_manifest_digest"`
	CreatedAt                 string          `json:"created_at"`
	Identity                  string          `json:"identity"`
}

// ProjectAcceptedStrategyflowLineage converts sealed accepted strategyflow
// evidence into the durable journal shape. It performs no write and grants no
// lease, activation, Gateway or broker authority.
func ProjectAcceptedStrategyflowLineage(request StrategyflowLineageProjectionRequest) (StrategyDecisionLineage, error) {
	projection, err := strategyflow.ProjectAccepted(request.Result)
	if err != nil {
		return StrategyDecisionLineage{}, fmt.Errorf("journal strategy issuance: strategyflow projection: %w", err)
	}
	risk, err := exactCanonicalRiskIntent(request.RiskIntent)
	if err != nil {
		return StrategyDecisionLineage{}, err
	}
	activation := request.ActivationManifestDigest
	if activation == "" || strings.TrimSpace(activation) != activation || request.CreatedAt.IsZero() {
		return StrategyDecisionLineage{}, errors.New("journal strategy issuance: strategyflow projection metadata invalid")
	}
	created := request.CreatedAt.UTC()
	if err := verifyProjectionRiskIntent(projection, risk, strategyflowRiskBindingSchemaVersion); err != nil {
		return StrategyDecisionLineage{}, err
	}
	record := strategyflowRiskBindingPayload{SchemaVersion: strategyflowRiskBindingSchemaVersion,
		Strategyflow: json.RawMessage(projection.Payload()), StrategyflowPayloadDigest: projection.PayloadDigest(), RiskIntent: risk,
		ActivationManifestDigest: activation, CreatedAt: created.Format(time.RFC3339Nano)}
	identity, err := strategyflowDecisionIdentity(record)
	if err != nil {
		return StrategyDecisionLineage{}, err
	}
	record.Identity = identity
	payload, err := json.Marshal(record)
	if err != nil {
		return StrategyDecisionLineage{}, fmt.Errorf("journal strategy issuance: strategyflow projection encoding: %w", err)
	}
	return strategyflowLineageFromProjection(projection, risk, activation, created, identity, string(payload), strategyflowRiskBindingSchemaVersion), nil
}

func exactCanonicalRiskIntent(value RiskIntent) (RiskIntent, error) {
	canonical, err := value.Canonical()
	if err != nil {
		return RiskIntent{}, fmt.Errorf("journal strategy issuance: strategyflow RiskIntent: %w", err)
	}
	parsed, err := ParsePreimage(PreimageKindRiskIntent, canonical)
	if err != nil {
		return RiskIntent{}, fmt.Errorf("journal strategy issuance: strategyflow RiskIntent parse: %w", err)
	}
	risk, ok := parsed.(RiskIntent)
	if !ok || value.AccountRef != risk.AccountRef || value.Market != strings.ToUpper(risk.Market) || value.Symbol != risk.Symbol ||
		value.Side != risk.Side || value.Quantity != risk.Quantity || value.EntryPrice != risk.EntryPrice || value.StopPrice != risk.StopPrice ||
		value.TargetPrice != risk.TargetPrice || value.PolicyVersion != risk.PolicyVersion {
		return RiskIntent{}, errors.New("journal strategy issuance: strategyflow RiskIntent canonical type mismatch")
	}
	return risk, nil
}

func verifyProjectionRiskIntent(projection strategyflow.AcceptedProjection, risk RiskIntent, schema string) error {
	lineage, terms := projection.Lineage(), projection.ExecutionTerms()
	quantity, canonical := NormalizeDecimal(risk.Quantity)
	qFinal, quantityErr := strconv.ParseUint(quantity, 10, 64)
	_, _, marked := splitQFinalPolicyVersion(risk.PolicyVersion)
	if !canonical || quantity != risk.Quantity || quantityErr != nil || qFinal == 0 || qFinal > terms.Quantity() || !marked ||
		risk.AccountRef != lineage.AccountRef || risk.AccountRef != terms.AccountRef() ||
		!strings.EqualFold(risk.Market, string(lineage.Market)) || !strings.EqualFold(risk.Market, string(terms.Market())) || risk.Symbol != lineage.Symbol || risk.Symbol != terms.Symbol() {
		return errors.New("journal strategy issuance: strategyflow RiskIntent/accepted-result exact binding mismatch")
	}
	entry, stop, target := terms.Entry().PriceMinor(), terms.EffectiveStop().PriceMinor(), terms.Target().PriceMinor()
	if schema == strategyflowRiskBindingSchemaVersion {
		var entryOK, stopOK, targetOK bool
		entry, entryOK = terms.Entry().MajorDecimal()
		stop, stopOK = terms.EffectiveStop().MajorDecimal()
		target, targetOK = terms.Target().MajorDecimal()
		if !entryOK || !stopOK || !targetOK {
			return errors.New("journal strategy issuance: strategyflow execution price unit invalid")
		}
	} else if schema != strategyflowRiskBindingSchemaVersionV2 {
		return errors.New("journal strategy issuance: strategyflow decision payload schema unsupported")
	}
	if risk.EntryPrice != entry || risk.StopPrice != stop || risk.TargetPrice != target {
		return errors.New("journal strategy issuance: strategyflow RiskIntent/accepted-result exact binding mismatch")
	}
	return nil
}

func strategyflowDecisionIdentity(record strategyflowRiskBindingPayload) (string, error) {
	record.Identity = ""
	payload, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("journal strategy issuance: strategyflow decision identity encoding: %w", err)
	}
	sum := sha256.Sum256(payload)
	version := "v3"
	if record.SchemaVersion == strategyflowRiskBindingSchemaVersionV2 {
		version = "v2"
	} else if record.SchemaVersion != strategyflowRiskBindingSchemaVersion {
		return "", errors.New("journal strategy issuance: strategyflow decision identity schema unsupported")
	}
	return "strategy-decision:" + version + ":sha256:" + hex.EncodeToString(sum[:]), nil
}

func strategyflowLineageFromProjection(projection strategyflow.AcceptedProjection, risk RiskIntent, activation string, created time.Time, identity, payload, schema string) StrategyDecisionLineage {
	lineage, terms := projection.Lineage(), projection.ExecutionTerms()
	payloadHash := sha256.Sum256([]byte(payload))
	entry, stop, target := terms.Entry().PriceMinor(), terms.EffectiveStop().PriceMinor(), terms.Target().PriceMinor()
	if schema == strategyflowRiskBindingSchemaVersion {
		entry, stop, target = risk.EntryPrice, risk.StopPrice, risk.TargetPrice
	}
	return StrategyDecisionLineage{DecisionIdentity: identity, CandidateLifeID: lineage.CandidateLifeID, Market: string(lineage.Market), Symbol: lineage.Symbol,
		ThresholdVersion: lineage.ThresholdVersion, ThresholdSetDigest: lineage.ThresholdSetDigest, EvidenceDigest: lineage.CandidateEvidenceDigest,
		LaneID: lineage.LaneID, LaneVersion: lineage.LaneVersion, LaneSourceDigest: lineage.LaneEvidenceDigest, LaneConstantsDigest: lineage.ConfigDigest,
		EntryPrice: entry, StopPrice: stop, TargetPrice: target,
		Quantity: risk.Quantity, PolicyVersion: risk.PolicyVersion, SettingsDigest: lineage.RiskBudgetDigest,
		DecisionPayload: payload, DecisionPayloadDigest: "sha256:" + hex.EncodeToString(payloadHash[:]), ActivationManifestDigest: activation, CreatedAt: created}
}

func verifyStrategyflowRiskBinding(risk RiskIntent, lineage StrategyDecisionLineage) error {
	record, err := decodeStrategyflowRiskBinding(lineage.DecisionPayload)
	if err != nil {
		return err
	}
	projection, err := strategyflow.VerifyAcceptedProjection(string(record.Strategyflow))
	if err != nil {
		return fmt.Errorf("journal strategy issuance: strategyflow accepted payload: %w", err)
	}
	if projection.PayloadDigest() != record.StrategyflowPayloadDigest || record.RiskIntent != risk {
		return errors.New("journal strategy issuance: strategyflow payload/RiskIntent mismatch")
	}
	if err := verifyProjectionRiskIntent(projection, risk, record.SchemaVersion); err != nil {
		return err
	}
	identity, err := strategyflowDecisionIdentity(record)
	if err != nil {
		return err
	}
	created, err := time.Parse(time.RFC3339Nano, record.CreatedAt)
	if err != nil || created.Location() != time.UTC || created.Format(time.RFC3339Nano) != record.CreatedAt {
		return errors.New("journal strategy issuance: strategyflow projection created_at invalid")
	}
	want := strategyflowLineageFromProjection(projection, risk, record.ActivationManifestDigest, created, identity, lineage.DecisionPayload, record.SchemaVersion)
	if record.Identity != identity || lineage != want {
		return errors.New("journal strategy issuance: strategyflow RiskIntent/decision exact binding mismatch")
	}
	return nil
}

func decodeStrategyflowRiskBinding(payload string) (strategyflowRiskBindingPayload, error) {
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	var record strategyflowRiskBindingPayload
	if err := decoder.Decode(&record); err != nil {
		return strategyflowRiskBindingPayload{}, fmt.Errorf("journal strategy issuance: decoding strategyflow decision payload: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return strategyflowRiskBindingPayload{}, errors.New("journal strategy issuance: strategyflow decision payload has trailing data")
	}
	canonical, err := json.Marshal(record)
	if err != nil || string(canonical) != payload {
		return strategyflowRiskBindingPayload{}, errors.New("journal strategy issuance: strategyflow decision payload is not canonical")
	}
	if record.SchemaVersion != strategyflowRiskBindingSchemaVersionV2 && record.SchemaVersion != strategyflowRiskBindingSchemaVersion {
		return strategyflowRiskBindingPayload{}, errors.New("journal strategy issuance: strategyflow decision payload schema unsupported")
	}
	if record.ActivationManifestDigest == "" || strings.TrimSpace(record.ActivationManifestDigest) != record.ActivationManifestDigest {
		return strategyflowRiskBindingPayload{}, errors.New("journal strategy issuance: strategyflow activation manifest binding invalid")
	}
	return record, nil
}
