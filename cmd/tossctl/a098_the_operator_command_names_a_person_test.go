//go:build unix

package main

// a098 4.4b-2 — 운영자 표면의 CLI 절반.
//
// 4.4a 가 진 것은 *"이름이 원장에 실제로 들어간다"* 이고, 여기가 지는 것은
// *"그 이름을 사람이 정한다"* 다. R5 가 이름으로 지목한 위험은 빈 actor 가 아니라
// **CLI 가 고정 문자열을 넣어 원장의 검사를 통과하면서 audit trail 만 죽는 것**이다
// (6라운드 P4). 그 위험은 엔진 쪽에서 잴 수 없다 — 인자를 만드는 쪽이 여기다.
//
// 그리고 R6: 승인에 **추가 입력이 없다** (사용자 지시 2026-07-27).

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/spf13/cobra"
)

var a098AlertNow = time.Date(2026, 8, 13, 9, 30, 0, 0, time.UTC)

// a098AlertContent 는 목록이 되풀이하면 안 되는 문자열들이다. 칸마다 하나씩 심는다 —
// 하나만 심으면 「그 칸만 뺀 구현」이 통과한다 (안전 불변식 8).
var a098AlertContent = map[string]string{
	"title":   "CLI-SENTINEL-TITLE-005930",
	"body":    "CLI-SENTINEL-BODY-account-11122233",
	"payload": "CLI-SENTINEL-PAYLOAD-{\"qty\":7}",
}

// a098AlertFixture 는 엔진 하나를 실제로 세운다 — 원장·notifier·게이트·소켓까지.
//
// 가짜 client 로 재면 「CLI 가 보낸 것」만 보이고 **원장에 무엇이 남았는지**는
// 안 보인다. R5 가 재려는 것이 정확히 그 뒤쪽이다.
func a098AlertFixture(t *testing.T) (string, *journal.Journal, *execgw.EntryGate) {
	t.Helper()
	// ⛔ `t.TempDir()` 이 아니라 짧은 경로다. Unix 소켓 경로에는 길이 상한이 있고
	// (sun_path), 테스트 이름이 섞인 임시 경로는 그 상한에 닿는다.
	dir, err := os.MkdirTemp("/tmp", "a098-alerts-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(dir, journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })

	gate := execgw.NewEntryGate(clock.NewFake(a098AlertNow), map[execgw.RequiredQuery]time.Duration{})
	notifier := &obs.Notifier{
		Log:        obs.NewLogger(obs.LogOptions{Writer: new(bytes.Buffer), JSON: true, Clock: clock.System()}),
		Journal:    j,
		Gate:       gate,
		Clock:      clock.System(),
		Attempts:   1,
		RetryDelay: time.Millisecond,
	}
	ops, err := (&engine.Context{Journal: j, Notifier: notifier, Entry: gate}).AlertOperations()
	if err != nil {
		t.Fatalf("AlertOperations: %v", err)
	}
	server, err := engine.StartAlertControlServer(dir, ops)
	if err != nil {
		t.Fatalf("StartAlertControlServer: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return dir, j, gate
}

func a098SeedCLIAlert(t *testing.T, j *journal.Journal, key string) int64 {
	t.Helper()
	id, err := j.EnqueueAlert(context.Background(), journal.Alert{
		EventKey: key,
		Type:     "engine.exit_proposal_refused",
		Severity: "critical",
		Title:    a098AlertContent["title"],
		Body:     a098AlertContent["body"],
		Payload:  a098AlertContent["payload"],
	})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	return id
}

// a098RefusingStdin fails the test the moment anybody reads it.
//
// R6 를 「확인 문구가 출력에 없다」로 재면 약하다 — 문구 없이 한 줄을 기다리는
// 구현도 통과한다. **읽으면 실패**가 그 성질을 직접 재는 형태다.
type a098RefusingStdin struct{ t *testing.T }

func (r a098RefusingStdin) Read([]byte) (int, error) {
	r.t.Error("the command read stdin; 승인에 추가 입력을 요구하면 안 된다 (사용자 지시 2026-07-27)")
	return 0, nil
}

// a098RunAlerts drives the real command tree with a real engine directory.
func a098RunAlerts(t *testing.T, dir string, args ...string) (string, string, error) {
	t.Helper()
	root := &rootOptions{outputFormat: "table", configDir: dir}
	cmd := newEngineCmd(root)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetIn(a098RefusingStdin{t: t})
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), errOut.String(), err
}

// TestTheAlertCommandsAreWiredUnderEngine pins the surface's shape.
//
// ⛔ `--confirm` 이 **없는 것**을 재려면 조회 방법이 실제로 동작한다는 대조군이
// 필요하다. `engine reconcile-resolve` 에는 있으므로 그것으로 확인한다 — 대조군이
// 없으면 이름을 오타 낸 조회도 「없다」로 통과한다.
func TestTheAlertCommandsAreWiredUnderEngine(t *testing.T) {
	var list, ack, resolve *cobra.Command
	for _, c := range leafCommands(newRootCmd()) {
		switch c.CommandPath() {
		case "tossctl engine alerts list":
			list = c
		case "tossctl engine alerts ack":
			ack = c
		case "tossctl engine reconcile-resolve":
			resolve = c
		}
	}
	if list == nil || ack == nil {
		t.Fatal("engine alerts list/ack are not registered; 승인 경로가 어느 명령에도 없다")
	}
	if resolve == nil || resolve.Flags().Lookup("confirm") == nil {
		t.Fatal("the control is missing: reconcile-resolve has no --confirm, so " +
			"「ack 에는 없다」를 이 방법으로는 못 잰다")
	}
	if ack.Flags().Lookup("confirm") != nil || ack.Flags().Lookup("yes") != nil {
		t.Error("ack has a confirmation flag; 추가 승인 마찰을 넣지 않는다 (R6)")
	}
	operator := ack.Flags().Lookup("operator")
	if operator == nil {
		t.Fatal("ack has no --operator")
	}
	if operator.DefValue != "" {
		t.Errorf("--operator defaults to %q — 아무도 안 고른 이름이 원장에 들어간다 (R5)",
			operator.DefValue)
	}
	if ack.Annotations["mutating"] == "true" {
		t.Error("ack declares mutating=true; 이 명령은 주문 경로에 안 닿는다")
	}
}

// TestTheAcknowledgedNameIsTheOneTheOperatorTyped is R5's CLI half.
//
// 화면이 아니라 **원장**에서 본다. CLI 가 고정 문자열을 넣어도 원장의 「공백 아닌
// 이름」 검사는 통과하고 응답도 성공이다 — 그때 죽는 것은 audit trail 뿐이라
// 대조할 곳이 `acknowledged_by` 밖에 없다.
func TestTheAcknowledgedNameIsTheOneTheOperatorTyped(t *testing.T) {
	dir, j, gate := a098AlertFixture(t)
	first := a098SeedCLIAlert(t, j, "a098-cli-first")
	second := a098SeedCLIAlert(t, j, "a098-cli-second")
	gate.Block(execgw.ReasonAlertUndelivered, "critical alerts are undelivered")

	const typed = "박지훈-야간당직"
	out, _, err := a098RunAlerts(t, dir, "alerts", "ack", "--operator", typed)
	if err != nil {
		t.Fatalf("ack: %v (%s)", err, out)
	}
	for _, id := range []int64{first, second} {
		row, err := j.LookupAlert(context.Background(), id)
		if err != nil {
			t.Fatalf("LookupAlert(%d): %v", id, err)
		}
		if row.AcknowledgedBy != typed {
			t.Errorf("alert %d acknowledged_by = %q, want %q — CLI 가 이름을 바꿔 넣었다",
				id, row.AcknowledgedBy, typed)
		}
	}
	if blocks := gate.Blocks(); len(blocks) != 0 {
		t.Errorf("gate.Blocks() = %v — CLI 를 통한 승인이 이 엔진의 진입을 안 풀었다", blocks)
	}
	if !strings.Contains(out, "진입 차단 없음") {
		t.Errorf("output does not tell the operator entry is open: %s", out)
	}
}

// TestAcknowledgingWithoutANameChangesNothing is the other half of R5.
//
// 이름을 안 주면 **엔진까지 가지 않아야 한다.** 원장을 대조로 쓴다: CLI 가 이름을
// 만들어 넣었다면 승인은 성공하고 `acknowledged_by` 가 차 있을 것이다.
func TestAcknowledgingWithoutANameChangesNothing(t *testing.T) {
	for name, args := range map[string][]string{
		"플래그 없음": {"alerts", "ack"},
		"공백 이름":  {"alerts", "ack", "--operator", "   "},
	} {
		t.Run(name, func(t *testing.T) {
			dir, j, gate := a098AlertFixture(t)
			id := a098SeedCLIAlert(t, j, "a098-cli-refused")
			gate.Block(execgw.ReasonAlertUndelivered, "critical alerts are undelivered")

			out, _, err := a098RunAlerts(t, dir, args...)
			if err == nil {
				t.Fatalf("the command accepted an unnamed acknowledgement: %s", out)
			}
			row, lookupErr := j.LookupAlert(context.Background(), id)
			if lookupErr != nil {
				t.Fatalf("LookupAlert: %v", lookupErr)
			}
			if row.AcknowledgedBy != "" {
				t.Errorf("acknowledged_by = %q — CLI 가 이름을 지어냈다 (R5)", row.AcknowledgedBy)
			}
			if _, still := gate.Blocks()[execgw.ReasonAlertUndelivered]; !still {
				t.Error("the refused acknowledgement released the entry latch anyway")
			}
		})
	}
}

// TestAcknowledgingAsksForNothingElse is R6.
//
// stdin 은 읽으면 실패하는 reader 다. 그리고 승인은 **성공해야** 한다 — 실패하는
// 승인은 아무것도 안 읽고도 통과하므로 그것만으로는 R6 를 못 잰다.
func TestAcknowledgingAsksForNothingElse(t *testing.T) {
	dir, j, gate := a098AlertFixture(t)
	a098SeedCLIAlert(t, j, "a098-cli-r6")
	gate.Block(execgw.ReasonAlertUndelivered, "critical alerts are undelivered")

	out, _, err := a098RunAlerts(t, dir, "alerts", "ack", "--operator", "kim.operator")
	if err != nil {
		t.Fatalf("ack: %v (%s)", err, out)
	}
	if !strings.Contains(out, "승인했다") {
		t.Fatalf("the acknowledgement did not report success: %s", out)
	}
	for _, friction := range []string{"y/N", "[y/n]", "다시 입력", "confirm", "Confirm", "입력하세요"} {
		if strings.Contains(out, friction) {
			t.Errorf("output asks for confirmation (%q): %s", friction, out)
		}
	}
}

// TestTheListingShowsTheLeaseAndLeaksNothing is 불변식 8 on the CLI side.
//
// **화면 바이트에서** 본다. 투영이 좁아도 CLI 가 원본을 덧붙이면 새는 곳은 여기다.
// 토큰도 같이 본다 — descriptor 를 읽는 것이 이 명령이므로 찍을 수 있는 위치다.
func TestTheListingShowsTheLeaseAndLeaksNothing(t *testing.T) {
	dir, j, _ := a098AlertFixture(t)
	id := a098SeedCLIAlert(t, j, "a098-cli-listing")
	ctx := context.Background()
	if _, err := j.ClaimAlertByID(ctx, id, "engine.alert_delivery"); err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	}

	out, _, err := a098RunAlerts(t, dir, "alerts", "list")
	if err != nil {
		t.Fatalf("list: %v (%s)", err, out)
	}
	for _, want := range []string{"engine.exit_proposal_refused", "engine.alert_delivery", "1개가 밀려 있다"} {
		if !strings.Contains(out, want) {
			t.Errorf("the listing does not show %q: %s", want, out)
		}
	}
	for column, sentinel := range a098AlertContent {
		if strings.Contains(out, sentinel) {
			t.Errorf("the listing prints the alert's %s: %s", column, out)
		}
	}
	body, err := os.ReadFile(engine.AlertControlDescriptorPath(dir))
	if err != nil {
		t.Fatalf("reading the descriptor: %v", err)
	}
	var descriptor engine.AlertControlDescriptor
	if err := json.Unmarshal(body, &descriptor); err != nil {
		t.Fatalf("decoding the descriptor: %v", err)
	}
	if strings.Contains(out, descriptor.Token) {
		t.Error("the listing prints the bearer token (안전 불변식 8)")
	}
}

// TestAcknowledgingSaysEntryIsStillBlocked is R17's operator-facing half.
//
// 「승인했다」만 찍으면 운영자는 끝난 줄 알고 재시작을 찾는다. 남은 사유를 이름으로
// 말해야 한다.
func TestAcknowledgingSaysEntryIsStillBlocked(t *testing.T) {
	dir, j, gate := a098AlertFixture(t)
	a098SeedCLIAlert(t, j, "a098-cli-r17")
	gate.Block(execgw.ReasonAlertUndelivered, "critical alerts are undelivered")
	gate.Block(execgw.ReasonAlertSenderDown, "the delivery executor stopped")

	out, _, err := a098RunAlerts(t, dir, "alerts", "ack", "--operator", "kim.operator")
	if err != nil {
		t.Fatalf("ack: %v (%s)", err, out)
	}
	if !strings.Contains(out, string(execgw.ReasonAlertSenderDown)) {
		t.Errorf("the operator is not told entry is still blocked: %s", out)
	}
	if strings.Contains(out, "진입 차단 없음") {
		t.Errorf("the command reported entry open while it is latched: %s", out)
	}
}

// TestTheCommandsRefuseWhenNoEngineIsRunning covers the dial refusal.
//
// 이 두 명령이 엔진 없이 「성공」하면 그것은 원장만 고친 것이고, 그때 진입은
// 재시작까지 막힌 채다 — design D7.1 이 없애려는 상태 그 자체다.
//
// # 왜 문자열이 아니라 sentinel 인가 — a109 §2-fix F6 (A2 P2-6)
//
// 이 단정은 원래 `strings.Contains(err.Error(), "엔진이 없다")` 였다. 그것은 **우연한
// 결합**이었다: a098 이 지는 요구는 「거부가 운영자에게 무엇을 할지 말한다」이지 그
// 문장의 글자가 아닌데, a109 가 같은 문구에서 **단정을 조건문으로** 바꾸자
// (「엔진이 없다」 → 「엔진이 없다면 … 돌고 있다면 …」) 이 테스트가 실제로 깨졌다.
// 문구가 더 정직해졌는데 옛 테스트가 그것을 회귀로 신고한 것이다.
//
// 그래서 여기서는 **거부의 정체**(어느 거부인가)를 잰다. 그 거부가 무엇을 말해야
// 하는지 — 강등 가능성·확인할 곳·두 갈래의 행동 — 는 문구의 소유자인 a109 가
// 잰다(`TestTheAlertsCLIDoesNotAssertTheEngineIsAbsent`). 한 사실을 두 테스트가
// 각자의 이유로 잡고, 문구를 고치는 사람은 한 곳만 본다.
func TestTheCommandsRefuseWhenNoEngineIsRunning(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "a098-no-engine-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"alerts", "list"},
		{"alerts", "ack", "--operator", "kim.operator"},
	} {
		out, _, err := a098RunAlerts(t, dir, args...)
		if err == nil {
			t.Fatalf("%v succeeded with no engine running: %s", args, out)
		}
		if !errors.Is(err, errEngineAlertsUnavailable) {
			t.Errorf("%v error = %v; want errEngineAlertsUnavailable — 운영자가 자기 "+
				"경로를 의심하게 만드는 날것의 오류(no such file 따위)로 새어 나왔다", args, err)
		}
	}
}
