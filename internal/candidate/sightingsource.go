package candidate

// sightingsource.go answers a question the veto tally cannot: which source's
// readings are the ones seen_late keeps refusing.
//
// # The question it exists for
//
// Nobody has measured whether the official rankings endpoint returns the hundred
// rows it is asked for. If it does not, the distribution the follow-up change would
// choose a seen_late threshold from does not exist — while the screen shows a large
// unmeasured count and no way to tell which of the four ranking sources produced it.
//
// # Which refusal a short endpoint actually shows (corrected 2026-07-28)
//
// This said READING_TRUNCATED, which is what a short endpoint would produce on its
// own. It is not what this build produces, because two other rules reach the reading
// first. A source that comes back short never replaces its own memory, so its
// new-entrant answer stays unknown for as long as the degradation lasts; and a
// position no reading in the tick could qualify is not stored at all, it is counted
// in ScanResult.FirstRanksHeld. So nothing is ever written for those candidates, and
// what the census fills with is NO_FIRST_RANK — turning into NO_FIRST_SIGHTING once
// the candidate is older than the identity window, since by then no reading left in
// the assessment window is near enough to first_seen_at to be the one.
//
// READING_TRUNCATED is what appears when the *other* half is intact: a source that
// answers about its previous reading and still arrives short, which is the shape a
// partial degradation takes rather than a permanent one.
//
// All four of those reasons — plus NEW_ENTRANT_UNKNOWN and REQUEST_UNRECORDED — are
// statements about a reading rather than about a symbol, and none of them says whose
// reading it was. That is what this file adds.
//
// A count per code, which is what VetoTally gives, cannot separate "the official
// trading-value list comes back short" from "the WTS popularity list does". Those
// are two different things to go and fix, and one of them is the reason the whole
// measurement would be empty.
//
// # Why it is a reducer here rather than a rendering there
//
// Both surfaces need it — `tossctl candidate scan` and /signals — and this
// repository has already paid once for the arrangement where each screen assembles
// its own: a fourth shadow band appeared on one and not the other, and a band
// missing from a page reads as a band nobody crossed rather than as one nobody
// counted. So the judgement is here and only the wording is theirs.
//
// It reads the Sighting rather than the veto, because the Sighting is what carries
// the source. AssessSeenLate's VetoState knows the reason and not who produced the
// reading.

import "sort"

// SourceSightings is one ranking source's first-sighting record across a market's
// live candidates.
//
// The denominator is candidates whose stored first sighting came from this source,
// not readings and not symbols — a candidate has exactly one first sighting, and
// the sources that never produced one for anybody are absent rather than zero.
type SourceSightings struct {
	// Source is who reported the reading each of these positions came from.
	Source SourceID
	// Total is how many of the market's candidates hold a first sighting from it.
	Total int
	// Measured is how many of those the sighting percentile could be computed for.
	Measured int
	// NotMeasured is the census of refusals, by reason. It is a map rather than
	// three fields because the reasons are veto.go's enumeration and a surface that
	// listed a fixed three would silently stop showing a fourth.
	NotMeasured map[VetoUnmeasured]int
}

// TallySightingSources reduces an assessment to one entry per source that produced
// a first sighting, ordered by source id so two screens list them identically.
//
// A verdict with no source on its sighting is skipped rather than filed under an
// empty id: "we do not know which source this came from" is not a source, and a row
// under "" would be read as one.
func TallySightingSources(in []Verdict) []SourceSightings {
	bySource := map[SourceID]*SourceSightings{}
	for _, v := range in {
		id := v.Sighting.Source
		if string(id) == "" {
			continue
		}
		entry, ok := bySource[id]
		if !ok {
			entry = &SourceSightings{Source: id, NotMeasured: map[VetoUnmeasured]int{}}
			bySource[id] = entry
		}
		entry.Total++
		if v.Sighting.Measured {
			entry.Measured++
			continue
		}
		// An unmeasured sighting with no reason would be counted and unattributable,
		// which is the shape this file exists to remove one level up. Sighting.Reason
		// answers NOT_EVALUATED for exactly that case, so the bucket is never empty.
		entry.NotMeasured[v.Sighting.Reason()]++
	}
	out := make([]SourceSightings, 0, len(bySource))
	for _, entry := range bySource {
		out = append(out, *entry)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Source < out[j].Source })
	return out
}
