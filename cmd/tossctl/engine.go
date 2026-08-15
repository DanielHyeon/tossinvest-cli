package main

// engine.go is `tossctl engine run`: the production caller the engine's loops
// never had (openspec change add-engine-runtime, tasks 1.1, 1.2 and 1.4).
//
// # What was missing
//
// internal/app/engine could assemble a whole engine — journal, gateway,
// interlock, reconciliation driver, exit observer — and nothing in cmd/ ever
// asked it to. The loops were tested and unreachable. This file is the boot
// sequence, and the sequence is the requirement rather than an implementation
// detail (engine-safety "엔진 런타임 수명주기", 기동 순서 SHALL):
//
//	1. flock on the journal directory   FIRST, before anything opens the journal
//	2. engine assembly                  config → journal(RW) → official-only broker
//	                                    → obs → startup interlock
//	3. gate off                         refuse: there is no loop set to start
//	4. gate on but unmet                refuse, enumerating the interlock's clauses
//	5. verify runlock fresh             refuse: one account, one rate budget
//	6. active marker                    advisory; the console and autostart read it
//	7. restart recovery, then the loops, under the runtime's two-layer supervision
//
// # Why the lock is first and not "somewhere before the loops"
//
// The journal is a single-writer database with a migration step. A lock taken
// after `journal.Open` leaves the open and the migration outside the exclusion,
// and the case that produces is not theoretical: a binary upgrade followed by the
// autostart script and the console's [엔진 시작] button starting the new engine
// within the same second is exactly two processes migrating one database (review
// round 3).
//
// # Why the gate-off refusal is this file's rule and not the interlock's
//
// The startup interlock is defined only for a gate that is ON — with the gate off
// there are no clauses to be unmet, and reporting a list of them would be
// inventing a refusal the spec does not have. What is true with the gate off is
// simpler and is this change's own rule: every loop in the set refuses to be
// constructed without a verified gate, so there is nothing to run.
//
// # Protection readiness gates entries, not the runtime
//
// ProtectionReady remains UNWIRED until the protective-order change flips it,
// but it no longer prevents reconciliation, exit observation or fill detection
// from running. Instead, the gateway refuses exposure-raising mutations while
// leaving reductions available. Startup still requires a verified automation
// gate and the real RiskGuardian assembled from its configured limits.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"
	"github.com/spf13/cobra"
)

// Refusals this command produces before a single loop starts. They are sentinels
// rather than strings because the console renders them and the tests assert on
// them, and a refusal an operator reads on a dashboard must not drift from the
// one the terminal printed.
var (
	// errEngineGateOff is the gate-off refusal (this change's own rule).
	errEngineGateOff = errors.New(
		"기동할 루프 집합이 없다: engine.automation_gate.enabled = false. " +
			"대사·exit 관측·체결 감지 루프는 전부 검증된 자동화 게이트를 요구한다. " +
			"게이트 flip은 콘솔 밖의 사람 승인 절차(§0.7)다")
	// errEngineInterlockUnmet is the gate-on-but-unmet refusal. The unmet clauses
	// are enumerated after it.
	errEngineInterlockUnmet = errors.New(
		"기동 인터록 미충족: 루프를 하나도 시작하지 않았다")
	// errVerifyInProgress is the rate-budget refusal.
	errVerifyInProgress = errors.New(
		"실계좌 검증이 진행 중이다. 엔진과 검증이 같은 계좌·같은 rate 예산을 다투면 " +
			"사람이 지켜보는 쪽이 손해를 본다 — 검증이 끝난 뒤 다시 시작하라")
)

// engineDefaultMarket is the venue a holding whose payload names no market
// country is recorded under.
//
// The broker's holdings payload carries `marketCountry` for every holding it has
// ever returned, so this is the fallback for a field that should not be empty
// rather than a policy choice. KR because that is the account's home venue and
// the one the engine's costs, sessions and price reads are configured for.
const engineDefaultMarket = "kr"

// engineHardExitCode is what a second stop signal exits with. It is not 0: the
// journal was not closed in an orderly way, and a supervisor reading the exit
// code has to be able to tell that from a clean stop.
const engineHardExitCode = 130

// engineHardExit is the second-signal exit. It is a package variable so this
// package's own tests can drive the signal discipline without ending the test
// binary; there is no flag and no environment variable that reaches it.
var engineHardExit = func(code int) { os.Exit(code) }

func newEngineCmd(root *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "engine",
		Short: "Run the automated trading engine",
	}
	// a098 4.4b — 밀린 critical 알림의 운영자 표면. `engine` 아래에 두는 이유는
	// 두 명령이 **엔진 디렉터리로** 대상을 정하기 때문이다: `engine run` 과 같은
	// `engineJournalDir` 를 쓰므로, 격리 프로파일은 자기 엔진의 알림만 본다.
	cmd.AddCommand(newEngineRunCmd(root), newEngineReconcileResolveCmd(root),
		newEngineAlertsCmd(root))
	return cmd
}

func newEngineRunCmd(root *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the engine's loops until stopped",
		Long: strings.TrimSpace(`
Run the automated trading engine: the reconciliation driver, the exit observation
loop and fill detection, supervised together until the process is stopped.

Startup is fail-closed and in a fixed order. The journal directory is locked
first, so a second engine — from the autostart script, from the console's button,
or from a second terminal — is refused rather than allowed to open the same
single-writer journal. Then the engine is assembled (config, journal, the official
Open API broker, observability) and the automation gate's startup interlock is
consulted. With the gate off there is no loop set to start and the command
refuses. With the gate on and any interlock clause unmet, the unmet clauses are
listed and no loop starts. A live account verification holding the run lock also
refuses the start: one account has one rate budget.

While it runs, an advisory marker in the journal directory is refreshed every
minute so the console can show the engine's state and the autostart script can
check before spawning anything. The marker is not the exclusion — the lock is.

SIGINT or SIGTERM stops it: the loops are cancelled, allowed to finish the cycle
they are in, and the journal is closed cleanly. A second signal exits at once.

Broker-resident protective execution is currently reported separately from the
startup interlock. When it is UNWIRED, the verified runtime may reconcile,
observe exits and detect fills, but the gateway refuses every mutation that
raises exposure. Reduce-only exits remain available.`),
		// official: every read and every order goes through the Open API. mutating:
		// the loops place real reduce-only exits once the gate is verified, which
		// is the whole point of the command.
		Annotations:  map[string]string{"source": "official", "mutating": "true"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEngineRun(cmd, root)
		},
	}
	return cmd
}

// engineJournalDir is the directory the journal, the lock and the marker share.
//
// It follows --config-dir the way the engine's own journal does
// (engine.openEngineJournal), so an isolated profile locks and marks itself and
// cannot exclude — or be excluded by — the real engine.
func engineJournalDir(root *rootOptions) (string, error) {
	if root != nil && strings.TrimSpace(root.configDir) != "" {
		return root.configDir, nil
	}
	path, err := journal.DefaultPath()
	if err != nil {
		return "", fmt.Errorf("engine: resolving the journal location: %w", err)
	}
	return filepath.Dir(path), nil
}

func runEngineRun(cmd *cobra.Command, root *rootOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	out, errOut := cmd.OutOrStdout(), cmd.ErrOrStderr()

	dir, err := engineJournalDir(root)
	if err != nil {
		return err
	}

	// --- 1. the exclusion, first ---------------------------------------------
	lock, err := enginelock.Acquire(dir)
	if err != nil {
		return err
	}
	defer lock.Release()
	fmt.Fprintf(out, "journal lock     %s\n", lock.Path())

	// --- 2. the engine, and 3/4. what the gate says --------------------------
	clk := clock.System()
	logger := obs.NewLogger(obs.LogOptions{Writer: errOut, JSON: true, Clock: clk})
	ectx, err := engineAssemble(ctx, root, clk, logger)
	if err != nil {
		if clauses := engine.UnmetInterlockClauses(err); clauses != nil {
			fmt.Fprintf(errOut, "%v\n", errEngineInterlockUnmet)
			for _, clause := range clauses {
				fmt.Fprintf(errOut, "  - %s\n", clause)
			}
			return errEngineInterlockUnmet
		}
		return err
	}
	defer func() { _ = ectx.Close() }()

	if !ectx.Automation.Verified {
		return errEngineGateOff
	}

	// --- 5. the other process that spends this account's rate budget ---------
	//
	// After the interlock rather than before it, because the order is the spec's:
	// a refusal about safety outranks a refusal about courtesy, and an operator
	// whose gate is misconfigured should be told that whether or not a
	// verification happens to be running.
	if lockPath, verr := engineVerifyLockPath(root); verr == nil {
		if fresh, at := runlock.Fresh(lockPath, clk.Now(), runlock.StaleAfter); fresh {
			fmt.Fprintf(errOut, "%v (%s, 마지막 갱신 %s)\n", errVerifyInProgress, lockPath,
				at.UTC().Format(time.RFC3339))
			return errVerifyInProgress
		}
	}

	// --- 6. the advisory marker ----------------------------------------------
	markerPath := enginelock.MarkerPath(dir)
	marker, merr := enginelock.Hold(ctx, markerPath, clk.Now())
	defer marker.Release()
	// Which run of this pid wrote the marker. The console compares it before it
	// believes a ready signal, because a container recreate can hand the
	// replacement engine the predecessor's pid while the predecessor's marker is
	// still fresh (a102 D4b-2). A build that cannot compute one says nothing, and
	// the console then waits instead of trusting.
	if token, terr := engineProcInstance(os.Getpid()); terr == nil {
		marker.Identify(token)
	}
	if merr != nil {
		// Not a refusal. The exclusion is already held; what a missing marker
		// costs is a status line the console cannot draw.
		fmt.Fprintf(errOut, "note: 엔진 활성 마커를 쓸 수 없다 (%v). 콘솔의 엔진 상태 표시가 비어 보인다.\n", merr)
	} else {
		fmt.Fprintf(out, "active marker    %s (갱신 %s · stale %s · 자문 신호)\n",
			markerPath, enginelock.RefreshEvery, enginelock.StaleAfter)
	}

	// --- 7. the loops --------------------------------------------------------
	//
	// The ready seam is handed down rather than published here because the thing
	// it announces happens inside the runtime: Recover runs before the first loop
	// starts, and only a Recover that finished may say the account is reconciled
	// (a102 D5). Step 6 owns the marker, step 7 owns the moment.
	rt, err := engineRuntimeFactory(ctx, ectx, clk, logger, func() { marker.Ready(clk.Now()) })
	if err != nil {
		return err
	}
	policyCommands, err := engine.NewPositionPolicyCommandService(ectx, clk)
	if err != nil {
		return err
	}
	// a109 D3 — a108의 강등을 형제 endpoint 셋으로 넓힌다.
	//
	// 이 셋도 2026-08-13의 projection과 **같은 병**을 앓았다: 잔재 하나가 Start를
	// 실패시키면 여기서 `return err` 한 줄이 엔진을 죽이고, autostart가 1분마다 다시
	// 세우지만 디스크 상태는 그대로이므로 **영구 기동 루프**가 된다. 그 상태의 실제
	// 이름은 「장중 손절 없음」이다.
	//
	// 강등이 fatal보다 나은 이유는 비교로만 성립한다. 잃는 것을 정직하게 적으면:
	//
	//	policy command   Preview/Apply와 **격리 해제** 표면. 격리 해제는 격리된
	//	                 포지션의 손절 포함 미판정 상태를 푸는 유일한 장중 경로다 —
	//	                 강등은 그 무보호를 **유지**한다. 그래도 fatal(전 포지션
	//	                 무보호)보다 엄격히 낫다.
	//	policy runtime   콘솔·httpapi의 관리 런타임 화면. 조회 전용이다.
	//	alert control    운영자 ack 표면. ack 불가는 미전달 critical의 entry latch를
	//	                 **유지**하므로 신규 진입이 계속 막힌다 — 보수 방향이다.
	//
	// 셋 다 루프의 입력이 아니라 표면이고, 엔진 싱글턴을 강제하는 것은 부팅 1단계의
	// journal flock이지 이 디렉터리들이 아니다.
	policyControl, policyControlErr := enginePositionPolicyCommandStart(dir, policyCommands)
	if policyControl != nil {
		defer policyControl.Close()
	}
	if policyControlErr != nil {
		reportEngineEndpointDegraded(ctx, ectx, errOut,
			enginePolicyCommandEndpoint(dir), policyControlErr)
	}
	policyRuntime, policyRuntimeErr := enginePositionPolicyRuntimeStart(dir, policyCommands)
	if policyRuntime != nil {
		defer policyRuntime.Close()
	}
	if policyRuntimeErr != nil {
		reportEngineEndpointDegraded(ctx, ectx, errOut,
			enginePolicyRuntimeEndpoint(dir), policyRuntimeErr)
	}
	// a108 D3 — 조회 전용 endpoint 하나가 손절을 든 프로세스를 죽이지 않는다.
	//
	// 2026-08-13 23:35 사고: 재부팅 잔재로 이 Start가 실패했고, 여기서 `return err`
	// 한 줄이 엔진을 exit 1 시켰다. autostart가 1분마다 다시 세웠지만 디스크 상태는
	// 그대로였으므로 영구 기동 루프가 됐고, 보호 루프 셋이 US 장중에 전부 정지했다.
	//
	// 강등이 안전한 이유는 하나다: **엔진 싱글턴을 강제하는 것은 부팅 1단계의
	// journal flock이지 이 디렉터리가 아니다.** projection은 콘솔·httpapi가 화면을
	// 그리려고 읽는 `strategyprojection.Reader` 내보내기일 뿐 루프의 입력이 아니므로,
	// 강등으로 잃는 것은 화면이지 판정이 아니다.
	strategyRuntime, projErr := engineStrategyProjectionStart(dir, ectx.StrategyRuntimeProjection())
	if strategyRuntime != nil {
		// nil 수신자에도 안전한 Close이지만, 강등 경로가 Close를 부를 이유는 없다.
		// 조건을 여기 두면 그 판단이 이 파일 안에 남는다.
		//
		// 불변식: 이 Close는 **journal flock을 쥔 채로** 돈다. defer는 LIFO이고
		// 1단계의 `lock.Release()`가 가장 먼저 등록됐으므로 가장 나중에 돈다 —
		// 그래서 이 프로세스의 잔재 제거와 다음 엔진의 회수·발행이 겹칠 수 없다
		// (회수 함수가 flock을 자기 방어로 인용하는 근거가 이 순서다).
		defer strategyRuntime.Close()
	}
	if projErr != nil {
		reportEngineEndpointDegraded(ctx, ectx, errOut, engineProjectionEndpoint(dir), projErr)
	}
	// a098 4.4 — 밀린 critical 알림의 운영자 표면. 승인은 이 프로세스 안에서
	// 일어나야 한다: 진입 게이트는 여기 메모리에 있고, 다른 프로세스가 원장만
	// 고치면 운영자가 승인해도 진입은 재시작까지 막힌 채다 (design D7.1).
	alertOps, err := ectx.AlertOperations()
	if err != nil {
		return err
	}
	alertControl, alertControlErr := engineAlertControlStart(dir, alertOps)
	if alertControl != nil {
		defer alertControl.Close()
	}
	if alertControlErr != nil {
		reportEngineEndpointDegraded(ctx, ectx, errOut, engineAlertControlEndpoint(dir),
			alertControlErr)
	}
	fmt.Fprintf(out, "account          %s\nloops            %s\n",
		ectx.Automation.MaskedAccount(), strings.Join(rt.LoopNames(), ", "))
	fmt.Fprintf(out, "stop             SIGINT/SIGTERM — 루프 완주 후 journal 정합 close. 두 번째 시그널은 즉시 종료\n\n")

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stopWatching := watchStopSignals(cancel, out)
	defer stopWatching()

	return rt.Run(runCtx)
}

// engineStrategyProjectionDegradedEvent 는 강등이 남기는 obs 이벤트 타입이다.
//
// # 이 이름이 obs 등급표에 없는 것이 **설계다** — a108 D3-2
//
// 다음 세 가지를 하지 마라. 셋 다 같은 사고로 이어진다.
//
//  1. 이 이름을 `internal/obs`의 `criticalEvents` 등급표에 올리지 마라.
//  2. severity 를 critical 로 올리지 마라.
//  3. 이 보고를 원장 outbox(`Journal.EnqueueAlert`)에 싣지 마라.
//
// 이유는 하나다. outbox 의 **미전달 행은 다음 부팅의 진입 게이트를 잠근다.**
// `Journal.UndeliveredCount`는 Type 을 가리지 않고 PENDING 행을 세고
// (`internal/journal` 의 `Journal.UndeliveredCount`), `restoreAlertEntryLatch`가 그 수가 0보다
// 크면 `execgw.ReasonAlertUndelivered`로 latch 한다
// (`internal/app/engine` 의 `restoreAlertEntryLatch`). 해제는 운영자 ack 뿐이다. 알림
// publisher 가 설정되지 않은 배포에서 그 행은 **영원히** PENDING 이므로, 결과는
// 「화면 하나를 잃었다는 보고가 실계좌의 신규 진입을 영구 차단한다」가 된다.
//
// obs 교리가 정확히 그것을 금지한다: "measurement failures are never critical …
// 화면의 오탈자가 실계좌 매매를 멈출 수 있어서는 안 된다"
// (`internal/obs` 의 `criticalEvents` 등급표 위 measurement 절). projection 은
// 콘솔·httpapi 가 읽는 조회 전용
// export 이므로 정확히 그 부류다.
//
// (원 D3 은 반대를 요구했고 `internal/execgw`의 park 알림을 선례로 들었다. 그것은
// 선례가 아니었다 — 그 이벤트는 등급표에 **등재돼 있고** entry.Block 이 **의도된**
// 판정이다. A2 적대 리뷰가 오독을 잡아 D3-2 로 결정을 반전시켰다.)
//
// 등급표 미등재 → `obs.SeverityOf`가 Normal 을 준다 → `Notify`는 로그 한 줄과
// best-effort 발행만 하고 원장에 닿지 않는다(`obs.Notifier.publishBestEffort`).
// 나중에 Notifier 에 Publisher 가 붙어도 Normal 인 것이 **의도**다.
const engineStrategyProjectionDegradedEvent obs.EventType = "engine.strategy_projection_unavailable"

// engineControlEndpointDegradedEvent 는 형제 셋(policy command·policy runtime·alert
// control)의 강등이 남기는 obs 이벤트 타입이다 — a109 D3a.
//
// ⛔ 바로 위 `engineStrategyProjectionDegradedEvent` 의 주석이 **이 이름에도 그대로
// 정본**이다. 금지 3종(등급표 등재·critical 승격·원장 outbox 적재)은 여기서도 불변이고,
// 이유도 같다: 미전달 PENDING 행 하나가 다음 부팅의 진입 게이트를 잠근다.
//
// 이름을 하나 더 두는 이유는 운영자가 로그에서 **어느 종류의 표면이 없는가**를 이벤트
// 타입으로도 가를 수 있어야 하기 때문이다. 어느 endpoint 인지는 scope 필드가 말한다.
const engineControlEndpointDegradedEvent obs.EventType = "engine.control_endpoint_unavailable"

// engineEndpoint 는 강등 보고가 「어느 표면인가」를 말하는 데 필요한 최소 좌표다.
//
// 강등이 넷으로 늘면 「강등했다」만 찍는 보고는 쓸모가 없다 — 운영자가 알아야 하는
// 것은 무엇을 잃었고 **어느 디렉터리의 무엇을 지워야 하는가**다.
type engineEndpoint struct {
	// surface 는 운영자가 읽는 표면 이름이다.
	surface string
	// lost 는 이 표면이 없을 때 실제로 잃는 것이다. 「아무것도 안 잃는다」로 적지
	// 않는다 — policy command 는 격리 해제를 잃고, 그것은 무보호의 유지다.
	lost string
	// control 은 원인이 놓여 있는 디렉터리다. obs scope 이자 운영자의 작업 대상이다.
	control string
	// event 는 obs 이벤트 타입이다. **둘 다 등급표에 없다**(위 두 const 주석).
	event obs.EventType
	// title 은 알림 제목이다.
	title string
}

func engineProjectionEndpoint(dir string) engineEndpoint {
	return engineEndpoint{
		surface: "전략 projection", lost: "콘솔·httpapi의 전략 화면",
		control: strategyprojectionrpc.ControlDirectory(dir),
		event:   engineStrategyProjectionDegradedEvent, title: "STRATEGY_PROJECTION_UNAVAILABLE",
	}
}

func enginePolicyCommandEndpoint(dir string) engineEndpoint {
	return engineEndpoint{
		surface: "position policy command",
		// 정직하게 적는다: 격리 해제는 격리된 포지션의 손절 포함 미판정을 푸는 유일한
		// 장중 경로다. 이 표면이 없으면 그 무보호가 **유지**된다.
		lost:    "콘솔의 정책 Preview/Apply와 판정 격리 해제 표면 (격리된 포지션의 미판정이 유지된다)",
		control: positionpolicyrpc.ControlDirectory(dir),
		event:   engineControlEndpointDegradedEvent, title: "POSITION_POLICY_COMMAND_UNAVAILABLE",
	}
}

func enginePolicyRuntimeEndpoint(dir string) engineEndpoint {
	return engineEndpoint{
		surface: "position policy runtime", lost: "콘솔·httpapi의 관리 런타임 화면 (조회 전용)",
		control: positionpolicyrpc.RuntimeControlDirectory(dir),
		event:   engineControlEndpointDegradedEvent, title: "POSITION_POLICY_RUNTIME_UNAVAILABLE",
	}
}

func engineAlertControlEndpoint(dir string) engineEndpoint {
	return engineEndpoint{
		surface: "alert control",
		// ack 불가는 미전달 critical 의 entry latch 를 유지시킨다 — 신규 진입이 계속
		// 막히는 쪽이므로 보수 방향이다.
		lost:    "운영자의 밀린 critical 알림 승인 표면 (미승인 latch는 유지되어 신규 진입이 계속 막힌다)",
		control: engine.AlertControlDirectory(dir),
		event:   engineControlEndpointDegradedEvent, title: "ALERT_CONTROL_UNAVAILABLE",
	}
}

// reportEngineEndpointDegraded 는 강등을 두 곳에 남긴다: 기동 stderr 한 줄과
// obs Normal 이벤트 로그.
//
// # 왜 「재시작하라」고 하지 않는가 — a109 D3/D3b
//
// D1·D2 가 자기 잔재를 원천 소거한 뒤 강등이 실제로 발동하는 원인은 이물·환경 이상
// 뿐이고, 그것들은 **결정적**이다. 재시작만 하면 같은 강등이 그대로 재현된다. 그래서
// 안내는 「원인을 제거한 뒤 재시작하라」여야 한다 — 이 한 줄의 차이가 운영자를
// 무한 재시작에서 꺼낸다.
func reportEngineEndpointDegraded(ctx context.Context, ectx *engine.Context, errOut io.Writer,
	endpoint engineEndpoint, cause error) {
	detail := fmt.Sprintf("%s endpoint를 열지 못했다 (%s): %v", endpoint.surface, endpoint.control, cause)
	fmt.Fprintf(errOut, "note: %s\n"+
		"      엔진은 보호 루프를 그대로 돌린다 — 잃은 것은 %s다.\n"+
		"      원인은 결정적이다: 위 디렉터리의 원인을 제거한 뒤 재시작하라. 재시작만으로는 같은 강등이 재현된다.\n",
		detail, endpoint.lost)
	if ectx == nil || ectx.Notifier == nil {
		return
	}
	// # 보고는 부팅을 기다리게 하지 않는다
	//
	// `Notify`는 Publisher 가 붙어 있으면 **그 자리에서** 발행한다
	// (`obs.Notifier.publishBestEffort`). ntfy publisher 는 네트워크이고 한 번의
	// 발행 상한이 `obs.DefaultPublishTimeout`(10s)다. 이 줄은 `rt.Run` **앞**이므로
	// 그 10초는 손절 루프가 시작되지 않는 10초가 된다 — 화면 하나를 잃었다는 보고가
	// 보호의 시작을 늦추면 안 된다. 그래서 stderr 한 줄만 동기로 남기고 발행은 떼어
	// 보낸다.
	//
	// ctx 를 부모에서 떼어 내는(`WithoutCancel`) 이유는 유실 방지다. 이 goroutine 이
	// 상속할 것은 값이지 취소가 아니다 — 부모 ctx 는 SIGTERM 에서 끊기는데, 그때
	// 취소하면 「종료하면서 남기는 마지막 말」이 늘 사라진다.
	//
	// 원장을 건드리지 않으므로 `ectx.Close()` 와 겹쳐도 안전하다: 이 이벤트는 obs
	// 등급표에 없어서 Normal 이고(`engineStrategyProjectionDegradedEvent` 의 주석),
	// Normal 경로는 로그 한 줄과 best-effort 발행뿐이라 journal handle 에 닿지 않는다.
	// 반환값을 버리는 것이 안전한 이유도 같다 — `Notify`는 **critical** 이벤트를
	// durable 하게 만들지 못했을 때만 오류를 돌려준다(`obs.Notifier.notifyCritical`).
	detached := context.WithoutCancel(ctx)
	go func() {
		_ = ectx.Notifier.Notify(detached, obs.Event{
			Type:   endpoint.event,
			Title:  endpoint.title,
			Body:   detail,
			Fields: map[string]any{obs.FieldScope: endpoint.control},
		})
	}()
}

// engineStrategyProjectionStart is 부팅 7단계가 여는 **조회 전용** projection
// endpoint다. 콘솔과 httpapi가 전략 화면을 그리려고 읽는 unix socket 하나이고,
// 엔진 루프의 입력은 아니다.
//
// engineAssemble·engineRuntimeFactory와 같은 package 변수 seam인 이유도 같다:
// 이 지점의 **실패**를 재려면 실패를 주입할 수 있어야 하는데, 실패를 디스크
// 잔재로 만들면 테스트가 internal/strategyprojectionrpc의 회수 규칙에 묶인다.
// 그러면 회수 규칙을 고칠 때마다 이쪽 테스트가 같이 부서진다 — 재는 것이
// "잔재가 어떻게 생겼는가"가 아니라 "실패했을 때 엔진이 죽는가"인데도.
var engineStrategyProjectionStart = strategyprojectionrpc.Start

// 형제 endpoint 셋의 기동도 같은 package 변수 seam 으로 뺀다 (a109 §2.1).
//
// # 왜 여기에 또 seam 을 파는가
//
// 재려는 것은 "잔재가 어떻게 생겼는가"가 아니라 **"기동이 실패했을 때 엔진이 죽는가"**
// 다. 실패를 디스크 잔재로 만들면 이 파일의 테스트가 `internal/positionpolicyrpc` ·
// `internal/app/engine` 의 회수 규칙에 묶이고, 그 규칙은 a109 가 **지금 바꾸고 있는
// 것**이라 두 작업이 서로를 깨뜨린다. `engineStrategyProjectionStart` 가 a108 에서
// 같은 이유로 생긴 seam 이고, 이것은 그 관례의 형제 셋이다.
//
// ⛔ `internal/app/engine` 의 `cli_testseams.go` 를 관례로 인용하지 마라 — 그 파일은
// 다른 패키지이고, 거기에 seam 을 더하면 이 편집이 그 패키지를 건드리게 된다.
var (
	// command endpoint 만 함수 리터럴로 감싼다. `StartPositionPolicyCommandServer` 의
	// 두 번째 인자 타입이 `internal/app/engine` 의 **비공개** 인터페이스라, 직접 대입한
	// 변수는 테스트가 대체 함수를 **적을 수 없는** 타입이 된다. 호출부가 실제로 넘기는
	// 구체 타입으로 좁혀 두면 그 문제가 사라진다.
	enginePositionPolicyCommandStart = func(dir string,
		commands *engine.PositionPolicyCommandService) (*engine.PositionPolicyCommandServer, error) {
		return engine.StartPositionPolicyCommandServer(dir, commands)
	}
	enginePositionPolicyRuntimeStart = engine.StartPositionPolicyRuntimeServer
	engineAlertControlStart          = engine.StartAlertControlServer
)

// engineAssemble builds the engine profile.
//
// It is a function of its own so the boot sequence above reads as a sequence, and
// so a test can point the whole command at an httptest broker through the same
// seam every other engine test uses.
var engineAssemble = func(ctx context.Context, root *rootOptions, clk clock.Clock,
	logger *obs.Logger) (*engine.Context, error) {
	return assembleEngine(ctx, root, clk, logger, nil)
}

// engineOperator names who started the engine, for the audit trail. It is the OS
// user or nothing; the engine resolves its own default when this is empty.
func engineOperator() string { return "" }

// engineVerifyLockPath is where the live verification's advisory marker lives.
// It is verify.go's own resolution, reached through the same helpers so the two
// commands cannot disagree about which file they mean.
func engineVerifyLockPath(root *rootOptions) (string, error) {
	record, err := resolveVerifyRecord(root, "")
	if err != nil {
		return "", err
	}
	return verifyRunLockPath(record), nil
}

// --- the loop set ------------------------------------------------------------------

// engineRuntime assembles the three safety loops, one inert strategy-entry
// outer loop, and the existing all-or-nothing supervisor over them.
//
// # Why the fill detector is built here and not in internal/app/engine
//
// internal/filldetect imports internal/push for its SSE hint consumer, and
// internal/push imports the WTS client. internal/app/engine's dependency-graph
// test forbids that import transitively — a web-session order mutator must be
// unspellable there — so the detector is assembled at the edge, where the CLI
// already lives with both. The engine package sees it only as an
// [engine.SupervisedLoop] and an [engine.LoopHealth].
//
// # The SLO adapter
//
// exitloop.go declares SLOPressure and says the adapter belongs to the engine's
// wiring precisely so that package does not depend on the detector
// (exitloop.go:144-148). This is that adapter, and wiring it is what makes the
// landed deference rule — 체결 감지 SLO에 양보 — reachable for the first time:
// before this the exit observer was constructed with a nil SLO because there was
// no detector to defer to.
var engineRuntimeFactory = engineRuntime

// engineRecoverySequence is the restart recovery the Recover option runs.
//
// It is a package variable for the same reason engineAssemble and
// engineRuntimeFactory above are, and the file comment on those says it: this
// package's own tests have to be able to drive the whole discipline without a
// broker. Here it buys one specific thing — a test can assemble the *real*
// engineRuntime and then run the Recover option it produced, which is the only
// way to observe that the ready seam survived the assembly.
//
// Without it that step was guarded by a source string, and a102's A2 review
// showed what that is worth: `ready = nil` inserted between the parameter and
// its use left every test green (mutation N5).
var engineRecoverySequence = func(r *reconcile.Recovery) func(context.Context) (reconcile.Report, error) {
	return r.Run
}

// engineRuntime assembles the loop set and hands it to the supervisor, refusing
// at the first constructor that cannot be built.
//
// ready is what the runtime's Recover step calls after it finishes. It is a
// parameter because the marker it writes into belongs to step 6 of the boot
// sequence while the moment it records belongs to step 7 (a102 D5), and the
// runtime offers no way to swap Recover once it is assembled. Tests pass nil;
// recoverThenReady tolerates that.
func engineRuntime(ctx context.Context, ectx *engine.Context, clk clock.Clock, logger *obs.Logger,
	ready func()) (*engine.Runtime, error) {
	detector, err := engineFillDetector(ectx, clk, nil)
	if err != nil {
		return nil, err
	}

	driver, err := ectx.ReconcileDriver(engine.ReconcileDriverOptions{
		Collector:     ectx.SnapshotCollector(clk),
		Clock:         clk,
		DefaultMarket: engineDefaultMarket,
	})
	if err != nil {
		return nil, err
	}

	observer, err := ectx.ExitObserver(engine.ExitObserverOptions{
		Clock:     clk,
		Costs:     costs.DefaultModel(),
		SLO:       detectorPressure{detector: detector},
		Escalate:  ectx.Journal,
		Announcer: ectx.Notifier,
	})
	if err != nil {
		return nil, err
	}

	recovery, err := ectx.Recovery(reconcile.Options{Clock: clk})
	if err != nil {
		return nil, err
	}
	strategyEntry, err := ectx.NewRefreshingPairedStrategyEntrySupervisor(clk)
	if err != nil {
		return nil, err
	}

	// The critical-alert outbox needs somebody to drain it. Before this, a row
	// written while the transport was down stayed PENDING until the same
	// condition happened to be observed again — the alert was durable and
	// undelivered, which is a record rather than an alert (a098).
	//
	// It is Auxiliary, not a supervised loop: a delivery fault must not stop the
	// loop that places a stop-loss.
	alertDelivery, err := ectx.AlertDeliverer(clk)
	if err != nil {
		return nil, err
	}

	return engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: ectx.AccountRef,
		Alerts:     ectx.Notifier,
		Escalate:   ectx.Journal,
		Announcer:  ectx.Notifier,
		Log:        logger,
		Clock:      clk,
		// The recovery's report used to be discarded here, which made §1's
		// rate-limit budget — up to five minutes with the entry gate shut —
		// a silent window (a102 A1 F1). Both consumers live in engineready.go
		// so that this stays a call rather than a judgement: the Recover
		// closure is measured at count=0 (this change's
		// cmd-tossctl--engineruntime/branch-test-map.md).
		Recover: recoverThenReady(engineRecoverySequence(recovery), ready, engineRecoveryObserver(logger)),
		Loops: []engine.SupervisedLoop{
			{
				Name:    "reconcile",
				Run:     driver.Run,
				Health:  driver,
				Trigger: journal.ModeTriggerReconcileCycleFailure,
			},
			{
				// No Health: the exit observer's degradation ladder is its own
				// landed 60-second observation-outage contract, and a second
				// threshold on top would be two definitions of "exit observation
				// is down" that could disagree.
				Name: "exit",
				Run:  observer.Run,
			},
			{
				Name:    "filldetect",
				Run:     detector.Run,
				Health:  detectorHealth{detector: detector},
				Trigger: journal.ModeTriggerFillDetectionFailure,
			},
			strategyEntry.SupervisedLoop(),
		},
		Auxiliary: []engine.AuxiliaryExecutor{alertDelivery},
	})
}

// engineFillDetector builds the polling detector.
//
// hints is the SSE hint consumer and is nil in this build: the engine holds no
// WTS session, and a hint path with nothing feeding it is a goroutine that idles.
// The parameter exists so the assembly-time check below is the real one rather
// than a comment — engine-safety requires an unwired Refresh to be refused *at
// assembly*, not discovered when the loop returns and the defensive termination
// contract reads it as an incident.
func engineFillDetector(ectx *engine.Context, clk clock.Clock,
	hints *filldetect.Hints) (*filldetect.Detector, error) {
	if hints != nil {
		if err := hints.Validate(); err != nil {
			return nil, fmt.Errorf("engine: the fill-detection hint consumer is not routed: %w", err)
		}
	}
	orders := execgw.OfficialOrders{Client: ectx.Official}
	sweep := ectx.AccountSweep()
	return &filldetect.Detector{
		Orders:    orders,
		Order:     orders,
		Positions: sweep,
		Balance:   sweep,
		Tracked: filldetect.JournalTracked{
			Journal: ectx.Journal, AccountRef: ectx.AccountRef,
		},
		Ledger: filldetect.JournalLedger{
			Journal: ectx.Journal, AccountRef: ectx.AccountRef, Refresh: ectx.Reconcile.Refresh,
		},
		Gate:   ectx.Entry,
		Clock:  clk,
		Config: filldetect.Config{Currencies: ectx.SnapshotCurrencies()},
	}, nil
}

// detectorPressure is exitloop.go's SLOPressure, answered by the detector's own
// verdict.
//
// "Behind" is Health().EntryBlocked — the SLO violated past its grace period,
// which is the same condition the detector already uses to hold the entry gate
// shut. One definition, asked twice, rather than a second opinion here about when
// fill detection is late.
type detectorPressure struct{ detector *filldetect.Detector }

func (p detectorPressure) FillDetectionBehind() bool {
	if p.detector == nil {
		return false
	}
	return p.detector.Health().EntryBlocked
}

// detectorHealth is engine.LoopHealth, answered by the outage tracker the
// detector already maintains. Nothing new counts anything: Outage.Consecutive is
// literally "how many cycles in a row have failed".
type detectorHealth struct{ detector *filldetect.Detector }

func (h detectorHealth) ConsecutiveFailures() int {
	if h.detector == nil {
		return 0
	}
	return h.detector.Health().Outage.Consecutive
}

// Compile-time proof that the adapters satisfy what the loops declare.
var (
	_ engine.SLOPressure = detectorPressure{}
	_ engine.LoopHealth  = detectorHealth{}
)

// --- the stop discipline -------------------------------------------------------------

// watchStopSignals implements "두 번째 시그널은 즉시 종료".
//
// The first SIGINT or SIGTERM cancels the runtime, which cancels the loops, waits
// for them to finish the cycle they are in and closes the journal. The second one
// does not wait: an operator sending a second signal is telling the process that
// whatever it is waiting for is not coming, and the journal's own crash recovery
// is designed for exactly the state a hard exit leaves behind.
// The two seams. They are package variables so this package's own tests can drive
// the whole discipline without sending a signal to the test binary; there is no
// flag and no environment variable that reaches them.
var (
	engineSignalNotify = signal.Notify
	engineSignalStop   = signal.Stop
)

// engineStopSignals is the set that means "stop". Both are graceful on the first
// delivery: SIGINT is an operator at a terminal, SIGTERM is the console's [엔진
// 정지] button and any process supervisor.
func engineStopSignals() []os.Signal { return []os.Signal{os.Interrupt, syscall.SIGTERM} }

func watchStopSignals(cancel context.CancelFunc, out io.Writer) func() {
	sigs := make(chan os.Signal, 2)
	engineSignalNotify(sigs, engineStopSignals()...)

	done := make(chan struct{})
	go func() {
		received := 0
		for {
			select {
			case <-done:
				return
			case sig := <-sigs:
				received++
				if received == 1 {
					fmt.Fprintf(out, "\n%v — 루프를 취소하고 완주를 기다린다. 다시 보내면 즉시 종료한다.\n", sig)
					cancel()
					continue
				}
				fmt.Fprintf(out, "%v — 즉시 종료한다. 재기동 시 journal의 복구 절차가 미결 시도를 정리한다.\n", sig)
				engineHardExit(engineHardExitCode)
			}
		}
	}()

	return func() {
		engineSignalStop(sigs)
		close(done)
	}
}
