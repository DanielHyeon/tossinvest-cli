package console

// a063 task 6: the operator surface for lifting an exit snapshot quarantine.
//
// The claims worth pinning on this side are what the operator sees before
// deciding, that the screen asks for a tick and not a typed phrase, and that a
// console wired to an engine without the capability degrades to no button rather
// than to an error.

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// quarantineCommander is a policy commander that also holds quarantines, which
// is the shape the loopback client has against an a063 engine.
type quarantineCommander struct {
	fakePositionPolicyCommander
	mu         sync.Mutex
	rows       []exitquarantine.Row
	listErr    error
	previewErr error
	releaseErr error
	released   []exitquarantine.ApplyRequest
	previewed  []exitquarantine.Request
}

func (q *quarantineCommander) Quarantines(context.Context) ([]exitquarantine.Row, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.listErr != nil {
		return nil, q.listErr
	}
	return append([]exitquarantine.Row(nil), q.rows...), nil
}

func (q *quarantineCommander) PreviewQuarantineRelease(_ context.Context,
	req exitquarantine.Request) (exitquarantine.Preview, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.previewed = append(q.previewed, req)
	if q.previewErr != nil {
		return exitquarantine.Preview{}, q.previewErr
	}
	for _, row := range q.rows {
		if row.PositionID == req.PositionID && row.Version == req.Version {
			return exitquarantine.Preview{Row: row, Capability: "opaque-release-capability", WaitSeconds: 3}, nil
		}
	}
	return exitquarantine.Preview{}, exitquarantine.ErrNotQuarantined
}

func (q *quarantineCommander) ReleaseQuarantine(_ context.Context,
	req exitquarantine.ApplyRequest) (exitquarantine.Result, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.releaseErr != nil {
		return exitquarantine.Result{}, q.releaseErr
	}
	if req.Capability != "opaque-release-capability" {
		return exitquarantine.Result{}, exitquarantine.ErrCapabilityInvalid
	}
	if !req.Confirmed {
		return exitquarantine.Result{}, exitquarantine.ErrConfirmationRequired
	}
	q.released = append(q.released, req)
	row := q.rows[0]
	q.rows = nil
	return exitquarantine.Result{Row: row, ReleasedAt: "2026-08-04T12:00:00Z",
		Evidence: exitquarantine.ComposeEvidence(exitquarantine.ActorLocalOperator, row,
			time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC))}, nil
}

func quarantinedRow() exitquarantine.Row {
	return exitquarantine.Row{
		PositionID: "p-1", Market: "kr", Symbol: "005930", Generation: 1, Version: 1,
		Reason: "ambiguous_recovery", Evidence: "exitpolicy: recovery candidate identity mismatch",
		QuarantinedAt: "2026-08-03T09:03:40Z", Protection: "24929",
	}
}

func quarantineHarness(t *testing.T, commander *quarantineCommander) *dashboardHarness {
	t.Helper()
	return newDashboardHarness(t, func(options *Options) { options.PositionPolicies = commander })
}

func quarantineToken(t *testing.T, page string) string {
	t.Helper()
	match := regexp.MustCompile(`name="quarantine_token" value="([^"]+)"`).FindStringSubmatch(page)
	if len(match) != 2 {
		t.Fatal("quarantine action token not found on the page")
	}
	return match[1]
}

func TestTheConsoleOffersReleaseOnlyForAQuarantinedRow(t *testing.T) {
	commander := &quarantineCommander{
		fakePositionPolicyCommander: fakePositionPolicyCommander{state: managedPolicyState()},
		rows:                        []exitquarantine.Row{quarantinedRow()},
	}
	h := quarantineHarness(t, commander)
	h.authenticate(t)

	page := h.page(t, "/position-management")
	for _, want := range []string{
		"판정 격리 v1", "복구 후보를 하나로 정할 수 없음", "유지 보호선 24929",
		"이 포지션은 지금 exit 판정 대상이 아니다", `name="quarantine_token"`, "판정 격리 해제",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("a quarantined row is missing %q", want)
		}
	}

	// The same screen with no quarantine must offer nothing.
	commander.mu.Lock()
	commander.rows = nil
	commander.mu.Unlock()
	clean := h.page(t, "/position-management")
	if strings.Contains(clean, `name="quarantine_token"`) || strings.Contains(clean, "판정 격리") {
		t.Error("an unquarantined row offers a release")
	}
}

func TestAnUnwiredCommanderShowsNoReleaseAction(t *testing.T) {
	// A plain policy commander: exactly what a console talking to an engine that
	// predates a063 has.
	h := policyHarness(t, &fakePositionPolicyCommander{state: managedPolicyState()})
	h.authenticate(t)

	page := h.page(t, "/position-management")
	if strings.Contains(page, `name="quarantine_token"`) {
		t.Error("a console with no quarantine capability offered a release")
	}

	resp := h.post(t, "/position-management/quarantine/preview", url.Values{
		"csrf": {h.csrf}, "quarantine_token": {"anything"},
	})
	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("preview without the capability = %d, want 501", resp.StatusCode)
	}
}

func TestAFailedQuarantineReadDoesNotBlankTheWholeScreen(t *testing.T) {
	commander := &quarantineCommander{
		fakePositionPolicyCommander: fakePositionPolicyCommander{state: managedPolicyState()},
		listErr:                     errors.New("engine control plane is not answering"),
	}
	h := quarantineHarness(t, commander)
	h.authenticate(t)

	page := h.page(t, "/position-management")
	if !strings.Contains(page, "종목별 정책") {
		t.Error("a secondary read failure took the screen down")
	}
	if !strings.Contains(page, "engine control plane is not answering") {
		t.Error("the screen hid why the quarantine column is blank")
	}
}

func TestTheReleasePreviewShowsWhatIsKeptAndWarnsAboutRequarantine(t *testing.T) {
	commander := &quarantineCommander{
		fakePositionPolicyCommander: fakePositionPolicyCommander{state: managedPolicyState()},
		rows:                        []exitquarantine.Row{quarantinedRow()},
	}
	h := quarantineHarness(t, commander)
	h.authenticate(t)
	token := quarantineToken(t, h.page(t, "/position-management"))

	page := body(t, h.post(t, "/position-management/quarantine/preview", url.Values{
		"csrf": {h.csrf}, "quarantine_token": {token},
	}))

	for _, want := range []string{
		"판정 격리 해제 preview", "v1", "복구 후보를 하나로 정할 수 없음",
		"exitpolicy: recovery candidate identity mismatch", "2026-08-03T09:03:40Z",
		"24929", "해제 후 유지되는 보호선", "진입가", "초기 손절", "high-water",
		"해제는 수리가 아니라 재시도다", "다시 격리된다", "3초",
		`name="confirm"`, `name="capability"`, `name="csrf"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the release preview is missing %q", want)
		}
	}
}

func TestTheReleaseScreenAsksForNoTypedConfirmation(t *testing.T) {
	commander := &quarantineCommander{
		fakePositionPolicyCommander: fakePositionPolicyCommander{state: managedPolicyState()},
		rows:                        []exitquarantine.Row{quarantinedRow()},
	}
	h := quarantineHarness(t, commander)
	h.authenticate(t)
	token := quarantineToken(t, h.page(t, "/position-management"))

	page := body(t, h.post(t, "/position-management/quarantine/preview", url.Values{
		"csrf": {h.csrf}, "quarantine_token": {token},
	}))

	// The user ruled typed confirmation out of this console. The danger gate is
	// the checkbox and the delay; there must be nowhere to type a phrase.
	for _, banned := range []string{
		`type="text"`, `type="number"`, "<textarea", "contenteditable", "<select",
	} {
		if strings.Contains(strings.ToLower(page), strings.ToLower(banned)) {
			t.Errorf("the release screen asks the operator to type: %q", banned)
		}
	}
	if !strings.Contains(page, `type="checkbox"`) {
		t.Error("the release screen has no danger acknowledgement at all")
	}
	if !strings.Contains(page, "문구 입력은 요구하지 않는다") {
		t.Error("the release screen does not say that no phrase is needed")
	}
}

func TestAConfirmedReleaseReportsWhatSurvivedIt(t *testing.T) {
	commander := &quarantineCommander{
		fakePositionPolicyCommander: fakePositionPolicyCommander{state: managedPolicyState()},
		rows:                        []exitquarantine.Row{quarantinedRow()},
	}
	h := quarantineHarness(t, commander)
	h.authenticate(t)
	token := quarantineToken(t, h.page(t, "/position-management"))
	preview := body(t, h.post(t, "/position-management/quarantine/preview", url.Values{
		"csrf": {h.csrf}, "quarantine_token": {token},
	}))
	capability := applyToken(t, preview)

	// The harness client follows the 303, so what lands here is the screen the
	// operator ends up looking at — which is the thing worth asserting.
	landed := body(t, h.post(t, "/position-management/quarantine/apply", url.Values{
		"csrf": {h.csrf}, "capability": {capability}, "confirm": {"yes"},
	}))

	if len(commander.released) != 1 {
		t.Fatalf("released %d times, want once", len(commander.released))
	}
	for _, want := range []string{"판정 격리 해제됨", "005930", "저장된 기준선은 그대로", "다시 격리된다"} {
		if !strings.Contains(landed, want) {
			t.Errorf("the screen after a release is missing %q", want)
		}
	}
	// And the row it released must be gone from the screen it landed on.
	if strings.Contains(landed, `name="quarantine_token"`) {
		t.Error("the released row still offers a release")
	}
}

func TestAnUnconfirmedReleaseIsRefusedInOperatorLanguage(t *testing.T) {
	commander := &quarantineCommander{
		fakePositionPolicyCommander: fakePositionPolicyCommander{state: managedPolicyState()},
		rows:                        []exitquarantine.Row{quarantinedRow()},
	}
	h := quarantineHarness(t, commander)
	h.authenticate(t)
	token := quarantineToken(t, h.page(t, "/position-management"))
	preview := body(t, h.post(t, "/position-management/quarantine/preview", url.Values{
		"csrf": {h.csrf}, "quarantine_token": {token},
	}))

	resp := h.post(t, "/position-management/quarantine/apply", url.Values{
		"csrf": {h.csrf}, "capability": {applyToken(t, preview)},
	})

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unconfirmed apply = %d, want 400", resp.StatusCode)
	}
	page := body(t, resp)
	if !strings.Contains(page, "위험 변경 확인 필요") || !strings.Contains(page, "아무것도 변경되지 않았다") {
		t.Errorf("the refusal does not say what happened: %s", page)
	}
	if len(commander.released) != 0 {
		t.Error("an unconfirmed apply released something")
	}
}

func TestReleaseErrorsAreMappedToOperatorLanguage(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		want   string
	}{
		{"격리 없음", exitquarantine.ErrNotQuarantined, http.StatusNotFound, "활성 격리 없음"},
		{"version 충돌", exitquarantine.ErrVersionMismatch, http.StatusPreconditionFailed, "격리 version 충돌"},
		{"너무 이름", exitquarantine.ErrCapabilityTooEarly, http.StatusTooEarly, "3초 확인 대기 중"},
		{"만료", exitquarantine.ErrCapabilityExpired, http.StatusGone, "격리 해제 capability 만료"},
		{"미배선", exitquarantine.ErrUnwired, http.StatusNotImplemented, "판정 격리 command seam 미배선"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			commander := &quarantineCommander{
				fakePositionPolicyCommander: fakePositionPolicyCommander{state: managedPolicyState()},
				rows:                        []exitquarantine.Row{quarantinedRow()},
				previewErr:                  tc.err,
			}
			h := quarantineHarness(t, commander)
			h.authenticate(t)
			// The token is still minted from the row; the refusal comes from the
			// engine, which is the case an operator actually hits.
			commander.mu.Lock()
			commander.previewErr = nil
			commander.mu.Unlock()
			token := quarantineToken(t, h.page(t, "/position-management"))
			commander.mu.Lock()
			commander.previewErr = tc.err
			commander.mu.Unlock()

			resp := h.post(t, "/position-management/quarantine/preview", url.Values{
				"csrf": {h.csrf}, "quarantine_token": {token},
			})
			if resp.StatusCode != tc.status {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.status)
			}
			if page := body(t, resp); !strings.Contains(page, tc.want) {
				t.Errorf("refusal is missing %q: %s", tc.want, page)
			}
		})
	}
}

func TestAPolicySelectionTokenCannotBeReplayedAsAQuarantineRelease(t *testing.T) {
	// Both tokens are sealed with the same console key. Only the purpose field
	// separates "I clicked 자동관리 해제" from "I clicked 판정 격리 해제", and
	// confusing those two would rebuild the baseline instead of keeping it.
	commander := &quarantineCommander{
		fakePositionPolicyCommander: fakePositionPolicyCommander{state: managedPolicyState()},
		rows:                        []exitquarantine.Row{quarantinedRow()},
	}
	h := quarantineHarness(t, commander)
	h.authenticate(t)
	page := h.page(t, "/position-management")
	policyToken := actionToken(t, page, "자동관리 해제")

	resp := h.post(t, "/position-management/quarantine/preview", url.Values{
		"csrf": {h.csrf}, "quarantine_token": {policyToken},
	})

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("replayed policy token = %d, want 403", resp.StatusCode)
	}
	if len(commander.previewed) != 0 {
		t.Error("a policy token reached the quarantine control plane")
	}
}

var _ = positionpolicy.ActionRelease
