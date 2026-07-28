package candidate

// sightingsource.go answers a question the veto tally cannot: which source's
// readings are the ones seen_late keeps refusing.
//
// # The question it exists for
//
// Nobody has measured whether the official rankings endpoint returns the hundred
// rows it is asked for. If it does not, every reading is truncated, every first
// sighting taken from it is refused under READING_TRUNCATED, and the distribution
// the follow-up change would choose a seen_late threshold from does not exist —
// while the screen shows a large unmeasured count and no way to tell which of the
// four ranking sources produced it. The same is true of NEW_ENTRANT_UNKNOWN and
// REQUEST_UNRECORDED, which are refusals about a reading rather than about a symbol.
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
