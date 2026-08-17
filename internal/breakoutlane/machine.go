package breakoutlane

import "strings"

// Evaluate derives every transition from a sealed complete snapshot. It has no
// raw event/proposal entry point and returns values whose lineage is sealed.
func Evaluate(snapshot EvidenceSnapshot, prior *Decision) Decision {
	input := snapshot.value
	if !validSnapshot(snapshot) {
		return refused(snapshot, RefusalEvidenceInvalid)
	}
	setup := setupID(input)
	if prior != nil {
		if !validDecision(*prior) {
			return refused(snapshot, RefusalLineageSealMismatch)
		}
		if prior.setupID == setup {
			if prior.snapshotDigest == snapshot.digest {
				return *prior
			}
			if !validCorrectionLineage(snapshot, *prior) {
				return refused(snapshot, RefusalEvidenceInvalid)
			}
			if prior.phase == phaseProposed || terminalPhase(prior.phase) {
				preserved := *prior
				preserved.lineage = append([]barLineage(nil), prior.lineage...)
				preserved.diagnostic = DiagnosticCorrectionAfterProposal
				preserved.seal = decisionSeal(preserved)
				return preserved
			}
		}
	}
	return evaluateFresh(snapshot, setup)
}

func evaluateFresh(snapshot EvidenceSnapshot, setup string) Decision {
	v := snapshot.value
	bars := v.Bars
	resistance, low := bars[0].value.HighMinor, bars[0].value.LowMinor
	for _, bar := range bars[:v1OpeningRangeBars] {
		if bar.value.HighMinor > resistance {
			resistance = bar.value.HighMinor
		}
		if bar.value.LowMinor < low {
			low = bar.value.LowMinor
		}
	}
	p := Provenance{ResistanceMinor: resistance, RangeLowMinor: low, Transitions: []string{string(phaseDiscovered), string(phaseRangeLocked)}}
	state := phaseRangeLocked
	breakout := -1
	firstTouch := false
	for i, bar := range bars[v1OpeningRangeBars:] {
		b := bar.value
		if b.valueHighAbove(resistance) && !BreakoutCloseQualifies(b.CloseMinor, resistance, v.ATRMinor, v.Config) {
			firstTouch = true
		}
		if BreakoutCloseQualifies(b.CloseMinor, resistance, v.ATRMinor, v.Config) && b.RVOLPPM >= v.Config.value.RVOLMinPPM && b.UpperWickRangePPM <= v.Config.value.UpperWickRangeMaxPPM {
			breakout = i + v1OpeningRangeBars
			p.RVOLAdmission = true
			p.RVOLAt1200000 = b.RVOLPPM >= 1_200_000
			p.RVOLAt2000000 = b.RVOLPPM >= 2_000_000
			p.RVOLAt2500000 = b.RVOLPPM >= 2_500_000
			state = phaseBreakoutClosed
			p.Transitions = append(p.Transitions, string(state), string(phaseRetestWait))
			state = phaseRetestWait
			break
		}
	}
	if breakout < 0 {
		d := newDecision(setup, snapshot, state, RefusalNone, p)
		if firstTouch {
			d.refusal = RefusalFirstTouch
			d.seal = decisionSeal(d)
		}
		return d
	}
	retest := false
	timeout := v.Config.value.TimeoutKR
	if v.Market == MarketUS {
		timeout = v.Config.value.TimeoutUS
	}
	for i := breakout + 1; i < len(bars); i++ {
		b := bars[i].value
		since := uint64(i - breakout)
		if b.CloseMinor < low {
			return newDecision(setup, snapshot, phaseInvalidated, RefusalNone, appendTransition(p, string(phaseInvalidated)))
		}
		if since > timeout {
			return newDecision(setup, snapshot, phaseTimedOut, RefusalNone, appendTransition(p, string(phaseTimedOut)))
		}
		if retest && b.CloseMinor >= resistance {
			p.Transitions = append(p.Transitions, string(phaseReclaimed), string(phaseArmed))
			state = phaseArmed
			break
		}
		if retest && b.CloseMinor < resistance && b.VolumeExpanded {
			return newDecision(setup, snapshot, phaseInvalidated, RefusalNone, appendTransition(p, string(phaseInvalidated)))
		}
		if since >= timeout {
			return newDecision(setup, snapshot, phaseTimedOut, RefusalNone, appendTransition(p, string(phaseTimedOut)))
		}
		if !retest && RetestQualifies(b.CloseMinor, resistance, v.ATRMinor, v.Config) {
			retest = true
		}
	}
	if state != phaseArmed {
		return newDecision(setup, snapshot, state, RefusalNone, p)
	}
	if refusal := validateQuote(v.Quote, v.EvaluatedAtMS, v.Sizing.ProposedEntryMinor, v.Config); refusal != RefusalNone {
		return newDecision(setup, snapshot, phaseArmed, refusal, p)
	}
	result := size(v.Sizing, v.Quote, v.FX, v.EvaluatedAtMS)
	if result.Refusal != RefusalNone {
		return newDecision(setup, snapshot, phaseArmed, result.Refusal, p)
	}
	p.Transitions = append(p.Transitions, string(phaseProposed))
	d := newDecision(setup, snapshot, phaseProposed, RefusalNone, p)
	d.candidate = result.CandidateQuantity
	d.final = result.FinalQuantity
	d.proposalID = hashFields("proposal.v1", setup, snapshot.digest, v.Config.Digest())
	d.seal = decisionSeal(d)
	return d
}
func (b ClosedBarInput) valueHighAbove(resistance uint64) bool { return b.HighMinor > resistance }
func appendTransition(p Provenance, state string) Provenance {
	p.Transitions = append(p.Transitions, state)
	return p
}
func terminalPhase(phase phase) bool {
	return phase == phaseInvalidated || phase == phaseTimedOut || phase == phaseConsumed
}
func refused(s EvidenceSnapshot, r RefusalCode) Decision {
	return newDecision("", s, phaseDiscovered, r, Provenance{})
}
func newDecision(setup string, s EvidenceSnapshot, phase phase, r RefusalCode, p Provenance) Decision {
	d := Decision{setupID: setup, snapshotDigest: s.digest, configDigest: s.value.Config.Digest(), phase: phase, refusal: r, provenance: p, lineage: lineageFrom(s)}
	d.seal = decisionSeal(d)
	return d
}
func validStructuralEvidenceInput(v EvidenceInput) bool {
	if (v.Market != MarketKR && v.Market != MarketUS) || !canonical(v.Symbol) || !canonical(v.SessionID) || !canonical(v.CalendarVersion) || !canonical(v.LaneID) || v.LaneVersion != LaneVersionV1 || !v.Config.Valid() || len(v.Bars) < v1OpeningRangeBars || len(v.Bars) > 512 || v.ATRMinor == 0 {
		return false
	}
	if v.Market == MarketKR && v.LaneID != KRLaneID || v.Market == MarketUS && v.LaneID != USLaneID {
		return false
	}
	seen := map[string]bool{}
	var prev uint64
	for n, bar := range v.Bars {
		b := bar.value
		if _, err := NewClosedBar(b); err != nil || b.SessionID != v.SessionID || seen[b.ID] || n > 0 && b.Sequence != prev+1 {
			return false
		}
		seen[b.ID] = true
		prev = b.Sequence
	}
	return quoteValid(v.Quote) && fxValid(v.FX)
}
func validSnapshot(s EvidenceSnapshot) bool {
	return s.digest != "" && s.digest == snapshotDigest(s.value) && validStructuralEvidenceInput(s.value)
}
func lineageFrom(s EvidenceSnapshot) []barLineage {
	out := make([]barLineage, len(s.value.Bars))
	for i, b := range s.value.Bars {
		out[i] = barLineage{sequence: b.value.Sequence, revision: b.value.Revision, id: b.value.ID, sessionID: b.value.SessionID, contentDigest: closedBarContentDigest(b.value)}
	}
	return out
}
func validCorrectionLineage(s EvidenceSnapshot, prior Decision) bool {
	now := lineageFrom(s)
	if len(now) < len(prior.lineage) {
		return false
	}
	changed := len(now) > len(prior.lineage)
	for i, old := range prior.lineage {
		next := now[i]
		if old.sequence != next.sequence || old.id != next.id || old.sessionID != next.sessionID || next.revision < old.revision || next.revision == old.revision && next.contentDigest != old.contentDigest {
			return false
		}
		if next.revision > old.revision {
			changed = true
		}
	}
	return changed
}
func validDecision(d Decision) bool {
	return d.seal != "" && len(d.lineage) <= 512 && d.seal == decisionSeal(d)
}
func decisionSeal(d Decision) string {
	parts := []string{"decision.v1", d.setupID, d.snapshotDigest, d.proposalID, d.configDigest, string(d.phase), string(d.refusal), string(d.diagnostic), u(d.candidate), u(d.final), strings.Join(d.provenance.Transitions, "\x00")}
	for _, b := range d.lineage {
		parts = append(parts, u(b.sequence), u(b.revision), b.id, b.sessionID, b.contentDigest)
	}
	return hashFields(parts...)
}

func closedBarContentDigest(b ClosedBarInput) string {
	return hashFields("tossos.breakout.closed-bar.v1", u(b.Sequence), u(b.Revision), u(b.IntervalMS), b.ID, b.SessionID, u(b.HighMinor), u(b.LowMinor), u(b.CloseMinor), u(b.RVOLPPM), u(b.UpperWickRangePPM), boolString(b.RegularSession), boolString(b.Closed), boolString(b.VolumeExpanded))
}
func setupID(v EvidenceInput) string {
	return "sha256:" + hashNUL("tossos.breakout.setup.v1", string(v.Market), v.Symbol, v.SessionID, v.CalendarVersion, v.Bars[0].value.ID, v.Bars[v1OpeningRangeBars-1].value.ID, v.LaneID, v.LaneVersion, v.Config.Digest())
}
