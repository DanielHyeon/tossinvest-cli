package positioncampaign

import (
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

type ReplayReason string

const (
	ReplayOK                        ReplayReason = "OK"
	ReplaySequenceGap               ReplayReason = "SEQUENCE_GAP"
	ReplayDuplicateCommandKey       ReplayReason = "DUPLICATE_COMMAND_KEY"
	ReplayGenerationRetreat         ReplayReason = "POSITION_GENERATION_RETREAT"
	ReplaySnapshotDrift             ReplayReason = "SNAPSHOT_DRIFT"
	ReplayClosedReopened            ReplayReason = "CLOSED_REOPENED"
	ReplayGenerationRebound         ReplayReason = "POSITION_GENERATION_REBOUND"
	ReplayProspectiveRebound        ReplayReason = "PROSPECTIVE_TOKEN_REBOUND"
	ReplayLegSequenceGap            ReplayReason = "LEG_SEQUENCE_GAP"
	ReplayLegIdentityChanged        ReplayReason = "LEG_IDENTITY_CHANGED"
	ReplayOrphanOrderLineage        ReplayReason = "ORPHAN_ORDER_LINEAGE"
	ReplayOrderIdentityChanged      ReplayReason = "ORDER_IDENTITY_CHANGED"
	ReplayWatermarkRetreat          ReplayReason = "ORDER_WATERMARK_RETREAT"
	ReplayStopRetreat               ReplayReason = "STOP_RETREAT"
	ReplayInvalidEvidence           ReplayReason = "INVALID_EVIDENCE"
	ReplayCampaignVersionMismatch   ReplayReason = "CAMPAIGN_VERSION_MISMATCH"
	ReplayInvalidCampaignTransition ReplayReason = "INVALID_CAMPAIGN_TRANSITION"
	ReplayInvalidLegTransition      ReplayReason = "INVALID_LEG_TRANSITION"
	ReplayOrderDeltaMismatch        ReplayReason = "ORDER_DELTA_MISMATCH"
	ReplayOrderRemainingMismatch    ReplayReason = "ORDER_REMAINING_MISMATCH"
	ReplayLegQuantityMismatch       ReplayReason = "LEG_QUANTITY_MISMATCH"
)

type Event struct {
	Sequence                   int64
	CampaignVersion            int64
	EventKind                  string
	CommandKind                string
	CommandKey                 string
	RequestDigest              string
	CampaignState              CampaignState
	PositionGeneration         int64
	ExpectedPositionGeneration int64
	ProspectiveToken           string
	LegSequence                int64
	LegState                   LegState
	PlanID                     string
	LegRequestedQuantity       string
	LegFilledQuantity          string
	LegResidualQuantity        string
	IntentID                   string
	AttemptID                  string
	OrderID                    string
	PredecessorOrderID         string
	CarryBaseline              string
	RequestedCap               string
	DeltaQuantity              string
	CumulativeQuantity         string
	OrderRemainingQuantity     string
	OrderTerminal              bool
	OrderLineageAmbiguous      bool
	EffectiveStop              string
	StopSource                 string
	StopPolicy                 string
	StopObservedAt             string
	EntryBlocked               bool
	ProjectionDigest           string
}

type Snapshot struct {
	CampaignState      CampaignState
	Version            int64
	PositionGeneration int64
	ProspectiveToken   string
	EffectiveStop      string
	EntryBlocked       bool
	ProjectionDigest   string
}

type ReplayResult struct {
	Valid             bool
	Reason            ReplayReason
	LastValidSequence int64
	State             Snapshot
}

type replayOrder struct {
	leg         int64
	predecessor string
	// attempt 는 broker order **하나**의 mutation attempt 다. 이 결속은 order
	// 수준에서 불변이고, 스키마도 attempt_id 를 campaign_order_watermarks 에만
	// 둔다 — campaign_legs 에는 없다.
	attempt    string
	cap        string
	cumulative string
	terminal   bool
}

// replayLeg 에 AttemptID 가 없는 것은 의도다. design D5 에서 교체 주문은 **같은 leg
// 위에** 새 order/attempt 를 만든다. leg 수준 attempt 불변을 요구하면 정상적인 교체가
// LEG_IDENTITY_CHANGED 로 신고된다 — 그 불변은 스키마에도 설계에도 없었다.
type replayLeg struct {
	Sequence          int64
	PlanID            string
	IntentID          string
	State             LegState
	RequestedQuantity string
	FilledQuantity    string
	ResidualQuantity  string
}

// Replay is deliberately pure and offline: it reads an event slice, performs
// no repair, and has no path to a broker, journal writer or runtime toggle.
func Replay(events []Event, snapshot Snapshot) ReplayResult {
	state := Snapshot{}
	seen := make(map[string]struct{}, len(events))
	legs := make(map[int64]replayLeg)
	orders := make(map[string]replayOrder)
	lastLeg := int64(0)
	expectedGeneration := int64(0)
	expectedGenerationSet := false
	result := ReplayResult{Valid: true, Reason: ReplayOK}
	for i, event := range events {
		want := int64(i + 1)
		if event.Sequence != want {
			result.Valid = false
			result.Reason = ReplaySequenceGap
			result.State = state
			return result
		}
		if event.CampaignVersion != event.Sequence {
			return replayFailure(result, state, ReplayCampaignVersionMismatch)
		}
		if strings.TrimSpace(event.EventKind) == "" || strings.TrimSpace(event.RequestDigest) == "" {
			return replayFailure(result, state, ReplayInvalidEvidence)
		}
		if err := ValidateCommand(event.CommandKind, event.CommandKey); err != nil {
			return replayFailure(result, state, ReplayInvalidEvidence)
		}
		key, _ := TypedCommandIdentity(event.CommandKind, event.CommandKey)
		if _, exists := seen[key]; exists {
			return replayFailure(result, state, ReplayDuplicateCommandKey)
		}
		seen[key] = struct{}{}
		nextCampaign, transitionOK := replayCampaignTransition(state, event)
		if !transitionOK {
			if state.CampaignState == CampaignClosed && event.CampaignState != CampaignClosed {
				return replayFailure(result, state, ReplayClosedReopened)
			}
			return replayFailure(result, state, ReplayInvalidCampaignTransition)
		}
		if event.CommandKind == string(CommandCreate) || event.ExpectedPositionGeneration != 0 {
			if expectedGenerationSet && expectedGeneration != event.ExpectedPositionGeneration {
				return replayFailure(result, state, ReplayGenerationRebound)
			}
			expectedGeneration = event.ExpectedPositionGeneration
			expectedGenerationSet = true
		}
		if event.ProspectiveToken != "" {
			if state.ProspectiveToken != "" && state.ProspectiveToken != event.ProspectiveToken {
				return replayFailure(result, state, ReplayProspectiveRebound)
			}
			state.ProspectiveToken = event.ProspectiveToken
		}
		if event.PositionGeneration != 0 {
			if state.PositionGeneration != 0 && event.PositionGeneration != state.PositionGeneration {
				if event.PositionGeneration < state.PositionGeneration {
					return replayFailure(result, state, ReplayGenerationRetreat)
				}
				return replayFailure(result, state, ReplayGenerationRebound)
			}
			if state.PositionGeneration == 0 && expectedGenerationSet && event.PositionGeneration != expectedGeneration+1 {
				return replayFailure(result, state, ReplayGenerationRebound)
			}
		}
		if event.OrderID != "" && event.PredecessorOrderID != "" {
			predecessor, found := orders[event.PredecessorOrderID]
			if !found || predecessor.leg != event.LegSequence {
				return replayFailure(result, state, ReplayOrphanOrderLineage)
			}
		}
		if event.LegSequence > 0 {
			leg, exists := legs[event.LegSequence]
			if !exists {
				if event.EventKind != "LEG_PLANNED" || event.PlanID == "" || event.LegSequence != lastLeg+1 {
					return replayFailure(result, state, ReplayLegSequenceGap)
				}
				leg = replayLeg{Sequence: event.LegSequence, PlanID: event.PlanID,
					IntentID: event.IntentID, State: LegPlanned,
					RequestedQuantity: event.LegRequestedQuantity, FilledQuantity: "0", ResidualQuantity: event.LegRequestedQuantity}
				legs[event.LegSequence] = leg
				lastLeg = event.LegSequence
			} else if (event.PlanID != "" && leg.PlanID != event.PlanID) ||
				(event.IntentID != "" && leg.IntentID != "" && leg.IntentID != event.IntentID) {
				return replayFailure(result, state, ReplayLegIdentityChanged)
			}
			if event.IntentID != "" && leg.IntentID == "" {
				leg.IntentID = event.IntentID
			}
			if event.EventKind == "ORDER_LINKED" {
				legEvent := LegOrderLinked
				if event.PredecessorOrderID != "" {
					legEvent = LegReplacementLinked
				}
				next, err := TransitionLeg(leg.State, legEvent)
				if err != nil && leg.State == LegCancelled && legEvent == LegReplacementLinked {
					if cmp, compareErr := riskcalc.CompareDecimal(leg.FilledQuantity, "0"); compareErr == nil && cmp > 0 {
						next, err = LegPartial, nil
					} else {
						next, err = LegSubmitted, nil
					}
				}
				if err != nil || event.LegState != next {
					return replayFailure(result, state, ReplayInvalidLegTransition)
				}
				leg.State = next
			} else if event.EventKind == "LEG_PLANNED" && event.LegState != LegPlanned {
				return replayFailure(result, state, ReplayInvalidLegTransition)
			}
			legs[event.LegSequence] = leg
		}
		if event.OrderID != "" {
			order, exists := orders[event.OrderID]
			if !exists {
				if event.LegSequence <= 0 || event.RequestedCap == "" {
					return replayFailure(result, state, ReplayInvalidEvidence)
				}
				if event.PredecessorOrderID != "" {
					predecessor, found := orders[event.PredecessorOrderID]
					if !found || predecessor.leg != event.LegSequence {
						return replayFailure(result, state, ReplayOrphanOrderLineage)
					}
				}
				if cap, err := positiveDecimal(event.RequestedCap); err != nil {
					return replayFailure(result, state, ReplayInvalidEvidence)
				} else {
					order = replayOrder{leg: event.LegSequence, predecessor: event.PredecessorOrderID,
						attempt: event.AttemptID, cap: cap, cumulative: "0"}
					orders[event.OrderID] = order
				}
			} else if event.LegSequence != 0 && (order.leg != event.LegSequence ||
				(event.PredecessorOrderID != "" && order.predecessor != event.PredecessorOrderID) ||
				(event.AttemptID != "" && order.attempt != "" && order.attempt != event.AttemptID)) {
				return replayFailure(result, state, ReplayOrderIdentityChanged)
			}
			if event.AttemptID != "" && order.attempt == "" {
				order.attempt = event.AttemptID
				orders[event.OrderID] = order
			}
			if event.CumulativeQuantity != "" {
				cumulative, err := nonNegative(event.CumulativeQuantity)
				if err != nil {
					return replayFailure(result, state, ReplayInvalidEvidence)
				}
				cmp, err := riskcalc.CompareDecimal(cumulative, order.cumulative)
				if err != nil || cmp < 0 {
					return replayFailure(result, state, ReplayWatermarkRetreat)
				}
				if event.EventKind == "ORDER_WATERMARK_ADVANCED" {
					delta, deltaErr := nonNegative(event.DeltaQuantity)
					sum, sumErr := riskcalc.AddDecimal(order.cumulative, delta)
					if deltaErr != nil || sumErr != nil || sum != cumulative {
						return replayFailure(result, state, ReplayOrderDeltaMismatch)
					}
				}
				order.cumulative = cumulative
				order.terminal = event.OrderTerminal
				orders[event.OrderID] = order
			}
			remaining, err := StoredOrderRemaining(order.predecessor != "", order.terminal,
				order.cap, order.cumulative, event.LegResidualQuantity)
			if err != nil || remaining != event.OrderRemainingQuantity {
				return replayFailure(result, state, ReplayOrderRemainingMismatch)
			}
			if event.OrderLineageAmbiguous {
				return replayFailure(result, state, ReplayInvalidEvidence)
			}
		}
		if event.LegSequence > 0 {
			leg := legs[event.LegSequence]
			if err := validateReplayLegQuantities(event, leg, orders); err != nil {
				return replayFailure(result, state, ReplayLegQuantityMismatch)
			}
			if event.EventKind == "ORDER_WATERMARK_ADVANCED" {
				delta, deltaErr := nonNegative(event.DeltaQuantity)
				if deltaErr != nil {
					return replayFailure(result, state, ReplayInvalidEvidence)
				}
				// 후속이 있으면 잔량은 취소된 것이 아니라 교체 주문으로 넘어간 것이다.
				hasSuccessor := false
				for _, candidate := range orders {
					if candidate.predecessor == event.OrderID {
						hasSuccessor = true
						break
					}
				}
				next, err := LegStateAfterFill(LegFillFacts{
					LegState: leg.State, Delta: delta,
					LegFilled: event.LegFilledQuantity, LegRequested: event.LegRequestedQuantity,
					OrderTerminal: event.OrderTerminal, OrderHasSuccessor: hasSuccessor,
				})
				if err != nil || next != event.LegState {
					return replayFailure(result, state, ReplayInvalidLegTransition)
				}
				leg.State = next
			}
			leg.RequestedQuantity = event.LegRequestedQuantity
			leg.FilledQuantity = event.LegFilledQuantity
			leg.ResidualQuantity = event.LegResidualQuantity
			legs[event.LegSequence] = leg
		}
		if event.EffectiveStop != "" {
			stop, err := positiveDecimal(event.EffectiveStop)
			if err != nil || strings.TrimSpace(event.StopSource) == "" ||
				strings.TrimSpace(event.StopPolicy) == "" || strings.TrimSpace(event.StopObservedAt) == "" {
				return replayFailure(result, state, ReplayInvalidEvidence)
			}
			if state.EffectiveStop != "" {
				cmp, err := riskcalc.CompareDecimal(stop, state.EffectiveStop)
				if err != nil || cmp < 0 {
					return replayFailure(result, state, ReplayStopRetreat)
				}
			}
			state.EffectiveStop = stop
		}
		state.CampaignState = nextCampaign.State
		state.EntryBlocked = nextCampaign.EntryBlocked
		state.ProjectionDigest = event.ProjectionDigest
		if event.PositionGeneration != 0 {
			state.PositionGeneration = event.PositionGeneration
		}
		state.Version = event.Sequence
		result.LastValidSequence = event.Sequence
	}
	result.State = state
	if snapshot.Version != 0 && (snapshot != state) {
		result.Valid = false
		result.Reason = ReplaySnapshotDrift
	}
	return result
}

func replayCampaignTransition(state Snapshot, event Event) (CampaignTransition, bool) {
	if event.Sequence == 1 {
		if event.EventKind != "CREATED" || event.CampaignState != CampaignPlanned || event.EntryBlocked {
			return CampaignTransition{}, false
		}
		return CampaignTransition{State: CampaignPlanned}, true
	}
	var transition CampaignTransition
	var err error
	switch event.EventKind {
	case "LEG_PLANNED":
		transition, err = TransitionCampaign(state.CampaignState, CampaignLegPlanned)
	case "ORDER_LINKED":
		transition, err = TransitionCampaign(state.CampaignState, CampaignOrderLinked)
	case "ORDER_LINK_REFUSED":
		transition, err = TransitionCampaign(state.CampaignState, CampaignLineageMismatch)
	case "PROSPECTIVE_CANCELLED":
		transition, err = TransitionCampaign(state.CampaignState, CampaignCancelledBeforeFill)
	case "AMBIGUOUS_ORDER_FILL":
		if state.CampaignState == CampaignClosed {
			transition, err = TransitionCampaign(state.CampaignState, CampaignLateFact)
		} else if state.CampaignState == CampaignReconcile {
			transition, err = TransitionCampaign(state.CampaignState, CampaignEvidenceIncomplete)
		} else {
			transition, err = TransitionCampaign(state.CampaignState, CampaignLineageMismatch)
		}
	case "ORDER_WATERMARK_ADVANCED":
		switch {
		case state.CampaignState == CampaignClosed:
			transition, err = TransitionCampaign(state.CampaignState, CampaignLateFact)
		case event.CampaignState == CampaignClosed:
			transition, err = TransitionCampaign(state.CampaignState, CampaignPositionClosed)
		case event.CampaignState == CampaignReconcile:
			if state.CampaignState == CampaignReconcile {
				transition, err = TransitionCampaign(state.CampaignState, CampaignEvidenceIncomplete)
			} else {
				transition, err = TransitionCampaign(state.CampaignState, CampaignWatermarkMismatch)
			}
		case state.CampaignState == CampaignExiting:
			transition, err = TransitionCampaign(state.CampaignState, CampaignRiskReducingProgress)
		default:
			transition, err = TransitionCampaign(state.CampaignState, CampaignLegProgress)
		}
	case "STOP_COMPOSED":
		// 이 갈래만 event 의 값을 입력으로 받는다. stop 후보(가격·유효성)는 event 에
		// 실리지 않으므로 replay 가 ComposeLongStop 의 block 판정을 되계산할 방법이
		// 없다. 되계산할 수 없는 것을 지어내는 대신 event 를 믿고, **단조성**만
		// 아래 LatchEntryBlocked 로 강제한다.
		transition = CampaignTransition{State: state.CampaignState, EntryBlocked: event.EntryBlocked}
	default:
		return CampaignTransition{}, false
	}
	// 이미 걸린 진입 차단은 어떤 전이도 풀지 않는다. 원장에 쓰는 쪽과 같은 함수다.
	transition = LatchEntryBlocked(state.EntryBlocked, transition)
	return transition, err == nil && transition.State == event.CampaignState && transition.EntryBlocked == event.EntryBlocked
}

func validateReplayLegQuantities(event Event, leg replayLeg, orders map[string]replayOrder) error {
	requested, err := positiveDecimal(event.LegRequestedQuantity)
	if err != nil {
		return err
	}
	if leg.RequestedQuantity != "" && leg.RequestedQuantity != requested {
		return ErrInvalidIdentity
	}
	filled, err := nonNegative(event.LegFilledQuantity)
	if err != nil {
		return err
	}
	residual, err := nonNegative(event.LegResidualQuantity)
	if err != nil {
		return err
	}
	wantResidual, err := remainingFromCap(requested, filled)
	if err != nil || residual != wantResidual {
		return ErrInvalidIdentity
	}
	total := "0"
	for _, order := range orders {
		if order.leg == event.LegSequence {
			total, err = riskcalc.AddDecimal(total, order.cumulative)
			if err != nil {
				return err
			}
		}
	}
	if total != filled {
		return ErrInvalidIdentity
	}
	return nil
}

// RemainingFromCap 은 주문 하나가 자기 상한 대비 아직 살 수 있는 수량이다.
// 원장 쪽에도 같은 산술이 따로 있었다 — 이제 그쪽이 이 함수를 부른다.
func RemainingFromCap(cap, cumulative string) (string, error) {
	cap, err := nonNegative(cap)
	if err != nil {
		return "", err
	}
	cumulative, err = nonNegative(cumulative)
	if err != nil {
		return "", err
	}
	cmp, err := riskcalc.CompareDecimal(cumulative, cap)
	if err != nil {
		return "", err
	}
	if cmp >= 0 {
		return "0", nil
	}
	return riskcalc.SubDecimal(cap, cumulative)
}

func remainingFromCap(cap, cumulative string) (string, error) {
	return RemainingFromCap(cap, cumulative)
}

func replayFailure(result ReplayResult, state Snapshot, reason ReplayReason) ReplayResult {
	result.Valid = false
	result.Reason = reason
	result.State = state
	return result
}

func positiveDecimal(value string) (string, error) {
	canonical, err := nonNegative(value)
	if err != nil {
		return "", err
	}
	cmp, err := riskcalc.CompareDecimal(canonical, "0")
	if err != nil || cmp <= 0 {
		return "", ErrInvalidIdentity
	}
	return canonical, nil
}
