package candidate

// scan.go is one pass over the sources.
//
// # What a scan is responsible for not doing
//
// The interesting work here is negative. A scan must not:
//
//	stop because a source failed     WTS sessions expire on ordinary days, so a
//	                                 panel that halts with them halts routinely.
//	hide that a source failed        a candidate produced from two of five sources
//	                                 is not wrong, but a reader who does not know
//	                                 it was two of five is judging on less than
//	                                 they think.
//	report a failure as an empty     "we looked and the market is quiet" and "we
//	market                           could not look" are different facts.
//	cool a symbol it never looked    the one that would destroy the record — see
//	for                              coolAbsent below.
//
// # The absent symbol, and why a 429 must not cool anything
//
// A scan cools the candidates its sources no longer raise. That is how a symbol
// leaving the ranking list starts its cooling clock promptly, instead of waiting
// for the staleness fallback, which exists for the scan that died rather than as
// the normal path.
//
// But "no longer raised" is a claim about evidence, and the evidence is missing
// exactly when a source fails. The rule is therefore narrower than "not seen this
// time":
//
//	A candidate may be cooled only when EVERY source that previously supported
//	it answered in this scan and none of them raised it.
//
// Anything looser destroys the record. If the ranking source 429s and the scan
// cools everything it did not see, a rate limit cools the whole watchlist at
// once — and the cooling clock then expires it, taking every first_seen_at with
// it. A transient limit would erase the evidence of everything we found early,
// and nothing would fail while it happened. Candidate.Sources is what makes the
// narrow rule checkable: it already records who raised this candidate.

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// SourceFailure is one source that did not answer, and why.
//
// The reason travels with the id because "degraded: true" on its own is a flag an
// operator cannot act on — an expired WTS session and a rate-limited official
// endpoint call for opposite responses.
type SourceFailure struct {
	Source SourceID
	// Reason is the error, as text, for a human.
	Reason string
	// RateLimited separates a 429 from every other failure. For the official
	// ranking — which has no WTS fallback — the rate at which this happens is the
	// only measurement available of an undocumented limit.
	RateLimited bool
}

// SymbolFailure is one symbol the store refused to record, and why.
//
// Separate from SourceFailure because the remedies are unrelated: a source
// failure is a network or session problem, a symbol failure is the lifecycle
// declining a write — most often an instant that runs backwards.
type SymbolFailure struct {
	Symbol string
	Reason string
}

// ScanResult is what one pass did and what it could not see.
type ScanResult struct {
	Market string
	At     time.Time

	// Attempted and Responded are the panel's completeness.
	Attempted, Responded int
	// Missing names every source that did not answer.
	Missing []SourceFailure
	// Degraded is true when any source is missing. It is derived rather than set,
	// so it cannot disagree with Missing.
	Degraded bool

	// Observations and Candidates count what was written. Observations is per
	// (source, symbol): two sources raising one symbol is two observations and
	// one candidate.
	Observations, Candidates int
	// Cooled counts the candidates this scan stopped seeing.
	Cooled int
	// FirstPrices and FirstRanks count the candidates whose baseline and first
	// sighting position this scan recorded for the first time (D17 and its other
	// half). They are the durable half of two vetoes, and a scan that writes
	// neither leaves both permanently unmeasured while nothing fails — so the
	// count is on the result rather than inferred from the store.
	FirstPrices, FirstRanks int
	// Rejected names the symbols the store declined to record. A scan with
	// rejections is not a failed scan — the rest of the market still went
	// through — but it is not a clean one either, and the count is what makes a
	// clock that runs backwards visible before it expires the watchlist.
	//
	// It also carries the two stored firsts' write failures. Those happen after a
	// successful promotion, so the candidate is counted and the failure is named:
	// a promoted candidate with no baseline is a live record with a veto that can
	// never be measured, which is worth seeing.
	Rejected []SymbolFailure

	// Budgets is the rate allowance each source reported, by source. A source
	// that reported nothing has a Budget with Reported false rather than an
	// absent entry, so a caller reading the map cannot mistake silence for zero.
	Budgets map[SourceID]Budget
}

// RateLimitedSources returns the sources this scan lost to a 429.
//
// The count over time is the measurement D13 decision 2 wants: an undocumented
// limit inferred from how often we hit it, rather than guessed.
func (r ScanResult) RateLimitedSources() []SourceID {
	var out []SourceID
	for _, m := range r.Missing {
		if m.RateLimited {
			out = append(out, m.Source)
		}
	}
	return out
}

// CollectOptions configures one scan.
type CollectOptions struct {
	// Market is the market being scanned.
	Market string
	// At is the instant to record. Supplied by the caller — the scan does not
	// call the clock itself, so a test can drive a whole session through it.
	At time.Time
	// Sources is the panel. Order is preserved in the result.
	Sources []Source
}

// Collect runs one scan: read every source, record what they said, and move the
// candidate lifecycle to match.
//
// It returns ErrNoSourceAnswered when the whole panel failed. Everything short of
// that is a degraded success, recorded as such.
//
// One case returns a populated result *and* an error: a source claiming a WTS
// fallback it cannot have (ErrFallbackNotPossible). That is a wiring defect
// rather than a market condition, so the scan completes — refusing to promote or
// cool would expire the whole watchlist to report a bug — and the error is the
// alarm. Callers should log it and keep the result.
func Collect(ctx context.Context, store *Store, opts CollectOptions) (ScanResult, error) {
	market := strings.ToUpper(strings.TrimSpace(opts.Market))
	if market == "" {
		return ScanResult{}, fmt.Errorf("candidate: a scan needs a market")
	}
	if opts.At.IsZero() {
		return ScanResult{}, fmt.Errorf("candidate: a scan of %s has no instant", market)
	}
	at := opts.At.UTC()

	// A panel with two sources under one id is refused before anything is read.
	//
	// `responded` and `Budgets` are keyed by source id, and the cooling rule is
	// "every source that raised this candidate answered". A collision makes a
	// source that answered vouch for one that did not, which is how a rate limit
	// on one ranking cools candidates only the other one ever saw. It is cheap to
	// check and impossible to notice once it is running.
	seen := make(map[SourceID]bool, len(opts.Sources))
	for _, src := range opts.Sources {
		id := src.ID()
		if seen[id] {
			return ScanResult{}, fmt.Errorf(
				"candidate: two sources in the %s panel share the id %q; "+
					"a source that answers would vouch for one that did not", market, id)
		}
		seen[id] = true
	}

	result := ScanResult{
		Market:    market,
		At:        at,
		Attempted: len(opts.Sources),
		Budgets:   make(map[SourceID]Budget, len(opts.Sources)),
	}

	var (
		observations []Observation
		// raisedBy is the candidate identity to the sources that raised it. It is
		// what makes two sources on one symbol one candidate with two supporters
		// rather than two candidates with half the evidence each.
		raisedBy = map[string][]SourceID{}
		// firstReading is, per symbol, the reading this scan would record as the
		// candidate's first price and first sighting position if it turns out not
		// to have one yet (D17 and its other half).
		//
		// Panel order decides, and it decides once: the first source in the panel
		// that carried a price wins the baseline, the first that carried a
		// position wins the rank. They are tracked separately because a source can
		// carry one without the other — the prices endpoint has no rank and a WTS
		// popularity row has no price — and folding them together would leave a
		// candidate raised by both with the field only one of them reported.
		firstPriced = map[string]Observation{}
		firstRanked = map[string]Observation{}
		// covered is every symbol a responding source reported.
		covered = map[string]bool{}
		// responded is every source that answered. Together with covered it is
		// what makes cooling a claim about evidence rather than about silence —
		// see the file header.
		responded = map[SourceID]bool{}
		// misWired are sources that claimed a fallback they cannot have. Kept
		// aside so the scan can finish and still fail loudly.
		misWired []SourceID
		anyRead  bool
	)

	for _, src := range opts.Sources {
		id := src.ID()
		reading, err := src.Read(ctx, market)
		if err != nil {
			result.Missing = append(result.Missing, SourceFailure{
				Source:      id,
				Reason:      err.Error(),
				RateLimited: isRateLimited(err),
			})
			continue
		}
		if reading.ViaFallback && !id.hasFallback() {
			// Loud, but not fatal to the whole pass. This is a wiring defect
			// rather than a market event, and a deterministic wiring defect that
			// aborted every scan would leave the watchlist un-promoted and
			// un-cooled until the staleness clock expired all of it — destroying
			// the first_seen_at record in order to report a bug.
			//
			// The source is dropped, the reason goes into Missing, and the error
			// is returned at the end alongside a fully populated result.
			result.Missing = append(result.Missing, SourceFailure{
				Source: id,
				Reason: fmt.Sprintf("%v (the reading is mis-wired)", ErrFallbackNotPossible),
			})
			misWired = append(misWired, id)
			continue
		}

		result.Responded++
		anyRead = true
		result.Budgets[id] = reading.Budget

		// An empty reading is NOT evidence that anything left the list.
		//
		// Cooling asks "did a source that covers this symbol answer and not list
		// it". A ranking that returns zero rows — out of hours, a server blip, a
		// body whose symbols were all blank — answered about nothing, and treating
		// it as full coverage cools the entire watchlist on a single 200. That
		// reaches the same destruction as the 429 case with nothing having failed
		// at all, so an empty reading responds without vouching.
		if len(reading.Rows) == 0 {
			result.Missing = append(result.Missing, SourceFailure{
				Source: id,
				Reason: "responded with no rows; not counted as coverage",
			})
			continue
		}
		responded[id] = true

		for _, r := range reading.Rows {
			symbol := strings.TrimSpace(r.Symbol)
			if symbol == "" {
				continue
			}
			o := Observation{
				Market: market, Symbol: symbol, Source: id, ObservedAt: at,
				Reported: Reported{
					Rank: r.Rank, RankTotal: r.RankTotal, NewlyListed: r.NewlyListed,
					TradingValue: r.TradingValue, TradingVolume: r.TradingVolume,
					Price: r.Price, DayHigh: r.DayHigh, DayLow: r.DayLow,
				},
				ViaFallback: reading.ViaFallback,
			}
			observations = append(observations, o)
			covered[symbol] = true
			raisedBy[symbol] = append(raisedBy[symbol], id)
			if _, taken := firstPriced[symbol]; !taken && strings.TrimSpace(r.Price) != "" {
				firstPriced[symbol] = o
			}
			if _, taken := firstRanked[symbol]; !taken && r.Rank > 0 && r.RankTotal > 0 {
				firstRanked[symbol] = o
			}
		}
	}

	result.Degraded = len(result.Missing) > 0
	if !anyRead {
		// The mis-wire wins when both apply. "No source answered" is true but it
		// is the symptom; a source claiming a fallback it cannot have is the
		// cause, and it is the one an operator can act on.
		if len(misWired) > 0 {
			return result, fmt.Errorf("%w: %v", ErrFallbackNotPossible, misWired)
		}
		return result, fmt.Errorf("%w: %d source(s) attempted for %s",
			ErrNoSourceAnswered, result.Attempted, market)
	}

	if err := store.RecordObservations(ctx, observations); err != nil {
		return result, err
	}
	result.Observations = len(observations)

	symbols := make([]string, 0, len(raisedBy))
	for symbol := range raisedBy {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	promoted := make([]string, 0, len(symbols))

	// One symbol's refusal is not the market's.
	//
	// Promote rejects an instant older than the candidate's last reading, which a
	// clock step backwards produces. Aborting the loop there would leave the rest
	// of the symbols un-promoted and skip cooling entirely — and because the loop
	// is sorted, the same symbol would block every subsequent scan for as long as
	// the clock was behind, expiring the whole watchlist while observations kept
	// accumulating. So failures are collected and the pass finishes.
	for _, symbol := range symbols {
		if _, err := store.Promote(ctx, market, symbol, at); err != nil {
			result.Rejected = append(result.Rejected, SymbolFailure{Symbol: symbol, Reason: err.Error()})
			continue
		}
		err := store.NoteSources(ctx, market, symbol,
			raisedBy[symbol], result.Attempted, result.Responded, result.Degraded)
		if err != nil {
			result.Rejected = append(result.Rejected, SymbolFailure{Symbol: symbol, Reason: err.Error()})
			continue
		}
		result.Candidates++
		promoted = append(promoted, symbol)
	}

	if err := recordFirsts(ctx, store, market, at, promoted,
		firstPriced, firstRanked, &result); err != nil {
		return result, err
	}

	cooled, err := coolAbsent(ctx, store, market, covered, responded, seen, at)
	result.Cooled = cooled
	if err != nil {
		return result, err
	}
	if len(misWired) > 0 {
		// The result is complete and usable; the error is the alarm. Both, because
		// silently continuing would let a mis-wire run for weeks feeding the
		// fallback-rate measurement a fact that cannot be true.
		return result, fmt.Errorf("%w: %v", ErrFallbackNotPossible, misWired)
	}
	return result, nil
}

// recordFirsts writes the two stored firsts for the candidates that still lack
// them (D17, and D17's other half).
//
// # Why the missing set is read once instead of asked per symbol
//
// Both NoteFirstPrice and NoteFirstRank are idempotent, so the naive wiring is to
// call them on every reading and let the store decide. That is correct and it is
// two IMMEDIATE transactions per symbol per tick — the write lock taken a few
// hundred times a scan to discover that there was nothing to write — against the
// table D16 names as the one that can fill the ledger's filesystem. issues.md 4
// left this decision to section 5 for that reason.
//
// So the summaries are read once, and a symbol is written to only when its column
// is still NULL. The store's own guard stays in place underneath: it is what makes
// two writers racing on one candidate keep the earlier baseline, and this pass is
// an optimisation on top of it rather than a replacement for it.
//
// # The read has to happen after the promotions
//
// Promote clears both columns when an expired candidate crosses again (D1), so a
// missing set gathered before the loop would say "already recorded" for a life that
// had just been reset — and the new life would run with the dead one's baseline
// still in place if the write were skipped. Reading after is what makes the new
// life's first reading the new life's first price.
//
// A write that fails is named in Rejected rather than aborting the pass. The
// candidate is already promoted at that point; losing the rest of the market
// because one baseline could not be stored is the shape scan.go refuses everywhere
// else.
func recordFirsts(ctx context.Context, store *Store, market string, at time.Time,
	promoted []string, priced, ranked map[string]Observation, result *ScanResult) error {

	if len(promoted) == 0 {
		return nil
	}
	summaries, err := store.Summaries(ctx, at)
	if err != nil {
		return err
	}
	needPrice := map[string]bool{}
	needRank := map[string]bool{}
	for _, s := range summaries {
		if s.Market != market {
			continue
		}
		needPrice[s.Symbol] = !s.Baseline.Recorded()
		needRank[s.Symbol] = !s.FirstRank.Recorded()
	}

	for _, symbol := range promoted {
		if o, ok := priced[symbol]; ok && needPrice[symbol] {
			_, err := store.NoteFirstPrice(ctx, market, symbol,
				o.Reported.Price, o.ObservedAt, o.Source)
			if err != nil {
				result.Rejected = append(result.Rejected, SymbolFailure{
					Symbol: symbol,
					Reason: fmt.Sprintf("the first price could not be recorded: %v", err),
				})
			} else {
				result.FirstPrices++
			}
		}
		o, ok := ranked[symbol]
		if !ok || !needRank[symbol] {
			continue
		}
		stored, err := store.NoteFirstRank(ctx, market, symbol,
			o.Reported.Rank, o.Reported.RankTotal, o.ObservedAt, o.Source)
		switch {
		case err != nil:
			result.Rejected = append(result.Rejected, SymbolFailure{
				Symbol: symbol,
				Reason: fmt.Sprintf("the first sighting position could not be recorded: %v", err),
			})
		case stored.Recorded():
			result.FirstRanks++
		}
		// A reading outside the identity window around first_seen_at is neither
		// stored nor an error — it is the ordinary state of a candidate raised by a
		// source that carries no position and only later seen on a ranking list.
		// Having no first rank is the honest answer there, and it makes seen_late
		// unmeasured rather than wrong.
	}
	return nil
}

// coolAbsent cools the active candidates whose supporters were all present in
// this scan, all answered, and none of them raised the candidate.
//
// Three guards, each closing a different way of getting this wrong.
//
//	covered    excludes symbols that were raised this scan.
//	responded  excludes candidates whose supporters did not all answer. Absent
//	           evidence is not evidence of absence, and treating it as such turns
//	           one 429 into a mass expiry of first_seen_at values.
//	panel      excludes supporters that are no longer in the panel at all.
//
// The third is subtler than it looks. Candidate.Sources accumulates over a
// candidate's life, so a source that raised it once and has since been removed
// from the configuration — an operator dropping WTS, which the design treats as
// routine — would never appear in `responded` again. Without this guard those
// candidates become permanently un-coolable and can only leave through the
// staleness fallback, which store.go reserves for the scan that died. A supporter
// that no longer exists is not missing evidence; it is a source that is gone.
//
// A candidate whose supporters are *all* outside the panel is still left alone:
// nothing in this scan was in a position to see it.
func coolAbsent(ctx context.Context, store *Store, market string,
	covered map[string]bool, responded, panel map[SourceID]bool, at time.Time) (int, error) {

	existing, err := store.Candidates(ctx, at)
	if err != nil {
		return 0, err
	}
	cooled := 0
	for _, c := range existing {
		if c.Market != market || c.State != StateActive || covered[c.Symbol] {
			continue
		}
		if !coverageAnswered(c.Sources, responded, panel) {
			continue
		}
		if err := store.Cool(ctx, market, c.Symbol, at); err != nil {
			return cooled, err
		}
		cooled++
	}
	return cooled, nil
}

// coverageAnswered reports that this scan was in a position to observe the
// candidate's absence: at least one of its supporters is still in the panel, and
// every supporter that is in the panel answered.
func coverageAnswered(sources []SourceID, responded, panel map[SourceID]bool) bool {
	present := 0
	for _, id := range sources {
		if !panel[id] {
			continue
		}
		present++
		if !responded[id] {
			return false
		}
	}
	return present > 0
}

// isRateLimited reports a 429.
//
// The sentinel is the contract: an adapter that wants its failure counted as a
// rate limit wraps ErrRateLimited, which candidatesrc does.
//
// The textual fallback is anchored, not a bare "429" search. An unanchored one
// matched any error whose text happened to contain those three digits — a Korean
// stock code, a trace id, a request id echoed inside an API error body, all of
// which reach err.Error() because classifyStatus embeds the whole response. Every
// false positive corrupts the one measurement available of the undocumented
// RANKING limit and triggers a retreat nobody asked for, so the loose match costs
// more than the cases it catches.
func isRateLimited(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	text := strings.ToLower(err.Error())
	for _, anchored := range []string{
		"status 429", "http 429", "429 too many requests", "rate limit", "too many requests",
	} {
		if strings.Contains(text, anchored) {
			return true
		}
	}
	return false
}
