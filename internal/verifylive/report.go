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
		if !isStepEntry(e) {
			// The batch-approval line is evidence about the run, not about a
			// measured property. It has no step to report and no observation the
			// checklist asks for.
			continue
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
	fmt.Fprintf(w, "실계좌 검증 리포트\n")
	fmt.Fprintf(w, "  기록             %s\n", rep.Record)
	fmt.Fprintf(w, "  계좌             %s\n", orNone(rep.AccountRef))
	fmt.Fprintf(w, "  생성 시각        %s\n", rep.GeneratedAt.Format(time.RFC3339))

	if len(rep.Steps) == 0 {
		fmt.Fprintln(w, "\n아직 기록된 검증이 없다. `tossctl verify run`으로 시작하라.")
		return
	}

	fmt.Fprintln(w, "\n단계")
	for _, s := range rep.Steps {
		fmt.Fprintf(w, "  %-22s %-24s %s\n", s.Step, VerdictLabel(s.Verdict), truncate(s.Reason, 90))
		fmt.Fprintf(w, "  %-22s %s\n", "", StepLabel(s.Step))
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

	fmt.Fprintf(w, "\n멱등 재생          %s\n", replayVerdict(rep.ReplayEnabled))
	fmt.Fprintf(w, "미검증             속성 %d건 — task 2.6의 자동 진입 금지 목록이다\n",
		len(rep.Unverified))
	for _, key := range rep.Unverified {
		fmt.Fprintf(w, "  - %s\n", key)
	}

	if len(rep.Outstanding) > 0 {
		fmt.Fprintf(w, "\n⚠ 이 도구가 만든 객체 %d건이 아직 계좌에 살아 있다:\n", len(rep.Outstanding))
		for _, a := range rep.Outstanding {
			note := a.Note
			if a.Deliberate {
				note = "존속 측정을 위해 의도적으로 남긴 것 — " + note
			}
			fmt.Fprintf(w, "  %s %s (%s) %s\n", a.Kind, a.ID, a.Symbol, truncate(note, 90))
		}
	}
}

func replayVerdict(enabled bool) string {
	if enabled {
		return "허용 — 같은 키 재요청이 첫 주문을 돌려주고 아무것도 새로 만들지 않았음이 기록에 있다"
	}
	return "비활성 — 기록이 달리 증명하기 전까지 응답 유실 해소는 조회 기반 경로만 쓴다"
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
	fmt.Fprintf(w, "기록               %s\n", p.Record)
	if len(p.Steps) == 0 {
		fmt.Fprintln(w, "\n아직 기록된 검증이 없다. `tossctl verify run`으로 시작하라.")
		fmt.Fprintln(w, "`tossctl verify run --list`는 계좌를 건드리지 않고 전체 절차를 출력한다.")
		return
	}
	fmt.Fprintf(w, "계좌               %s\n", orNone(p.AccountRef))

	fmt.Fprintln(w, "\n단계")
	for _, s := range p.Steps {
		fmt.Fprintf(w, "  %-22s %-24s %s\n", s.Step, VerdictLabel(s.Verdict), truncate(s.Reason, 90))
	}
	if len(p.Pending) > 0 {
		fmt.Fprintf(w, "\n남은 단계          %s\n", joinSteps(p.Pending))
	}
	if p.AwaitingRestart != "" {
		fmt.Fprintf(w, "\n%s 단계는 새 프로세스를 기다린다: `tossctl verify run --resume` (콘솔에서는 [콘솔 재시작]).\n",
			p.AwaitingRestart)
	}
	if len(p.Outstanding) > 0 {
		fmt.Fprintf(w, "\n⚠ 이 도구가 만든 객체 %d건이 아직 계좌에 살아 있다:\n", len(p.Outstanding))
		for _, a := range p.Outstanding {
			fmt.Fprintf(w, "  %s %s (%s)\n", a.Kind, a.ID, a.Symbol)
		}
		fmt.Fprintln(w, "  `tossctl verify run --resume`가 절차를 마치고 이들을 취소한다.")
	}
}

// WriteSteps prints the catalogue without touching the account. It is what
// `verify run --list` shows, and it is deliberately the same data the runner
// walks — an operator should be able to read exactly what will happen.
func WriteSteps(w io.Writer, includeTTLEdge, includeTrigger bool) {
	fmt.Fprintln(w, "실계좌 검증 절차")
	writeWrapped(w, "  승인          ", "실행 전체에 대해 타이핑하는 만료 확인 문자열 1개. 무엇이든 전송되기 "+
		"전에 이 실행이 계획한 모든 라이브 요청 — 동작·종목·방향·수량·가격 도출 방식·노출이 끝나는 방식 — 을 "+
		"출력하고 그 한 문자열을 기다린다. 그 외의 입력은 첫 요청 이전에 실행을 중단시킨다. "+
		"`--confirm-each`는 대신 mutation 하나하나 직전에 별도 확인을 받는다.")
	writeWrapped(w, "  그 외 없음    ", "단계는 승인된 목록에 줄이 있는 것만 보낼 수 있다. 목록 밖의 것 — 다른 "+
		"종목, 다른 방향, 승인된 수량 초과 — 을 보내야 하면 적응하는 대신 멈추고 새 승인을 요구한다. "+
		"어느 프롬프트도 대신 답해 주는 플래그는 없다.")
	writeWrapped(w, "  노출          ", "모든 주문은 1주, LIMIT 전용, 시장에서 먼 가격이고, 접수한 단계 안에서 "+
		"취소된다. 조건주문 1건만은 의도적으로 등록된 채 남는다 — 존속 확인이 새 프로세스에서 그것을 다시 "+
		"읽어야 하기 때문이고, 이어하기 실행이 취소한다.")
	fmt.Fprintf(w, "\n  아래 각 단계에 나열된 mutation은 승인 요약에 나타나는 줄 그 자체이며,\n"+
		"  당신 계좌의 종목·수량으로 해석된 것이다.\n")
	for _, s := range Steps() {
		tags := []string{}
		if s.Mutates {
			tags = append(tags, "mutating")
		}
		if s.NeedsHolding {
			tags = append(tags, "보유 필요")
		}
		if s.OptIn != "" {
			state := "옵트인 " + s.OptIn
			asked := (includeTTLEdge && s.ID == StepIdempotencyTTLEdge) ||
				(includeTrigger && s.ID == StepConditionalTrigger)
			if asked {
				state += " (요청됨)"
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
		fmt.Fprintf(w, "\n%-22s %s%s\n", s.ID, StepLabel(s.ID), suffix)
		fmt.Fprintf(w, "  task    %s\n", strings.Join(s.Tasks, ", "))
		fmt.Fprintf(w, "  증명    %s\n", s.Proves)
		for _, line := range s.Procedure {
			fmt.Fprintf(w, "    - %s\n", line)
		}
		if len(s.Mutations) == 0 {
			continue
		}
		fmt.Fprintf(w, "  전송    라이브 요청 %d건, 모두 승인 목록에 나타난다:\n", len(s.Mutations))
		for _, m := range s.Mutations {
			fmt.Fprintf(w, "    · %-22s %s\n", m.Kind, mutationSideAndQuantity(m))
			writeWrapped(w, "        종료  ", orFallback(m.EndsKO, m.Ends))
		}
	}
}

// mutationSideAndQuantity renders a declared mutation's shape without an account to
// resolve it against, which is what --list has to do.
func mutationSideAndQuantity(m StepMutation) string {
	parts := []string{}
	if m.Side != "" {
		parts = append(parts, strings.ToUpper(m.Side))
	}
	switch m.Quantity {
	case QuantityOne:
		parts = append(parts, "1주")
	case QuantityPartial:
		parts = append(parts, "더 큰 보유 중 1주")
	case QuantityWholeHolding:
		parts = append(parts, "매도가능 전량 (--max-sell-quantity 이내)")
	case QuantityOverHolding:
		parts = append(parts, "보유보다 1주 많음 — 거부될 것으로 예상")
	}
	if len(parts) == 0 {
		switch m.Kind {
		case MutateCancelOrder:
			return "이 단계가 남긴 것 전부"
		case MutateCancelConditional:
			return "살아 있는 조건주문"
		default:
			return "자기 방향·수량이 없다"
		}
	}
	return strings.Join(parts, ", ")
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
