package verifylive

// report.go turns the evidence record into the thing the next three tasks
// actually consume.
//
//	task 2.6   needs a list of properties that were NOT established, because that
//	           list is the set of markets and order types automatic entry stays
//	           forbidden on.
//	task 1.4   needs the established ones as attestation attributes.
//	task 2.9   needs the realised commission, or an honest statement that nothing
//	           filled.
//
// So the report is organised around properties, not around steps. A step is how a
// property got measured; it is not what anybody downstream is asking about. The
// grouping is by observation key prefix (steps.go writes those keys), and a
// property nobody measured appears in the report as "unverified" rather than
// being absent — an attribute that is silently missing is an attribute somebody
// will assume was fine.

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// Attribute is one measured (or unmeasured) property.
type Attribute struct {
	// Key is the observation key, or the property name when nothing measured it.
	Key string `json:"key"`
	// Value is what was observed. Empty means unverified.
	Value string `json:"value,omitempty"`
	// Detail is the observation's own explanation.
	Detail string `json:"detail,omitempty"`
	// Step and At say where the value came from, so a reader can go back to the
	// evidence line.
	Step StepID    `json:"step,omitempty"`
	At   time.Time `json:"at,omitempty"`
	// Verified reports that an observation exists for this property.
	Verified bool `json:"verified"`
}

// Group is a section of the report.
type Group struct {
	Name       string      `json:"name"`
	Tasks      []string    `json:"tasks"`
	Attributes []Attribute `json:"attributes"`
}

// Report is the whole thing.
type Report struct {
	Record      string     `json:"record"`
	AccountRef  string     `json:"account_ref"`
	GeneratedAt time.Time  `json:"generated_at"`
	Steps       []Outcome  `json:"steps"`
	Groups      []Group    `json:"groups"`
	Outstanding []Artifact `json:"outstanding,omitempty"`
	// Unverified is the flat list task 2.6 turns into the no-automatic-entry
	// list. It is duplicated out of Groups on purpose: the one question that
	// gets asked most should not require walking a tree.
	Unverified []string `json:"unverified"`
	// ReplayEnabled reports whether idempotent replay may be used as a
	// response-loss resolution path. It is false unless the record proves
	// otherwise, which is the direction execution-verification requires.
	ReplayEnabled bool `json:"replay_enabled"`
}

// requiredProperties is the checklist.
//
// Every entry is a property some downstream task needs an answer about. Listing
// them here — rather than reporting whatever happens to be in the record — is
// what makes "unverified" a positive statement instead of an omission.
func requiredProperties() []Group {
	return []Group{
		{
			Name:  "Order status derivation (2.1)",
			Tasks: []string{"2.1"},
			Attributes: []Attribute{
				{Key: "order.status.observed"},
				{Key: "order.status.documented_unobserved"},
				{Key: "order.status.cancel_rejected.listed"},
				{Key: "order.status.cancel_rejected.links_original"},
				{Key: "order.status.replace_rejected.listed"},
				{Key: "order.status.replace_rejected.links_original"},
			},
		},
		{
			Name:  "Order path (2.2)",
			Tasks: []string{"2.2"},
			Attributes: []Attribute{
				{Key: "order.place.ok"},
				{Key: "order.cancel.ok"},
				{Key: "order.status.after_place"},
				{Key: "order.status.after_cancel"},
				{Key: "order.canceled_at.present"},
				{Key: "order.amend.ok"},
				{Key: "order.amend.issues_new_id"},
				{Key: "order.amend.original_status"},
				{Key: "sell.boundary.partial_accepted"},
				{Key: "sell.boundary.full_accepted"},
				{Key: "sell.boundary.over_holding_rejected"},
			},
		},
		{
			Name:  "ProtectiveCapability (2.5 → 2.6)",
			Tasks: []string{"2.5", "2.6"},
			Attributes: []Attribute{
				{Key: "conditional.register.market"},
				{Key: "conditional.register.type"},
				{Key: "conditional.register.order_type"},
				{Key: "conditional.register.session"},
				{Key: "conditional.register.ok"},
				{Key: "conditional.read_by_id.ok"},
				{Key: "conditional.list_by_status.contains_new"},
				{Key: "conditional.survives_process_exit"},
				{Key: "conditional.modify_issues_new_id"},
				{Key: "conditional.modify_invalidates_old_id"},
				{Key: "conditional.cancel.ok"},
				{Key: "conditional.cancel.gone_after"},
				{Key: "conditional.trigger_observed"},
				{Key: "conditional.triggered_order_id_exposed"},
				{Key: "conditional.reserves_sellable_quantity"},
			},
		},
		{
			Name:  "Idempotency key (2.7)",
			Tasks: []string{"2.7"},
			Attributes: []Attribute{
				{Key: "idempotency.replay_returns_same_order_id"},
				{Key: "idempotency.no_second_order_created"},
				{Key: "idempotency.conflict_error_code"},
				{Key: "idempotency.conflict_rejected"},
				{Key: "idempotency.conditional_replay_returns_same_id"},
				{Key: "idempotency.ttl_window_closed"},
				{Key: "idempotency.key_scope"},
				{Key: "idempotency.place_round_trip_ms_max"},
			},
		},
		{
			Name:  "Sellable-quantity semantics (2.8)",
			Tasks: []string{"2.8"},
			Attributes: []Attribute{
				{Key: "sellable.baseline.measurable"},
				{Key: "sell.reservation.resting_sell_reserves"},
				{Key: "conditional.reserves_sellable_quantity"},
			},
		},
		{
			Name:  "Realised costs (2.9)",
			Tasks: []string{"2.9"},
			Attributes: []Attribute{
				{Key: "costs.collected"},
				{Key: "costs.commission_total"},
				{Key: "costs.tax_total"},
				{Key: "costs.orders_filled"},
			},
		},
	}
}

// BuildReport reads the record into the report.
func BuildReport(recordPath string, entries []Entry, now time.Time) Report {
	rep := Report{Record: recordPath, GeneratedAt: now.UTC()}

	// The newest observation for a key wins: a resumed run re-measuring a
	// property should not be shadowed by the older answer.
	type located struct {
		obs  Observation
		step StepID
		at   time.Time
	}
	latest := map[string]located{}
	for _, e := range entries {
		if strings.TrimSpace(rep.AccountRef) == "" {
			rep.AccountRef = e.AccountRef
		}
		rep.Steps = append(rep.Steps, Outcome{
			Step: e.StepID, Title: e.Title, Verdict: e.Verdict, Reason: e.Reason,
		})
		for _, o := range e.Observations {
			latest[o.Key] = located{obs: o, step: e.StepID, at: e.FinishedAt}
		}
	}

	for _, group := range requiredProperties() {
		g := Group{Name: group.Name, Tasks: group.Tasks}
		for _, want := range group.Attributes {
			a := Attribute{Key: want.Key}
			if found, ok := latest[want.Key]; ok && isMeasured(found.obs.Value) {
				a.Value, a.Detail, a.Step, a.At, a.Verified = found.obs.Value, found.obs.Detail, found.step, found.at, true
			} else if ok {
				// The step ran and said "unverified"/"unknown" in as many words.
				a.Value, a.Detail, a.Step, a.At = found.obs.Value, found.obs.Detail, found.step, found.at
			}
			if !a.Verified {
				rep.Unverified = append(rep.Unverified, want.Key)
			}
			g.Attributes = append(g.Attributes, a)
		}
		rep.Groups = append(rep.Groups, g)
	}

	rep.Outstanding = Outstanding(entries)
	rep.ReplayEnabled = replayEnabled(latest["idempotency.replay_returns_same_order_id"].obs,
		latest["idempotency.no_second_order_created"].obs)
	return rep
}

// isMeasured rejects the values a step writes to say "I could not establish this".
// They are recorded because the attempt is evidence; they are not answers.
func isMeasured(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "unverified", "unknown", "not-applicable", "refused-by-operator":
		return false
	}
	return true
}

// replayEnabled is the gate execution-verification describes: "검증되지 않은
// 상태에서 멱등 재생을 해소 절차로 사용해서는 안 된다".
//
// Two positives, both required, and the absence of either is a no. It fails
// closed in the strongest sense available: the default return is false and there
// is no branch that reaches true without both observations saying so.
func replayEnabled(sameID, noSecond Observation) bool {
	return strings.EqualFold(sameID.Value, "true") && strings.EqualFold(noSecond.Value, "true")
}

// WriteText renders the report for a terminal.
func (rep Report) WriteText(w io.Writer) {
	fmt.Fprintf(w, "live verification report\n")
	fmt.Fprintf(w, "  record           %s\n", rep.Record)
	fmt.Fprintf(w, "  account          %s\n", orNone(rep.AccountRef))
	fmt.Fprintf(w, "  generated        %s\n", rep.GeneratedAt.Format(time.RFC3339))

	if len(rep.Steps) == 0 {
		fmt.Fprintln(w, "\nNo verification has been recorded yet. Start one with `tossctl verify run`.")
		return
	}

	fmt.Fprintln(w, "\nsteps")
	for _, s := range rep.Steps {
		fmt.Fprintf(w, "  %-22s %-16s %s\n", s.Step, s.Verdict, truncate(s.Reason, 90))
	}

	for _, g := range rep.Groups {
		fmt.Fprintf(w, "\n%s\n", g.Name)
		for _, a := range g.Attributes {
			value := a.Value
			if value == "" {
				value = "unverified"
			}
			mark := " "
			if !a.Verified {
				mark = "!"
			}
			fmt.Fprintf(w, " %s %-48s %s\n", mark, a.Key, value)
			if a.Detail != "" {
				fmt.Fprintf(w, "   %-48s %s\n", "", truncate(a.Detail, 100))
			}
		}
	}

	fmt.Fprintf(w, "\nidempotent replay  %s\n", replayVerdict(rep.ReplayEnabled))
	fmt.Fprintf(w, "unverified         %d propert(ies) — these are task 2.6's no-automatic-entry list\n",
		len(rep.Unverified))
	for _, key := range rep.Unverified {
		fmt.Fprintf(w, "  - %s\n", key)
	}

	if len(rep.Outstanding) > 0 {
		fmt.Fprintf(w, "\n⚠ %d live object(s) this tool created are still on the account:\n", len(rep.Outstanding))
		for _, a := range rep.Outstanding {
			note := a.Note
			if a.Deliberate {
				note = "left on purpose for the persistence check — " + note
			}
			fmt.Fprintf(w, "  %s %s (%s) %s\n", a.Kind, a.ID, a.Symbol, truncate(note, 90))
		}
	}
}

func replayVerdict(enabled bool) string {
	if enabled {
		return "PERMITTED — the record shows a repeated key returning the first order and creating nothing"
	}
	return "DISABLED — response-loss resolution must use the query-based path only until the record proves otherwise"
}

// Progress summarises how far the verification has got, for `verify status`.
type Progress struct {
	Record      string     `json:"record"`
	AccountRef  string     `json:"account_ref"`
	Steps       []Outcome  `json:"steps"`
	Pending     []StepID   `json:"pending"`
	Outstanding []Artifact `json:"outstanding,omitempty"`
	// AwaitingRestart names the step that is waiting for a new process.
	AwaitingRestart StepID `json:"awaiting_restart,omitempty"`
}

// BuildProgress reads the record into the status view.
func BuildProgress(recordPath string, entries []Entry) Progress {
	p := Progress{Record: recordPath}
	for _, e := range entries {
		if strings.TrimSpace(p.AccountRef) == "" {
			p.AccountRef = e.AccountRef
		}
	}
	for _, step := range Steps() {
		e, ok := LastEntry(entries, step.ID)
		if !ok {
			p.Pending = append(p.Pending, step.ID)
			continue
		}
		p.Steps = append(p.Steps, Outcome{
			Step: step.ID, Title: step.Title, Verdict: e.Verdict, Reason: e.Reason,
		})
		if !e.Verdict.Terminal() {
			p.AwaitingRestart = step.ID
			p.Pending = append(p.Pending, step.ID)
		}
	}
	p.Outstanding = Outstanding(entries)
	return p
}

// WriteText renders the progress view.
func (p Progress) WriteText(w io.Writer) {
	fmt.Fprintf(w, "record             %s\n", p.Record)
	if len(p.Steps) == 0 {
		fmt.Fprintln(w, "\nNo verification has been recorded yet. Start one with `tossctl verify run`.")
		fmt.Fprintln(w, "`tossctl verify run --list` prints the whole procedure without touching the account.")
		return
	}
	fmt.Fprintf(w, "account            %s\n", orNone(p.AccountRef))

	fmt.Fprintln(w, "\nsteps")
	for _, s := range p.Steps {
		fmt.Fprintf(w, "  %-22s %-16s %s\n", s.Step, s.Verdict, truncate(s.Reason, 90))
	}
	if len(p.Pending) > 0 {
		fmt.Fprintf(w, "\npending            %s\n", joinSteps(p.Pending))
	}
	if p.AwaitingRestart != "" {
		fmt.Fprintf(w, "\n%s is waiting for a NEW process: run `tossctl verify run --resume`.\n", p.AwaitingRestart)
	}
	if len(p.Outstanding) > 0 {
		fmt.Fprintf(w, "\n⚠ %d live object(s) this tool created are still on the account:\n", len(p.Outstanding))
		for _, a := range p.Outstanding {
			fmt.Fprintf(w, "  %s %s (%s)\n", a.Kind, a.ID, a.Symbol)
		}
		fmt.Fprintln(w, "  `tossctl verify run --resume` finishes the procedure and cancels them.")
	}
}

// WriteSteps prints the catalogue without touching the account. It is what
// `verify run --list` shows, and it is deliberately the same data the runner
// walks — an operator should be able to read exactly what will happen.
func WriteSteps(w io.Writer, includeTTLEdge bool) {
	fmt.Fprintln(w, "live verification procedure")
	fmt.Fprintln(w, "  every step marked [mutating] asks for a typed confirmation immediately before it sends")
	fmt.Fprintln(w, "  every order is one share, LIMIT only, priced far from the market, and cancelled in its own step")
	for _, s := range Steps() {
		tags := []string{}
		if s.Mutates {
			tags = append(tags, "mutating")
		}
		if s.NeedsHolding {
			tags = append(tags, "needs a holding")
		}
		if s.OptIn != "" {
			state := "opt-in " + s.OptIn
			if includeTTLEdge && s.ID == StepIdempotencyTTLEdge {
				state += " (requested)"
			}
			tags = append(tags, state)
		}
		if s.Deferred != "" {
			tags = append(tags, "deferred")
		}
		suffix := ""
		if len(tags) > 0 {
			suffix = "  [" + strings.Join(tags, ", ") + "]"
		}
		fmt.Fprintf(w, "\n%-22s %s%s\n", s.ID, s.Title, suffix)
		fmt.Fprintf(w, "  tasks   %s\n", strings.Join(s.Tasks, ", "))
		fmt.Fprintf(w, "  proves  %s\n", s.Proves)
		for _, line := range s.Procedure {
			fmt.Fprintf(w, "    - %s\n", line)
		}
	}
}

func joinSteps(ids []StepID) string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return strings.Join(out, ", ")
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// sortedAttributeKeys is used by the tests to assert the checklist has no
// duplicate keys inside one group.
func sortedAttributeKeys(g Group) []string {
	out := make([]string, 0, len(g.Attributes))
	for _, a := range g.Attributes {
		out = append(out, a.Key)
	}
	sort.Strings(out)
	return out
}
