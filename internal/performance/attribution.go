package performance

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"time"
)

var (
	ErrMarketRequired       = errors.New("performance: market is required for ticker attribution")
	ErrDivergentFillReplay  = errors.New("performance: fill identity has divergent evidence")
	ErrCorrectionLineage    = errors.New("performance: correction does not match the original fill lineage")
	ErrCorrectionDelta      = errors.New("performance: correction must be an explicit signed negative delta")
	ErrCorrectionOverrun    = errors.New("performance: correction exceeds the referenced fill")
	ErrQuantityConservation = errors.New("performance: acquired quantity does not equal closed plus residual quantity")
	ErrBasisConservation    = errors.New("performance: entry basis does not equal allocated plus residual basis")
	ErrPolicyMismatch       = errors.New("performance: fill cost policy differs from authoritative position policy")
	ErrLineageConflict      = errors.New("performance: fill lineage differs from authoritative attribution scope")
	ErrDuplicateAttribution = errors.New("performance: duplicate attribution key")
)

type FillKind string

const (
	FillKindEntry FillKind = "ENTRY"
	FillKindClose FillKind = "CLOSE"
)

type RoundingMode string

const (
	RoundHalfEven         RoundingMode = "HALF_EVEN"
	RoundHalfAwayFromZero RoundingMode = "HALF_AWAY_FROM_ZERO"
)

// CompositeLineage is persisted identity, not an inferred symbol/time join.
// Ticker is display metadata and is deliberately excluded from completeness.
type CompositeLineage struct {
	Market        string
	Ticker        string
	CandidateID   string
	LaneID        string
	LaneVersion   string
	CampaignID    string
	LegID         string
	DecisionID    string
	AttemptID     string
	OrderID       string
	FillID        string
	PositionID    string
	CloseID       string
	CloseLegID    string
	PolicyID      string
	PolicyVersion string
}

func (l CompositeLineage) missing(kind FillKind) []string {
	fields := []struct {
		name  string
		value string
	}{
		{"market", l.Market}, {"candidate_id", l.CandidateID}, {"lane_id", l.LaneID},
		{"lane_version", l.LaneVersion}, {"campaign_id", l.CampaignID}, {"leg_id", l.LegID},
		{"decision_id", l.DecisionID}, {"attempt_id", l.AttemptID}, {"order_id", l.OrderID},
		{"fill_id", l.FillID}, {"position_id", l.PositionID}, {"policy_id", l.PolicyID},
		{"policy_version", l.PolicyVersion},
	}
	if kind == FillKindClose {
		fields = append(fields, struct {
			name  string
			value string
		}{"close_id", l.CloseID}, struct {
			name  string
			value string
		}{"close_leg_id", l.CloseLegID})
	}
	missing := make([]string, 0)
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			missing = append(missing, field.name)
		}
	}
	sort.Strings(missing)
	return missing
}

type CostPolicy struct {
	ID                string
	Version           string
	SourceCurrency    string
	ReportingCurrency string
	RoundingMode      RoundingMode
	RoundingScale     int
	RoundingVersion   string
}

type PositionEvidence struct {
	Market             string
	PositionID         string
	CandidateID        string
	LaneID             string
	LaneVersion        string
	CampaignID         string
	LegID              string
	AcquiredQuantity   string
	ResidualQuantity   string
	TotalEntryBasis    string
	ResidualEntryBasis string
	Policy             CostPolicy
}

type AmountEvidence struct {
	Value         string
	Source        string
	SourceVersion string
	ObservedAt    time.Time
}

type FXEvidence struct {
	Source          string
	SourceVersion   string
	Rate            string
	AsOf            time.Time
	QuoteCurrency   string
	RoundingVersion string
}

type FillDelta struct {
	EventID             string
	Kind                FillKind
	Lineage             CompositeLineage
	QuantityDelta       string
	CorrectionOfFillID  string
	AllocatedEntryBasis AmountEvidence
	ExitProceeds        AmountEvidence
	EntryFees           *AmountEvidence
	ExitFees            *AmountEvidence
	Taxes               *AmountEvidence
	FXCost              *AmountEvidence
	FX                  *FXEvidence
}

type AttributionKey struct {
	Market        string
	Ticker        string
	LaneID        string
	LaneVersion   string
	CampaignID    string
	LegID         string
	PositionID    string
	PolicyID      string
	PolicyVersion string
}

type AmountMetric struct {
	Status   Status
	Value    string
	Currency string
}

type PnLBreakdown struct {
	Currency         string
	GrossPnL         AmountMetric
	EntryFees        AmountMetric
	ExitFees         AmountMetric
	Taxes            AmountMetric
	FXCost           AmountMetric
	NetPnL           AmountMetric
	RoundingResidual AmountMetric
	FXSource         string
	FXSourceVersion  string
	FXAsOf           time.Time
	RoundingVersion  string
}

type CloseLegAttribution struct {
	Lineage          CompositeLineage
	CloseID          string
	CloseLegID       string
	FillID           string
	Quantity         string
	EntryBasis       string
	ExitProceeds     string
	Source           PnLBreakdown
	Reporting        PnLBreakdown
	EntryFeeEvidence *AmountEvidence
	ExitFeeEvidence  *AmountEvidence
	TaxEvidence      *AmountEvidence
	FXCostEvidence   *AmountEvidence
	FXEvidence       *FXEvidence
}

type Attribution struct {
	Key                 AttributionKey
	LineageStatus       Status
	MissingLineage      []string
	MissingMeasurements []string
	// ObservedLineage preserves exact identifiers supplied by a source even
	// when missing legs/fills prevent a numeric attribution projection.
	ObservedLineage     []CompositeLineage
	CostPolicy          CostPolicy
	EntryFills          []CompositeLineage
	AcquiredQuantity    string
	ClosedQuantity      string
	ResidualQuantity    string
	FullyClosed         bool
	TotalEntryBasis     string
	AllocatedEntryBasis string
	ResidualEntryBasis  string
	Source              PnLBreakdown
	Reporting           PnLBreakdown
	CloseLegs           []CloseLegAttribution
}

type AttributionQuery struct {
	Market             string
	Ticker             string
	LaneID             string
	LaneVersion        string
	CampaignID         string
	LegID              string
	IncludeLinkMissing bool
}

// DerivedAttributionStore is an immutable, rebuildable read model. It has no
// journal writer, broker, config, lane toggle, quote collector or LIVE control.
type DerivedAttributionStore struct {
	rows []Attribution
}

func (s DerivedAttributionStore) QueryAttribution(query AttributionQuery) ([]Attribution, error) {
	query.Market = strings.TrimSpace(query.Market)
	query.Ticker = strings.TrimSpace(query.Ticker)
	if query.Ticker != "" && query.Market == "" {
		return nil, ErrMarketRequired
	}
	out := make([]Attribution, 0)
	for _, row := range s.rows {
		if !query.IncludeLinkMissing && row.LineageStatus != StatusComplete ||
			query.Market != "" && row.Key.Market != query.Market ||
			query.Ticker != "" && row.Key.Ticker != query.Ticker ||
			strings.TrimSpace(query.LaneID) != "" && row.Key.LaneID != strings.TrimSpace(query.LaneID) ||
			strings.TrimSpace(query.LaneVersion) != "" && row.Key.LaneVersion != strings.TrimSpace(query.LaneVersion) ||
			strings.TrimSpace(query.CampaignID) != "" && row.Key.CampaignID != strings.TrimSpace(query.CampaignID) ||
			strings.TrimSpace(query.LegID) != "" && row.Key.LegID != strings.TrimSpace(query.LegID) {
			continue
		}
		out = append(out, cloneAttribution(row))
	}
	return out, nil
}

type positionAccumulator struct {
	position PositionEvidence
	events   []FillDelta
	missing  map[string]struct{}
}

func BuildDerivedAttributionStore(positions []PositionEvidence, input []FillDelta) (DerivedAttributionStore, error) {
	byPosition := make(map[string]*positionAccumulator, len(positions))
	byBasePosition := make(map[string][]*positionAccumulator, len(positions))
	for _, position := range positions {
		position = normalizePosition(position)
		if err := validatePosition(position); err != nil {
			return DerivedAttributionStore{}, err
		}
		key := positionKey(position.Market, position.PositionID, position.CampaignID, position.LegID)
		if _, exists := byPosition[key]; exists {
			return DerivedAttributionStore{}, fmt.Errorf("performance: duplicate authoritative position %s", key)
		}
		accumulator := &positionAccumulator{position: position, missing: make(map[string]struct{})}
		byPosition[key] = accumulator
		baseKey := positionKey(position.Market, position.PositionID)
		byBasePosition[baseKey] = append(byBasePosition[baseKey], accumulator)
	}

	events, err := deduplicateFillDeltas(input)
	if err != nil {
		return DerivedAttributionStore{}, err
	}
	originals := make(map[string]FillDelta)
	for _, event := range events {
		if strings.TrimSpace(event.CorrectionOfFillID) == "" {
			originals[originalFillKey(event.Lineage, event.Kind, event.Lineage.FillID)] = event
		}
	}
	if err := validateCorrections(events, originals); err != nil {
		return DerivedAttributionStore{}, err
	}
	for _, event := range events {
		key := positionKey(event.Lineage.Market, event.Lineage.PositionID, event.Lineage.CampaignID, event.Lineage.LegID)
		position := byPosition[key]
		if position == nil {
			candidates := byBasePosition[positionKey(event.Lineage.Market, event.Lineage.PositionID)]
			if len(candidates) == 1 {
				position = candidates[0]
			}
		}
		if position == nil {
			if event.CorrectionOfFillID != "" {
				return DerivedAttributionStore{}, ErrCorrectionLineage
			}
			return DerivedAttributionStore{}, fmt.Errorf("performance: fill %s has no exact authoritative position", event.Lineage.FillID)
		}
		if event.Lineage.PolicyID != position.position.Policy.ID || event.Lineage.PolicyVersion != position.position.Policy.Version {
			return DerivedAttributionStore{}, ErrPolicyMismatch
		}
		if !lineageMatchesPosition(event.Lineage, position.position) {
			return DerivedAttributionStore{}, ErrLineageConflict
		}
		if err := validateFillDelta(event, position.position.Policy); err != nil {
			return DerivedAttributionStore{}, err
		}
		for _, field := range event.Lineage.missing(event.Kind) {
			position.missing[field] = struct{}{}
		}
		position.events = append(position.events, cloneFillDelta(event))
	}

	keys := make([]string, 0, len(byPosition))
	for key := range byPosition {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	store := DerivedAttributionStore{rows: make([]Attribution, 0, len(keys))}
	for _, key := range keys {
		row, err := projectPosition(*byPosition[key])
		if err != nil {
			return DerivedAttributionStore{}, err
		}
		store.rows = append(store.rows, row)
	}
	return store, nil
}

func deduplicateFillDeltas(input []FillDelta) ([]FillDelta, error) {
	byEvent := make(map[string]FillDelta)
	byFill := make(map[string]FillDelta)
	out := make([]FillDelta, 0, len(input))
	for _, raw := range input {
		event := normalizeFillDelta(raw)
		if event.EventID == "" || event.Lineage.FillID == "" {
			// Missing fill identity remains link_missing, but EventID is still
			// required to prevent an unbounded anonymous replay.
			if event.EventID == "" {
				return nil, errors.New("performance: fill event id is required")
			}
		}
		eventKey := positionKey(event.Lineage.Market, event.Lineage.PositionID, event.Lineage.CampaignID, event.Lineage.LegID) + "\x00" + event.EventID
		if previous, exists := byEvent[eventKey]; exists {
			if !reflect.DeepEqual(previous, event) {
				return nil, ErrDivergentFillReplay
			}
			continue
		}
		byEvent[eventKey] = event
		fillKey := eventIdentityKey(event)
		if event.Lineage.FillID == "" {
			fillKey += "\x00event:" + event.EventID
		}
		if previous, exists := byFill[fillKey]; exists {
			left, right := previous, event
			left.EventID, right.EventID = "", ""
			if !reflect.DeepEqual(left, right) {
				return nil, ErrDivergentFillReplay
			}
			continue
		}
		byFill[fillKey] = event
		out = append(out, event)
	}
	return out, nil
}

func projectPosition(acc positionAccumulator) (Attribution, error) {
	acquired, closed := new(big.Rat), new(big.Rat)
	allocatedBasis := new(big.Rat)
	var template CompositeLineage
	var closes []FillDelta
	var entries []CompositeLineage
	for _, event := range acc.events {
		quantity, _ := decimal(event.QuantityDelta)
		if template.PositionID == "" {
			template = event.Lineage
		}
		switch event.Kind {
		case FillKindEntry:
			acquired.Add(acquired, quantity)
			entries = append(entries, event.Lineage)
		case FillKindClose:
			closed.Add(closed, quantity)
			basis, _ := decimal(event.AllocatedEntryBasis.Value)
			allocatedBasis.Add(allocatedBasis, basis)
			closes = append(closes, event)
		}
	}
	wantAcquired, _ := decimal(acc.position.AcquiredQuantity)
	wantResidual, _ := decimal(acc.position.ResidualQuantity)
	if acquired.Cmp(wantAcquired) != 0 || new(big.Rat).Add(closed, wantResidual).Cmp(wantAcquired) != 0 ||
		acquired.Sign() < 0 || closed.Sign() < 0 || wantResidual.Sign() < 0 {
		return Attribution{}, ErrQuantityConservation
	}
	wantBasis, _ := decimal(acc.position.TotalEntryBasis)
	residualBasis, _ := decimal(acc.position.ResidualEntryBasis)
	if new(big.Rat).Add(allocatedBasis, residualBasis).Cmp(wantBasis) != 0 ||
		allocatedBasis.Sign() < 0 || residualBasis.Sign() < 0 {
		return Attribution{}, ErrBasisConservation
	}

	missing := sortedKeys(acc.missing)
	status := StatusComplete
	if len(missing) != 0 {
		status = StatusLinkMissing
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].FillID < entries[j].FillID })
	row := Attribution{
		Key: AttributionKey{Market: template.Market, Ticker: template.Ticker, LaneID: template.LaneID,
			LaneVersion: template.LaneVersion, CampaignID: template.CampaignID, LegID: template.LegID,
			PositionID: acc.position.PositionID, PolicyID: acc.position.Policy.ID, PolicyVersion: acc.position.Policy.Version},
		LineageStatus: status, MissingLineage: missing, AcquiredQuantity: ratText(acquired),
		CostPolicy: acc.position.Policy, EntryFills: entries,
		ClosedQuantity: ratText(closed), ResidualQuantity: ratText(wantResidual), FullyClosed: wantResidual.Sign() == 0,
		TotalEntryBasis: ratText(wantBasis), AllocatedEntryBasis: ratText(allocatedBasis),
		ResidualEntryBasis: ratText(residualBasis),
	}
	row.Source, row.Reporting, row.CloseLegs, row.MissingMeasurements = projectPnL(closes, acc.position.Policy)
	return row, nil
}

type financialTotals struct {
	basis, proceeds, entryFees, exitFees, taxes, fxCost                                                      *big.Rat
	entryFeesComplete, exitFeesComplete, taxesComplete, fxCostComplete                                       bool
	reportBasis, reportProceeds, reportEntryFees, reportExitFees, reportTaxes, reportFXCost, reportDirectNet *big.Rat
	reportComplete                                                                                           bool
	fxSource, fxSourceVersion                                                                                string
	fxAsOf                                                                                                   time.Time
}

func projectPnL(closes []FillDelta, policy CostPolicy) (PnLBreakdown, PnLBreakdown, []CloseLegAttribution, []string) {
	if len(closes) == 0 {
		return notMeasuredBreakdown(policy.SourceCurrency), notMeasuredBreakdown(policy.ReportingCurrency), nil,
			[]string{"close_fill"}
	}
	totals := financialTotals{
		basis: new(big.Rat), proceeds: new(big.Rat), entryFees: new(big.Rat), exitFees: new(big.Rat),
		taxes: new(big.Rat), fxCost: new(big.Rat), entryFeesComplete: true, exitFeesComplete: true,
		taxesComplete: true, fxCostComplete: true, reportBasis: new(big.Rat), reportProceeds: new(big.Rat),
		reportEntryFees: new(big.Rat), reportExitFees: new(big.Rat), reportTaxes: new(big.Rat),
		reportFXCost: new(big.Rat), reportDirectNet: new(big.Rat), reportComplete: true,
	}
	missingSet := make(map[string]struct{})
	legs := make([]CloseLegAttribution, 0, len(closes))
	for _, close := range closes {
		basis, _ := decimal(close.AllocatedEntryBasis.Value)
		proceeds, _ := decimal(close.ExitProceeds.Value)
		totals.basis.Add(totals.basis, basis)
		totals.proceeds.Add(totals.proceeds, proceeds)
		entryFee, entryOK := evidenceRat(close.EntryFees)
		exitFee, exitOK := evidenceRat(close.ExitFees)
		tax, taxOK := evidenceRat(close.Taxes)
		fxCost, fxCostOK := evidenceRat(close.FXCost)
		accumulateOptional(totals.entryFees, entryFee, entryOK, &totals.entryFeesComplete, missingSet, "entry_fees")
		accumulateOptional(totals.exitFees, exitFee, exitOK, &totals.exitFeesComplete, missingSet, "exit_fees")
		accumulateOptional(totals.taxes, tax, taxOK, &totals.taxesComplete, missingSet, "taxes")
		accumulateOptional(totals.fxCost, fxCost, fxCostOK, &totals.fxCostComplete, missingSet, "fx_cost")

		source := breakdown(policy.SourceCurrency, basis, proceeds, entryFee, entryOK, exitFee, exitOK, tax, taxOK, fxCost, fxCostOK)
		reporting := notMeasuredBreakdown(policy.ReportingCurrency)
		if close.FX == nil {
			totals.reportComplete = false
			missingSet["fx"] = struct{}{}
		} else {
			rate, _ := decimal(close.FX.Rate)
			if totals.fxSource == "" {
				totals.fxSource, totals.fxSourceVersion = close.FX.Source, close.FX.SourceVersion
			} else if totals.fxSource != close.FX.Source || totals.fxSourceVersion != close.FX.SourceVersion {
				totals.fxSource, totals.fxSourceVersion = "multiple-persisted", "multiple-persisted"
			}
			if close.FX.AsOf.After(totals.fxAsOf) {
				totals.fxAsOf = close.FX.AsOf.UTC()
			}
			addConverted(totals.reportBasis, basis, rate)
			addConverted(totals.reportProceeds, proceeds, rate)
			if entryOK {
				addConverted(totals.reportEntryFees, entryFee, rate)
			} else {
				totals.reportComplete = false
			}
			if exitOK {
				addConverted(totals.reportExitFees, exitFee, rate)
			} else {
				totals.reportComplete = false
			}
			if taxOK {
				addConverted(totals.reportTaxes, tax, rate)
			} else {
				totals.reportComplete = false
			}
			if fxCostOK {
				addConverted(totals.reportFXCost, fxCost, rate)
			} else {
				totals.reportComplete = false
			}
			if entryOK && exitOK && taxOK && fxCostOK {
				net := netPnL(basis, proceeds, entryFee, exitFee, tax, fxCost)
				addConverted(totals.reportDirectNet, net, rate)
				reporting = reportingBreakdown(policy, basis, proceeds, entryFee, exitFee, tax, fxCost, rate,
					close.FX.Source, close.FX.SourceVersion, close.FX.AsOf)
			}
		}
		legs = append(legs, CloseLegAttribution{Lineage: close.Lineage, CloseID: close.Lineage.CloseID, CloseLegID: close.Lineage.CloseLegID,
			FillID: close.Lineage.FillID, Quantity: close.QuantityDelta, EntryBasis: close.AllocatedEntryBasis.Value,
			ExitProceeds: close.ExitProceeds.Value, Source: source, Reporting: reporting,
			EntryFeeEvidence: cloneAmountPointer(close.EntryFees), ExitFeeEvidence: cloneAmountPointer(close.ExitFees),
			TaxEvidence: cloneAmountPointer(close.Taxes), FXCostEvidence: cloneAmountPointer(close.FXCost),
			FXEvidence: cloneFXPointer(close.FX)})
	}
	source := breakdown(policy.SourceCurrency, totals.basis, totals.proceeds, totals.entryFees, totals.entryFeesComplete,
		totals.exitFees, totals.exitFeesComplete, totals.taxes, totals.taxesComplete, totals.fxCost, totals.fxCostComplete)
	reporting := notMeasuredBreakdown(policy.ReportingCurrency)
	if totals.reportComplete {
		reporting = roundedReportingBreakdown(policy, totals)
	}
	sort.SliceStable(legs, func(i, j int) bool {
		if legs[i].CloseLegID == legs[j].CloseLegID {
			return legs[i].FillID < legs[j].FillID
		}
		return legs[i].CloseLegID < legs[j].CloseLegID
	})
	return source, reporting, legs, sortedKeys(missingSet)
}

func breakdown(currency string, basis, proceeds, entryFee *big.Rat, entryOK bool, exitFee *big.Rat, exitOK bool,
	tax *big.Rat, taxOK bool, fxCost *big.Rat, fxCostOK bool) PnLBreakdown {
	gross := new(big.Rat).Sub(proceeds, basis)
	out := PnLBreakdown{Currency: currency, GrossPnL: completeMetric(gross, currency),
		EntryFees: optionalMetric(entryFee, entryOK, currency), ExitFees: optionalMetric(exitFee, exitOK, currency),
		Taxes: optionalMetric(tax, taxOK, currency), FXCost: optionalMetric(fxCost, fxCostOK, currency),
		NetPnL: missingMetric(currency), RoundingResidual: missingMetric(currency)}
	if entryOK && exitOK && taxOK && fxCostOK {
		out.NetPnL = completeMetric(netPnL(basis, proceeds, entryFee, exitFee, tax, fxCost), currency)
	}
	return out
}

func reportingBreakdown(policy CostPolicy, basis, proceeds, entryFee, exitFee, tax, fxCost, rate *big.Rat,
	source, sourceVersion string, asOf time.Time) PnLBreakdown {
	totals := financialTotals{basis: basis, proceeds: proceeds, entryFees: entryFee, exitFees: exitFee, taxes: tax, fxCost: fxCost,
		reportBasis: multiplied(basis, rate), reportProceeds: multiplied(proceeds, rate),
		reportEntryFees: multiplied(entryFee, rate), reportExitFees: multiplied(exitFee, rate),
		reportTaxes: multiplied(tax, rate), reportFXCost: multiplied(fxCost, rate),
		reportDirectNet: multiplied(netPnL(basis, proceeds, entryFee, exitFee, tax, fxCost), rate),
		fxSource:        source, fxSourceVersion: sourceVersion, fxAsOf: asOf.UTC()}
	return roundedReportingBreakdown(policy, totals)
}

func roundedReportingBreakdown(policy CostPolicy, totals financialTotals) PnLBreakdown {
	gross := roundRat(new(big.Rat).Sub(totals.reportProceeds, totals.reportBasis), policy.RoundingScale, policy.RoundingMode)
	entry := roundRat(totals.reportEntryFees, policy.RoundingScale, policy.RoundingMode)
	exit := roundRat(totals.reportExitFees, policy.RoundingScale, policy.RoundingMode)
	tax := roundRat(totals.reportTaxes, policy.RoundingScale, policy.RoundingMode)
	fxCost := roundRat(totals.reportFXCost, policy.RoundingScale, policy.RoundingMode)
	net := netPnL(new(big.Rat), gross, entry, exit, tax, fxCost)
	directNet := roundRat(totals.reportDirectNet, policy.RoundingScale, policy.RoundingMode)
	residual := new(big.Rat).Sub(net, directNet)
	return PnLBreakdown{Currency: policy.ReportingCurrency, GrossPnL: completeMetric(gross, policy.ReportingCurrency),
		EntryFees: completeMetric(entry, policy.ReportingCurrency), ExitFees: completeMetric(exit, policy.ReportingCurrency),
		Taxes: completeMetric(tax, policy.ReportingCurrency), FXCost: completeMetric(fxCost, policy.ReportingCurrency),
		NetPnL: completeMetric(net, policy.ReportingCurrency), RoundingResidual: completeMetric(residual, policy.ReportingCurrency),
		FXSource: totals.fxSource, FXSourceVersion: totals.fxSourceVersion, FXAsOf: totals.fxAsOf.UTC(),
		RoundingVersion: policy.RoundingVersion}
}

func validatePosition(position PositionEvidence) error {
	if position.Market != "KR" && position.Market != "US" || position.PositionID == "" || position.CandidateID == "" ||
		position.LaneID == "" || position.LaneVersion == "" || position.CampaignID == "" || position.LegID == "" {
		return errors.New("performance: authoritative composite position scope is incomplete")
	}
	for name, value := range map[string]string{"acquired quantity": position.AcquiredQuantity,
		"residual quantity": position.ResidualQuantity, "total entry basis": position.TotalEntryBasis,
		"residual entry basis": position.ResidualEntryBasis} {
		parsed, ok := decimal(value)
		if !ok || parsed.Sign() < 0 {
			return fmt.Errorf("performance: %s is not a non-negative decimal", name)
		}
	}
	policy := position.Policy
	if policy.ID == "" || policy.Version == "" || policy.SourceCurrency == "" || policy.ReportingCurrency == "" ||
		policy.RoundingVersion == "" || policy.RoundingScale < 0 || policy.RoundingScale > 12 ||
		policy.RoundingMode != RoundHalfEven && policy.RoundingMode != RoundHalfAwayFromZero {
		return errors.New("performance: authoritative cost/rounding policy is incomplete")
	}
	return nil
}

func validateFillDelta(event FillDelta, policy CostPolicy) error {
	quantity, ok := decimal(event.QuantityDelta)
	if !ok || quantity.Sign() == 0 {
		return errors.New("performance: signed fill delta is required")
	}
	if event.Kind != FillKindEntry && event.Kind != FillKindClose {
		return errors.New("performance: fill kind is invalid")
	}
	if event.CorrectionOfFillID != "" && quantity.Sign() >= 0 {
		return ErrCorrectionDelta
	}
	if event.CorrectionOfFillID == "" && quantity.Sign() <= 0 {
		return ErrCorrectionDelta
	}
	if event.Kind == FillKindClose {
		correction := event.CorrectionOfFillID != ""
		if err := validateRequiredAmount(event.AllocatedEntryBasis, correction); err != nil {
			return err
		}
		if err := validateRequiredAmount(event.ExitProceeds, correction); err != nil {
			return err
		}
		for _, evidence := range []*AmountEvidence{event.EntryFees, event.ExitFees, event.Taxes, event.FXCost} {
			if evidence != nil {
				if err := validateOptionalAmount(*evidence, correction); err != nil {
					return err
				}
			}
		}
		if event.FX != nil {
			if event.FX.Source == "" || event.FX.SourceVersion == "" || event.FX.QuoteCurrency != policy.ReportingCurrency ||
				event.FX.RoundingVersion != policy.RoundingVersion || !validUTCTime(event.FX.AsOf) {
				return errors.New("performance: persisted FX provenance is incomplete or mismatched")
			}
			rate, ok := decimal(event.FX.Rate)
			if !ok || rate.Sign() <= 0 {
				return errors.New("performance: persisted FX rate is not positive")
			}
		}
	}
	return nil
}

func validateRequiredAmount(evidence AmountEvidence, correction bool) error {
	if evidence.Source == "" || evidence.SourceVersion == "" || !validUTCTime(evidence.ObservedAt) {
		return errors.New("performance: authoritative amount evidence is incomplete")
	}
	value, ok := decimal(evidence.Value)
	if !ok || correction && value.Sign() > 0 || !correction && value.Sign() < 0 {
		return errors.New("performance: authoritative amount sign is inconsistent with fill delta")
	}
	return nil
}

func validateOptionalAmount(evidence AmountEvidence, correction bool) error {
	return validateRequiredAmount(evidence, correction)
}

func sameCorrectionScope(original, correction FillDelta) bool {
	left, right := original.Lineage, correction.Lineage
	left.FillID, right.FillID = "", ""
	return original.Kind == correction.Kind && reflect.DeepEqual(left, right)
}

func lineageMatchesPosition(lineage CompositeLineage, position PositionEvidence) bool {
	return exactOrMissing(lineage.Market, position.Market) && exactOrMissing(lineage.PositionID, position.PositionID) &&
		exactOrMissing(lineage.CandidateID, position.CandidateID) && exactOrMissing(lineage.LaneID, position.LaneID) &&
		exactOrMissing(lineage.LaneVersion, position.LaneVersion) && exactOrMissing(lineage.CampaignID, position.CampaignID) &&
		exactOrMissing(lineage.LegID, position.LegID)
}

func exactOrMissing(observed, authoritative string) bool {
	return observed == "" || observed == authoritative
}

func validateCorrections(events []FillDelta, originals map[string]FillDelta) error {
	quantityCorrections := make(map[string]*big.Rat)
	amountCorrections := make(map[string]map[string]*big.Rat)
	for _, correction := range events {
		if correction.CorrectionOfFillID == "" {
			continue
		}
		quantity, ok := decimal(correction.QuantityDelta)
		if !ok || quantity.Sign() >= 0 {
			return ErrCorrectionDelta
		}
		key := originalFillKey(correction.Lineage, correction.Kind, correction.CorrectionOfFillID)
		original, exists := originals[key]
		if !exists || !sameCorrectionScope(original, correction) {
			return ErrCorrectionLineage
		}
		originalQuantity, _ := decimal(original.QuantityDelta)
		corrected := quantityCorrections[key]
		if corrected == nil {
			corrected = new(big.Rat)
			quantityCorrections[key] = corrected
		}
		corrected.Add(corrected, new(big.Rat).Neg(quantity))
		if corrected.Cmp(originalQuantity) > 0 {
			return ErrCorrectionOverrun
		}
		if correction.Kind == FillKindClose {
			pairs := []struct {
				name       string
				original   *AmountEvidence
				correction *AmountEvidence
			}{
				{"entry_basis", &original.AllocatedEntryBasis, &correction.AllocatedEntryBasis},
				{"exit_proceeds", &original.ExitProceeds, &correction.ExitProceeds},
				{"entry_fees", original.EntryFees, correction.EntryFees},
				{"exit_fees", original.ExitFees, correction.ExitFees},
				{"taxes", original.Taxes, correction.Taxes},
				{"fx_cost", original.FXCost, correction.FXCost},
			}
			if amountCorrections[key] == nil {
				amountCorrections[key] = make(map[string]*big.Rat)
			}
			for _, pair := range pairs {
				if pair.correction == nil {
					continue
				}
				if pair.original == nil {
					return ErrCorrectionLineage
				}
				originalValue, originalOK := decimal(pair.original.Value)
				correctionValue, correctionOK := decimal(pair.correction.Value)
				if !originalOK || !correctionOK || correctionValue.Sign() > 0 {
					return ErrCorrectionDelta
				}
				total := amountCorrections[key][pair.name]
				if total == nil {
					total = new(big.Rat)
					amountCorrections[key][pair.name] = total
				}
				total.Add(total, new(big.Rat).Neg(correctionValue))
				if total.Cmp(originalValue) > 0 {
					return ErrCorrectionOverrun
				}
			}
		}
	}
	return nil
}

func normalizePosition(input PositionEvidence) PositionEvidence {
	input.Market, input.PositionID = strings.TrimSpace(input.Market), strings.TrimSpace(input.PositionID)
	input.CandidateID, input.LaneID, input.LaneVersion = strings.TrimSpace(input.CandidateID), strings.TrimSpace(input.LaneID), strings.TrimSpace(input.LaneVersion)
	input.CampaignID, input.LegID = strings.TrimSpace(input.CampaignID), strings.TrimSpace(input.LegID)
	input.AcquiredQuantity, input.ResidualQuantity = strings.TrimSpace(input.AcquiredQuantity), strings.TrimSpace(input.ResidualQuantity)
	input.TotalEntryBasis, input.ResidualEntryBasis = strings.TrimSpace(input.TotalEntryBasis), strings.TrimSpace(input.ResidualEntryBasis)
	input.Policy.ID, input.Policy.Version = strings.TrimSpace(input.Policy.ID), strings.TrimSpace(input.Policy.Version)
	input.Policy.SourceCurrency, input.Policy.ReportingCurrency = strings.TrimSpace(input.Policy.SourceCurrency), strings.TrimSpace(input.Policy.ReportingCurrency)
	input.Policy.RoundingVersion = strings.TrimSpace(input.Policy.RoundingVersion)
	return input
}

func normalizeFillDelta(input FillDelta) FillDelta {
	input = cloneFillDelta(input)
	input.EventID, input.QuantityDelta, input.CorrectionOfFillID = strings.TrimSpace(input.EventID), strings.TrimSpace(input.QuantityDelta), strings.TrimSpace(input.CorrectionOfFillID)
	normalizeLineage(&input.Lineage)
	normalizeAmount(&input.AllocatedEntryBasis)
	normalizeAmount(&input.ExitProceeds)
	for _, value := range []*AmountEvidence{input.EntryFees, input.ExitFees, input.Taxes, input.FXCost} {
		if value != nil {
			normalizeAmount(value)
		}
	}
	if input.FX != nil {
		input.FX.Source, input.FX.SourceVersion, input.FX.Rate = strings.TrimSpace(input.FX.Source), strings.TrimSpace(input.FX.SourceVersion), strings.TrimSpace(input.FX.Rate)
		input.FX.QuoteCurrency, input.FX.RoundingVersion = strings.TrimSpace(input.FX.QuoteCurrency), strings.TrimSpace(input.FX.RoundingVersion)
	}
	return input
}

func normalizeLineage(lineage *CompositeLineage) {
	values := []*string{&lineage.Market, &lineage.Ticker, &lineage.CandidateID, &lineage.LaneID,
		&lineage.LaneVersion, &lineage.CampaignID, &lineage.LegID, &lineage.DecisionID, &lineage.AttemptID,
		&lineage.OrderID, &lineage.FillID, &lineage.PositionID, &lineage.CloseID, &lineage.CloseLegID,
		&lineage.PolicyID, &lineage.PolicyVersion}
	for _, value := range values {
		*value = strings.TrimSpace(*value)
	}
}

func normalizeAmount(value *AmountEvidence) {
	value.Value, value.Source, value.SourceVersion = strings.TrimSpace(value.Value), strings.TrimSpace(value.Source), strings.TrimSpace(value.SourceVersion)
}

func cloneFillDelta(input FillDelta) FillDelta {
	out := input
	if input.EntryFees != nil {
		copy := *input.EntryFees
		out.EntryFees = &copy
	}
	if input.ExitFees != nil {
		copy := *input.ExitFees
		out.ExitFees = &copy
	}
	if input.Taxes != nil {
		copy := *input.Taxes
		out.Taxes = &copy
	}
	if input.FXCost != nil {
		copy := *input.FXCost
		out.FXCost = &copy
	}
	if input.FX != nil {
		copy := *input.FX
		out.FX = &copy
	}
	return out
}

func cloneAmountPointer(input *AmountEvidence) *AmountEvidence {
	if input == nil {
		return nil
	}
	out := *input
	return &out
}

func cloneFXPointer(input *FXEvidence) *FXEvidence {
	if input == nil {
		return nil
	}
	out := *input
	return &out
}

func cloneAttribution(input Attribution) Attribution {
	out := input
	out.MissingLineage = append([]string(nil), input.MissingLineage...)
	out.MissingMeasurements = append([]string(nil), input.MissingMeasurements...)
	out.ObservedLineage = append([]CompositeLineage(nil), input.ObservedLineage...)
	out.EntryFills = append([]CompositeLineage(nil), input.EntryFills...)
	out.CloseLegs = make([]CloseLegAttribution, len(input.CloseLegs))
	for index, leg := range input.CloseLegs {
		out.CloseLegs[index] = leg
		out.CloseLegs[index].EntryFeeEvidence = cloneAmountPointer(leg.EntryFeeEvidence)
		out.CloseLegs[index].ExitFeeEvidence = cloneAmountPointer(leg.ExitFeeEvidence)
		out.CloseLegs[index].TaxEvidence = cloneAmountPointer(leg.TaxEvidence)
		out.CloseLegs[index].FXCostEvidence = cloneAmountPointer(leg.FXCostEvidence)
		out.CloseLegs[index].FXEvidence = cloneFXPointer(leg.FXEvidence)
	}
	return out
}

func eventIdentityKey(event FillDelta) string {
	return strings.Join([]string{event.Lineage.Market, event.Lineage.PositionID, event.Lineage.LaneID,
		event.Lineage.LaneVersion, event.Lineage.CampaignID, event.Lineage.LegID, string(event.Kind), event.Lineage.FillID}, "\x00")
}

func positionKey(parts ...string) string { return strings.Join(parts, "\x00") }

func originalFillKey(lineage CompositeLineage, kind FillKind, fillID string) string {
	return positionKey(lineage.Market, lineage.CandidateID, lineage.LaneID, lineage.LaneVersion,
		lineage.CampaignID, lineage.LegID, lineage.DecisionID, lineage.AttemptID, lineage.OrderID,
		lineage.PositionID, lineage.CloseID, lineage.CloseLegID, lineage.PolicyID, lineage.PolicyVersion,
		string(kind), fillID)
}

func evidenceRat(evidence *AmountEvidence) (*big.Rat, bool) {
	if evidence == nil {
		return new(big.Rat), false
	}
	value, ok := decimal(evidence.Value)
	return value, ok
}

func accumulateOptional(total, value *big.Rat, complete bool, allComplete *bool, missing map[string]struct{}, name string) {
	if !complete {
		*allComplete = false
		missing[name] = struct{}{}
		return
	}
	total.Add(total, value)
}

func addConverted(total, value, rate *big.Rat) { total.Add(total, multiplied(value, rate)) }
func multiplied(left, right *big.Rat) *big.Rat { return new(big.Rat).Mul(left, right) }

func netPnL(basis, proceeds, entryFee, exitFee, taxes, fxCost *big.Rat) *big.Rat {
	value := new(big.Rat).Sub(proceeds, basis)
	for _, cost := range []*big.Rat{entryFee, exitFee, taxes, fxCost} {
		value.Sub(value, cost)
	}
	return value
}

func completeMetric(value *big.Rat, currency string) AmountMetric {
	return AmountMetric{Status: StatusComplete, Value: ratText(value), Currency: currency}
}
func missingMetric(currency string) AmountMetric {
	return AmountMetric{Status: StatusNotMeasured, Currency: currency}
}
func optionalMetric(value *big.Rat, ok bool, currency string) AmountMetric {
	if !ok {
		return missingMetric(currency)
	}
	return completeMetric(value, currency)
}

func notMeasuredBreakdown(currency string) PnLBreakdown {
	missing := missingMetric(currency)
	return PnLBreakdown{Currency: currency, GrossPnL: missing, EntryFees: missing, ExitFees: missing,
		Taxes: missing, FXCost: missing, NetPnL: missing, RoundingResidual: missing}
}

func roundRat(value *big.Rat, scale int, mode RoundingMode) *big.Rat {
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	scaledNum := new(big.Int).Mul(value.Num(), factor)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(scaledNum, value.Denom(), remainder)
	absRemainder := new(big.Int).Abs(new(big.Int).Set(remainder))
	doubled := new(big.Int).Lsh(absRemainder, 1)
	comparison := doubled.Cmp(value.Denom())
	increment := comparison > 0 || comparison == 0 && (mode == RoundHalfAwayFromZero || quotient.Bit(0) == 1)
	if increment {
		if scaledNum.Sign() < 0 {
			quotient.Sub(quotient, big.NewInt(1))
		} else {
			quotient.Add(quotient, big.NewInt(1))
		}
	}
	return new(big.Rat).SetFrac(quotient, factor)
}

func subtractDecimals(gross string, costs ...string) (string, bool) {
	value, ok := decimal(gross)
	if !ok {
		return "", false
	}
	value = new(big.Rat).Set(value)
	for _, raw := range costs {
		cost, valid := decimal(raw)
		if !valid {
			return "", false
		}
		value.Sub(value, cost)
	}
	return ratText(value), true
}

func sortedKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func validUTCTime(value time.Time) bool { return !value.IsZero() && value.Location() == time.UTC }
