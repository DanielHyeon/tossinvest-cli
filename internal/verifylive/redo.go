package verifylive

// redo.go answers one question: which steps is it worth asking the broker about
// again?
//
// # Why the question exists at all
//
// Verdict.Terminal() is true for everything except awaiting-restart, and that is
// deliberate — a resumed run must not place a second live order for a measurement
// already made. But the first live run of this change (2026-07-26, a Sunday)
// recorded nine terminal verdicts and measured almost nothing, for three reasons
// that were all about the day rather than about the account:
//
//	fail      the exchange was closed, so every mutation came back
//	          HTTP 422 order-hours-closed
//	fail      the lazy account-seq resolution was rate limited (429)
//	skipped   the account held no KR symbol, so the sell-side and conditional
//	          steps had nothing to measure against
//
// None of those verdicts is a measurement. They are the tool reporting the
// weather. `tossctl verify run --redo <id>` already exists for a terminal verdict
// worth re-attempting, but the operator drives this change from the console and
// the console is always-resume with no --redo — so the verification was stuck at
// "finished, measured nothing".
//
// # What is and is not in the set
//
//	fail        included. The step ran and did not establish its claim. Re-running
//	            it is the only way to find out whether that was the market or the
//	            broker.
//	skipped     included. The step never ran. If the precondition it was waiting
//	            for now holds — a holding exists, an opt-in was passed — it can.
//	            If it does not, preflight skips it again and nothing is sent, which
//	            is why a skip is safe to re-offer without inspecting why it skipped.
//	pass        excluded. The property was established; re-measuring it would place
//	            a live order to learn something already known.
//	deferred    excluded. The step says up front that this tool cannot drive it at
//	            all (a trigger needs the market to come to the price). Re-running it
//	            produces the same "unverified" line.
//	refused     excluded. A person declined that confirmation. Undoing somebody's
//	            "no" without them asking is not a retry, it is an override.
//	awaiting-   excluded, and it is not in the set because it is not terminal:
//	restart     the ordinary resume already picks it up.
//
// The set is a suggestion, not an authorisation. Everything it names still goes
// through a fresh plan and a fresh batch approval with a new nonce before a single
// request is sent — see runner.go's approveBatch, which does not know this file
// exists.

// RedoSet returns the steps whose newest verdict is one the environment could
// have caused, in catalogue order.
//
// Catalogue order rather than record order: it is what the operator sees on the
// step list and what the runner will walk, so a set presented in any other order
// would read as a different procedure.
func RedoSet(entries []Entry) []StepID {
	var out []StepID
	for _, step := range Steps() {
		e, ok := LastEntry(entries, step.ID)
		if !ok {
			// Never attempted. An ordinary resume runs it; it is not a redo.
			continue
		}
		if RedoableVerdict(e.Verdict) {
			out = append(out, step.ID)
		}
	}
	return out
}

// RedoableVerdict reports a verdict a re-measurement could change.
//
// It is exported next to RedoSet so a surface that has one entry in its hand —
// a progress table marking the rows a redo would touch — asks the same question
// the set does, rather than re-deriving the list of verdicts.
func RedoableVerdict(v Verdict) bool {
	return v == VerdictFail || v == VerdictSkipped
}
