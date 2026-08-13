package main

// console.go is `tossctl console`: the local operator console for the one-off
// live-account verification (openspec change verify-execution-capability,
// task 1.6).
//
// # Why the console exists at all
//
// `tossctl verify run` already does the measurement, and it will keep doing it.
// The console exists because the user decided (사용자 결정 2026-07-26) to drive
// the verification from a screen rather than a terminal stream: the batch summary
// is a dozen live requests with prices and reversals, and reading that in a
// scrollback while deciding whether to type a confirmation is worse than reading
// it on a page.
//
// It is a stopgap. internal/console's package doc says so and means it: single
// user, loopback only, deleted when the Phase 4 daemon lands.
//
// # What this file is responsible for
//
// Exactly one thing that matters: it is the only place in the binary that hands
// internal/console a way to reach a live account, and it hands it the same runner
// `verify run` builds, gated on the console's web confirmer instead of the
// terminal's. Everything else — the approval, the session, the rendering — is
// internal/console's.
//
// # The absences, which are the point
//
// There is no flag that presets the session token, answers the approval, or moves
// the console off the loopback interface. `verify run` gains nothing: its
// confirmers are still terminalConfirmer and terminalBatchConfirmer, and
// console_test.go asserts that in the source, because "the CLI must not gain a
// non-interactive approval path" is the condition under which task 1.6 permits the
// web form to exist at all.
//
// --confirm-each is not offered here either. The console is batch-only, and the
// per-mutation confirmer it wires refuses — so the finer gate fails closed rather
// than being silently satisfied if a future edit ever reaches it.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/handoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/localupdate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/JungHoonGhae/tossinvest-cli/internal/version"
	"github.com/spf13/cobra"
)

// These aliases keep every cmd/tossctl dependency on internal/console in this
// integration file. Helper files can implement the narrow seam without creating
// a second package boundary, which TestOnlyConsoleGoReachesTheConsolePackage
// deliberately forbids.
type consoleInstrumentNameReader = console.InstrumentNameReader
type consoleInstrumentRef = console.InstrumentRef
type consoleInstrumentName = console.InstrumentName

// consoleProbeSymbolKR is the KR symbol the buy-side probes are placed against.
//
// It is `verify run --symbol`'s default, restated rather than shared because
// verify.go is a live-order path and this task's scope is additive. Drift is not
// left to review: console_test.go asserts the two agree.
const consoleProbeSymbolKR = "005930"

// consoleProbeSymbol picks the symbol the buy-side probes use in a market.
//
// KR has a default worth pinning: a large, liquid, always-quoted name. US has no
// equivalent constant here on purpose — which US symbols an account can afford to
// place a one-share order against differs per account, so the probe uses a symbol
// the account already holds. Buy and sell probes then share one symbol, which
// keeps the exposure to a single name and makes the opposite-pending rejection
// observable in the same run. No usable US holding means the US probes are
// skipped by the runner's own gates, with the reason recorded.
func consoleProbeSymbol(ctx context.Context, broker verifylive.Broker, market string) string {
	if verifylive.NormalizeMarket(market) != verifylive.MarketUS {
		return consoleProbeSymbolKR
	}
	return firstUsableHoldingIn(ctx, broker, verifylive.MarketUS)
}

type consoleOptions struct {
	port            int
	bind            string
	allowedCIDRs    []string
	publicURL       string
	tlsCert         string
	tlsKey          string
	remoteTokenFile string
	trustedNetwork  bool
}

// HTTP API read aliases keep the Phase 4 daemon on the same narrow console
// seams without creating a second cmd/tossctl file that imports the console
// package. console_test.go intentionally freezes this file as that boundary.
type httpAPIHoldingsReader = console.HoldingsReader
type httpAPIOrdersReader = console.OrdersReader
type httpAPISignalsReader = console.SignalsReader

func newConsoleCmd(root *rootOptions) *cobra.Command {
	opts := &consoleOptions{}

	cmd := &cobra.Command{
		Use:   "console",
		Short: "Serve the local operator console for the live-account verification",
		Long: strings.TrimSpace(`
Serve a small web console for driving the one-off live-account verification
from a browser instead of a terminal.

	By default it binds 127.0.0.1 and prints a URL carrying this process's session
token — it stays
valid until the console stops, and it is not single-use (the single-use one is
the restart handoff token). Opening that URL in this machine's browser is what
authenticates you, so possession of this terminal is the credential. Do not
paste the link anywhere else.

	Remote access is an explicit, all-or-nothing VPN mode. --bind, --allowed-cidr,
	--public-url, --tls-cert, and --tls-key are required together with exactly one
	access mode. --trusted-network uses authenticated VPN membership as the
	application access boundary and presents no login. --remote-token-file retains
	the compatibility login mode. The two access modes cannot be combined.

  overview    /dashboard — engine state, holdings, today's realised P&L, the
              leftovers and the Guardian limits, gathered per market. Read-only:
              no form, and it makes no broker call of its own
  console     /verify-console — soak progress, attestation state, verification
              progress. / redirects to the overview
  positions   what the account holds, joined to the engine's exit lines
  history     completed round trips and the exit judgement stream
  signals     /signals — what the discovery sources have been saying over time,
              with what nobody has checked spelled out rather than left blank. It
              reads the discovery store and calls no source
  verify      the step list, the batch summary, the typed approval, live progress
  report      the measured attributes and the ones still unverified, plus JSON

Account access is read-only except the verification approval: the console places
no order of its own and shows no credential. The settings screen can edit its
explicit operating fields and can install only the fixed sibling
` + "`<current tossctl>.candidate`" + ` after a human reviews its hash and build
identity. It accepts no update path or command, refuses while engine or
verification work is active, preserves a rollback binary, and restarts on the
same loopback port only after a successful replacement.

	Approving still requires the CSRF token the page carries and the expiring
	confirmation string the page shows you, typed back by hand. Native/token-auth
	mode also requires its session token. Trusted-network access removes only
	application login; it does not answer a confirmation or approval for you.

The conditional-order persistence check needs a NEW process, so this console runs
at most one verification per start: when it stops there, quit with Ctrl-C, start
the console again, and press 이어하기.

A step whose last verdict is fail or skipped can be attempted again from the
verify screen's 재측정 button — the market was closed, the broker throttled, the
account held nothing. The set is worked out from the evidence record, never from
the page: a step that passed is never re-measured, and a re-measurement asks for
its own batch approval with a new confirmation string like any other run.

While a verification is running this command marks the account busy so a
concurrent ` + "`tossctl soak run`" + ` delays its cycle rather than spending the same
rate limit.`),
		// official: the verification it drives reaches the Open API. mutating: it
		// can place live orders — through the verify runner, after a typed approval.
		Annotations:  map[string]string{"source": "official", "mutating": "true"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runConsole(cmd, root, opts)
		},
	}

	cmd.Flags().IntVar(&opts.port, "port", 0,
		"Loopback port to serve on; 0 lets the OS pick a free one")
	cmd.Flags().StringVar(&opts.bind, "bind", "",
		"Remote-mode bind IP (requires every remote access flag)")
	cmd.Flags().StringSliceVar(&opts.allowedCIDRs, "allowed-cidr", nil,
		"VPN client CIDR allowed to reach the console; repeat for multiple networks")
	cmd.Flags().StringVar(&opts.publicURL, "public-url", "",
		"Canonical HTTPS URL used by remote browsers, including the port")
	cmd.Flags().StringVar(&opts.tlsCert, "tls-cert", "",
		"PEM TLS certificate file for the remote public URL")
	cmd.Flags().StringVar(&opts.tlsKey, "tls-key", "",
		"PEM TLS private-key file for the remote public URL")
	cmd.Flags().StringVar(&opts.remoteTokenFile, "remote-token-file", "",
		"0600 file containing the remote login token (minimum 32 bytes)")
	cmd.Flags().BoolVar(&opts.trustedNetwork, "trusted-network", false,
		"Trust host loopback or allowed VPN membership; no application login")
	return cmd
}

func runConsole(cmd *cobra.Command, root *rootOptions, opts *consoleOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	// Ctrl-C and the container's SIGTERM have to let account work settle before
	// the socket and any child engine disappear.
	ctx, stop := signal.NotifyContext(ctx, consoleTerminationSignals()...)
	defer stop()

	remote, err := remoteAccessOptions(opts)
	if err != nil {
		return err
	}

	verifyRecord, err := resolveVerifyRecord(root, "")
	if err != nil {
		return err
	}
	verifyRecordUS, err := resolveVerifyRecordFor(root, "", verifylive.MarketUS)
	if err != nil {
		return err
	}
	soakRecord, err := resolveSoakRecord(root, "")
	if err != nil {
		return err
	}
	attestation, err := resolveSoakAttestationPath(root, "")
	if err != nil {
		return err
	}
	openAPISeam, err := newConsoleOpenAPISeam(root, defaultConsoleOpenAPIDeps())
	if err != nil {
		return err
	}

	journalPath, err := consoleJournalPath(root)
	if err != nil {
		// Not fatal. The dashboard's journal half reports "배선되지 않았다" and the
		// verification console — which is what this command is for — is unaffected.
		fmt.Fprintf(cmd.ErrOrStderr(), "원장 경로를 해석할 수 없다 (%v). 포지션·이력 화면은 원장 없이 뜬다.\n", err)
		journalPath = ""
	}
	var optimizationCommander *optimization.Store
	var performanceCapabilities consolePerformanceCapabilities
	if journalPath != "" {
		performanceCapabilities, err = openConsolePerformanceCapabilities(filepath.Dir(journalPath), time.Now)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "성과 DB를 열 수 없다 (%v). 성과·최적화 evidence는 unavailable/read-only로 뜬다.\n", err)
			performanceCapabilities = consolePerformanceCapabilities{}
		} else {
			defer performanceCapabilities.Close()
		}
		optimizationCommander, err = newConsoleOptimizationCommander(ctx, journalPath, performanceCapabilities.Evidence)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "최적화 control plane을 열 수 없다 (%v). 최적화 화면은 조회 전용으로 뜬다.\n", err)
			optimizationCommander = nil
		} else {
			defer optimizationCommander.Close()
		}
	}

	// The engine's advisory marker lives beside its journal. A resolution failure
	// is not fatal for the same reason the journal path's is not: the dashboard
	// reports the engine section as unwired and the verification console — which
	// is what this command is for — is unaffected.
	engineMarkerPath := ""
	engineDir := ""
	if dir, derr := engineJournalDir(root); derr == nil {
		engineDir = dir
		engineMarkerPath = enginelock.MarkerPath(dir)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "엔진 마커 경로를 해석할 수 없다 (%v). 엔진 상태 표시는 비어 있다.\n", derr)
	}

	out := cmd.OutOrStdout()
	var systemUpdater console.SystemUpdater
	var releaseDownloader console.ReleaseDownloader
	var releaseCandidateStager console.ReleaseCandidateStager
	if os.Getenv("TOSSOS_CONTAINER") == "1" {
		fmt.Fprintln(cmd.ErrOrStderr(),
			"컨테이너 실행에서는 시스템 업데이트가 비활성이다. 검증된 image를 교체해 업데이트하라.")
	} else if self, serr := binstamp.SelfPath(); serr != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "시스템 업데이트 경로를 해석할 수 없다 (%v). 설정의 업데이트 섹션은 비활성이다.\n", serr)
	} else {
		cachePath, cerr := resolveUpdateCachePath(root)
		if cerr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "서명 검증 캐시 경로를 해석할 수 없다 (%v). 서명된 릴리스 다운로드는 비활성이다.\n", cerr)
			if updater, uerr := localupdate.New(self); uerr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "시스템 업데이트를 배선할 수 없다 (%v). 설정의 업데이트 섹션은 비활성이다.\n", uerr)
			} else {
				systemUpdater = updater
				releaseCandidateStager = updater
			}
		} else {
			updater, downloader, uerr := assembleConsoleSystemUpdate(
				self, filepath.Join(filepath.Dir(cachePath), "sigstore"), version.Version)
			if updater != nil {
				systemUpdater = updater
				releaseCandidateStager = updater
			}
			if uerr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "서명된 릴리스 다운로드를 배선할 수 없다 (%v). 기존 후보 검토·설치는 계속 사용할 수 있다.\n", uerr)
			} else {
				releaseDownloader = downloader
			}
		}
	}

	var acquireUpdateEngineLock console.AcquireUpdateEngineLock
	if engineDir != "" {
		acquireUpdateEngineLock = func() (func(), error) {
			lock, err := enginelock.Acquire(engineDir)
			if err != nil {
				return nil, err
			}
			return lock.Release, nil
		}
	}

	// The console's one live client is shared by every seam below. Whichever live
	// screen opens first lazily resolves the account and caches that exact official
	// client, so history's /stocks read and every account-scoped screen reuse one
	// OAuth token manager and one account resolution per console process.
	reads := newConsoleBroker(root)
	engineBoot := consoleEngineBootSeam(root)
	var engineBootLoad func() (bool, error)
	if engineBoot != nil {
		engineBootLoad = engineBoot.Load
	}
	engineBootNote := runConfiguredEngineAutostart(
		engineBootLoad,
		func() (string, error) { return startEngine(root) },
	)
	if engineBootNote != "" {
		fmt.Fprintln(cmd.ErrOrStderr(), engineBootNote)
	}

	// The survey, on the same terms and deliberately after the engine: the two
	// share one account's rate budget, and the engine's startup interlock is the
	// side that reads what the survey produces.
	//
	// Both paths below go through restartSoak, which is what carries the profile
	// flags and the record-scoped ownership judgement the spec requires. They differ
	// in two things, and both differences are the point (a101 review):
	//
	//   - boot goes through bootSurvey first, so a survey that is already running
	//     is left alone instead of interrupted. setsid exists so a survey outlives
	//     its console, so that is the designed state here, not an anomaly.
	//   - boot does not pass PrepareSpawn. Clearing the shared token cache is right
	//     for the button, where the operator may have just changed credentials, and
	//     wrong here, where nothing changed: that file is how the engine, the API
	//     daemon and the survey stop taking the token away from each other (a082).
	soakBoot := newConsoleSoakBoot(root)
	var soakBootLoad func() (bool, error)
	if soakBoot != nil {
		soakBootLoad = soakBoot.Load
	}
	startSurvey := func() (string, error) {
		return restartSoak(root, soakRecord, openAPISeam.PrepareSpawn)
	}
	bootSurveyIfAbsent := func() (string, error) {
		return bootSurvey(
			func() ([]int, error) { return soakFindProcesses(soakRecord) },
			func() (string, error) { return restartSoak(root, soakRecord) },
		)
	}
	// a102 §3(D7) — 그 순서를 시간이 아니라 **신호**에 묶는다. a101의 순서가 준 것은
	// engineStartProbe만큼(3초)의 머리 시작뿐이었고, 26페이지 주문 순회는 3초로 끝나지
	// 않는다: 2026-08-13 02:03:29.561Z, 서베이가 probe 만료 2ms 뒤 같은 rate 예산을
	// 때렸고 엔진의 재시작 복구가 429로 죽었다.
	//
	// 판정도 배선도 전부 engineready.go에 있다 — 관측 seam, 시계, 상한·간격까지.
	// 여기 남는 것은 호출 하나뿐이다(a101 규율: runConsole은 0.0%다). ctx는 위에서
	// signal.NotifyContext가 만든 것이므로 콘솔이 내려가면 대기는 시작을 포기한다.
	startSoakAutostartAsync(ctx, engineDir, engineMarkerPath, soakBootLoad, bootSurveyIfAbsent,
		func(note string) { fmt.Fprintln(cmd.ErrOrStderr(), note) })

	// The console receives only an authenticated loopback client. The running
	// engine owns the server and its already-open journal; this process cannot
	// create, migrate, or directly write that journal through this seam.
	var positionPolicyCommander console.PositionPolicyCommander
	var strategyRuntime console.MultiMarketStrategyRuntimeReader
	if engineDir != "" {
		descriptorPath := positionpolicyrpc.DescriptorPath(engineDir)
		if _, statErr := os.Stat(descriptorPath); statErr == nil {
			client, dialErr := positionpolicyrpc.Dial(ctx, descriptorPath)
			if dialErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "엔진 포지션 정책 control plane에 연결할 수 없다 (%v). 정책 화면은 조회 전용으로 뜬다.\n", dialErr)
			} else {
				positionPolicyCommander = &consolePositionPolicyCommander{
					lifecycle: client,
					runtime: positionPolicyRuntimeDescriptorReader{
						descriptorPath: positionpolicyrpc.RuntimeDescriptorPath(engineDir),
					},
				}
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			fmt.Fprintf(cmd.ErrOrStderr(), "엔진 포지션 정책 endpoint를 확인할 수 없다 (%v). 정책 화면은 조회 전용으로 뜬다.\n", statErr)
		}
		strategyDescriptor := strategyprojectionrpc.DescriptorPath(engineDir)
		if _, statErr := os.Stat(strategyDescriptor); statErr == nil {
			client, dialErr := strategyprojectionrpc.Dial(ctx, strategyDescriptor)
			if dialErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "엔진 전략 runtime projection에 연결할 수 없다 (%v). 전략 화면은 dormant로 뜬다.\n", dialErr)
			} else {
				strategyRuntime = client
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			fmt.Fprintf(cmd.ErrOrStderr(), "엔진 전략 runtime endpoint를 확인할 수 없다 (%v). 전략 화면은 dormant로 뜬다.\n", statErr)
		}
	}

	serveErr := console.ListenAndServe(ctx, console.Options{
		Port:                    opts.port,
		Remote:                  remote,
		StartVerify:             consoleVerifyStarter(root),
		SoakRecord:              soakRecord,
		VerifyRecord:            verifyRecord,
		VerifyRecordUS:          verifyRecordUS,
		Attestation:             attestation,
		MinSoakDays:             soak.DefaultCriteria().MinConsecutiveDays,
		RequiredEndpoints:       engine.RequiredEndpoints(),
		Out:                     out,
		SystemUpdater:           systemUpdater,
		ReleaseDownloader:       releaseDownloader,
		ReleaseCandidateStager:  releaseCandidateStager,
		AcquireUpdateEngineLock: acquireUpdateEngineLock,
		CheckUpdateVerifyActivity: func() error {
			return strictVerifyActivity(verifyRunLockPath(verifyRecord), time.Now().UTC())
		},

		// The dashboard's narrow read seams (change add-operator-dashboard). The console
		// reads the journal itself — read-only, per request — and is handed a
		// broker that can do exactly one thing.
		Holdings:        newConsoleHoldings(reads),
		InstrumentNames: newConsoleInstrumentNames(reads),
		// The orders screen's read (change console-orders-screen). One method,
		// behind which this file makes the three calls a refresh costs.
		Orders:           consoleOrdersSeam(reads),
		JournalPath:      journalPath,
		RunLockPath:      verifyRunLockPath(verifyRecord),
		Settings:         consoleSettingsSeam(root),
		ExitPolicies:     consoleExitPolicySettingsSeam(root),
		Optimization:     optimizationCommander,
		PositionPolicies: positionPolicyCommander,
		StrategyRuntime:  strategyRuntime,
		Performance:      performanceCapabilities.Performance,

		// The market schedule is a read-only projection. Keep it in its own
		// capability block so the older dashboard source contract remains stable.
		MarketSchedule: consoleMarketScheduleView{reader: consoleMarketScheduleSeam(root, reads)},

		// The overview's read-only view of the Guardian's ceilings (change
		// console-operator-overview). Five numbers and a currency: this seam
		// displays what the file says and can change none of it.
		GateLimits: consoleGateLimitsSeam(root),

		// The settings screen's editor for those same ceilings (change
		// console-sets-guardian-limits). Separate from the reader above, and its
		// Save takes five numbers and a currency — there is no field on the wire
		// for `enabled`, so the console still cannot open the automation gate.
		Limits: consoleLimitSettingsSeam(root),

		// The operating toggles (change console-owns-the-operating-toggles). Two
		// seams and not one: each writes its own keys, so a policy save cannot
		// carry a stale switch and a switch flip cannot carry stale ceilings.
		// Both append the audit line the hand-edit path never left.
		TradingPolicy: consoleTradingPolicySeam(root),
		Gate:          consoleGateSwitchSeam(root),
		EngineBoot:    engineBoot,

		// Where critical alerts go (change a075). Its Enable takes no argument, so
		// the screen supplies no value: the channel is 128 bits of crypto/rand made
		// here, and the transport's token is never asked for on any form.
		Notifications: consoleNotificationSeam(root),

		// The discovery screen's read (change add-candidate-discovery, task 5.5).
		// It opens internal/candidate's store, runs its assessment and hands over
		// values — no source is called, so an open /signals tab spends none of the
		// account's rate budget and does not become a second discoverer.
		Signals: consoleSignalsSeam(root),

		// The three seams task 1.8 puts behind the console's two restart buttons.
		// internal/console executes nothing: it decides whether the person asking
		// has cleared the session and CSRF gates, and then calls one of these.
		Relaunch: consoleRelaunch(out),
		Handoff:  handoff.New(consoleHandoffPath(verifyRecord)),
		RestartSoak: func() (string, error) {
			// Pressing this is the operator saying "this profile runs the survey".
			// Persisting that is what makes the next container replacement bring it
			// back instead of silently dropping it (a101).
			//
			// It goes through the same spawn gate the boot path takes (a102 D7b):
			// since the boot wait became asynchronous the two can be live at the
			// same time, and both are check-then-act over one record.
			var save func(bool) error
			if soakBoot != nil {
				save = soakBoot.Save
			}
			return guardedSoakRestart(startSurvey, save)
		},
		CheckOpenAPI: func(ctx context.Context) console.OpenAPICredentialCheck {
			return toConsoleOpenAPICredentialCheck(openAPISeam.Check(ctx))
		},
		SaveOpenAPI: func(ctx context.Context, key, secret string) console.OpenAPICredentialCheck {
			return toConsoleOpenAPICredentialCheck(openAPISeam.Save(ctx, key, secret))
		},

		// The engine's status and its two buttons (change add-engine-runtime,
		// task 2.1). Same arrangement as the two restarts above: internal/console
		// decides whether the person asking cleared both gates, and these do the
		// work. The marker is read-only for the console — the exclusion is the
		// flock the engine holds — and neither button can make the engine able to
		// trade: `engine run` re-checks the §0.7-approved gate and the whole
		// startup interlock every time it comes up, and a refusal comes back here
		// as the engine's own words.
		EngineMarker:   engineMarkerPath,
		StartEngine:    func() (string, error) { return startEngine(root) },
		StopEngine:     func() (string, error) { return stopEngine(root) },
		EngineBootNote: engineBootNote,
	})
	return finishConsole(
		os.Getenv("TOSSOS_CONTAINER") == "1",
		serveErr,
		func() (string, error) { return stopEngine(root) },
		cmd.ErrOrStderr(),
	)
}

func toConsoleOpenAPICredentialCheck(result consoleOpenAPIResult) console.OpenAPICredentialCheck {
	return console.OpenAPICredentialCheck{
		State:   console.OpenAPICredentialState(result.State),
		Message: result.Message,
	}
}

// finishConsole gives a container-owned engine its graceful-stop budget while
// preserving the HTTP server's original error. Keeping this policy outside the
// assembly function makes every shutdown branch testable without signals,
// sockets, processes, or a broker session.
func finishConsole(
	container bool,
	serveErr error,
	stopEngine func() (string, error),
	stderr io.Writer,
) error {
	if !container {
		return serveErr
	}
	if stopEngine == nil {
		return serveErr
	}
	note, engineStopErr := stopEngine()
	if strings.TrimSpace(note) != "" {
		fmt.Fprintln(stderr, note)
	}
	if engineStopErr != nil {
		fmt.Fprintf(stderr, "컨테이너 종료 중 엔진 정지 실패: %v\n", engineStopErr)
		if serveErr == nil {
			return engineStopErr
		}
	}
	return serveErr
}

func remoteAccessOptions(opts *consoleOptions) (console.RemoteAccess, error) {
	if opts == nil {
		return console.RemoteAccess{}, nil
	}
	enabled := strings.TrimSpace(opts.bind) != "" || len(opts.allowedCIDRs) != 0 ||
		strings.TrimSpace(opts.publicURL) != "" || strings.TrimSpace(opts.tlsCert) != "" ||
		strings.TrimSpace(opts.tlsKey) != "" || strings.TrimSpace(opts.remoteTokenFile) != "" ||
		opts.trustedNetwork
	if !enabled {
		return console.RemoteAccess{}, nil
	}
	if opts.trustedNetwork && strings.TrimSpace(opts.remoteTokenFile) != "" {
		return console.RemoteAccess{}, errors.New("console: --trusted-network and --remote-token-file cannot be combined")
	}
	if !opts.trustedNetwork && strings.TrimSpace(opts.remoteTokenFile) == "" {
		return console.RemoteAccess{}, errors.New("console: remote mode requires --remote-token-file")
	}
	var token string
	if !opts.trustedNetwork {
		var err error
		token, err = loadRemoteAccessToken(opts.remoteTokenFile)
		if err != nil {
			return console.RemoteAccess{}, err
		}
	}
	return console.RemoteAccess{
		Bind:           opts.bind,
		AllowedCIDRs:   append([]string(nil), opts.allowedCIDRs...),
		PublicURL:      opts.publicURL,
		TLSCertFile:    opts.tlsCert,
		TLSKeyFile:     opts.tlsKey,
		AccessToken:    token,
		TrustedNetwork: opts.trustedNetwork,
		RecordAccess: func(event console.RemoteAccessEvent) error {
			log := openAuditLog()
			if log == nil {
				return errors.New("console: remote access audit log is unavailable")
			}
			return log.RecordAction(event.Action, "console.remote", event.Peer, event.Detail)
		},
	}, nil
}

func loadRemoteAccessToken(path string) (string, error) {
	path = strings.TrimSpace(path)
	info, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("console: reading remote token metadata: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("console: remote token must be a regular file")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return "", fmt.Errorf("console: remote token file %s must not be readable or writable by group/others (use chmod 600)", path)
	}
	if info.Size() > 4096 {
		return "", errors.New("console: remote token file is unexpectedly large")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("console: reading remote token: %w", err)
	}
	token := strings.TrimSpace(string(body))
	if len(token) < 32 {
		return "", errors.New("console: remote token must contain at least 32 bytes")
	}
	return token, nil
}

// --- the dashboard's broker (change add-operator-dashboard, task 1.3) ------------
//
// The console declares a one-method interface (console.HoldingsReader) and this
// is the only place in the binary that satisfies it. What crosses the boundary is
// a *bound method value*, not a broker: the Open API client that owns
// PlaceOrder, CancelOrder and the conditional-order mutations never becomes
// reachable from internal/console, so "the dashboard cannot send an order" is a
// fact about the wiring rather than about the handlers.

// consoleSettingsSeam adapts the adoption-settings seam
// (adoptionsettings.go) to the console's interface. The nil check happens on
// the concrete pointer HERE: a typed-nil inside the interface would defeat the
// console's own `Settings != nil` wiring test.
func consoleSettingsSeam(root *rootOptions) console.AdoptionSettings {
	if s := newAdoptionSettingsSeam(root); s != nil {
		return s
	}
	return nil
}

func consoleExitPolicySettingsSeam(root *rootOptions) console.ExitPolicySettings {
	if s := newExitPolicySettingsSeam(root); s != nil {
		return s
	}
	return nil
}

// consoleLimitSettingsSeam adapts the Guardian-limit editor (limitsettings.go).
// Same nil-on-the-concrete-pointer care as above: a typed-nil inside the
// interface would look wired and the screen would offer controls that panic
// instead of explaining themselves.
func consoleLimitSettingsSeam(root *rootOptions) console.LimitSettings {
	if s := newLimitSettingsSeam(root); s != nil {
		return s
	}
	return nil
}

// consoleTradingPolicySeam and consoleGateSwitchSeam adapt the two operating
// editors (operatingsettings.go, change console-owns-the-operating-toggles).
//
// They live here rather than beside their implementations for the reason the
// package's own guard states: only console.go may import internal/console, so
// the file that names the interface is this one and the files that satisfy it
// stay unaware of the consumer. Same nil-on-the-concrete-pointer care as above.
func consoleTradingPolicySeam(root *rootOptions) console.TradingPolicySettings {
	if s := newTradingPolicySeam(root); s != nil {
		return s
	}
	return nil
}

func consoleGateSwitchSeam(root *rootOptions) console.GateSwitch {
	if s := newGateSwitchSeam(root); s != nil {
		return s
	}
	return nil
}

// consoleNotificationSeam adapts the alert-delivery editor (a075). Same
// nil-on-the-concrete-pointer care as the seams above: a typed nil inside the
// interface would render the card as wired and fail on the first press.
func consoleNotificationSeam(root *rootOptions) console.NotificationSettings {
	if s := newNotificationSeam(root); s != nil {
		return s
	}
	return nil
}

func consoleEngineBootSeam(root *rootOptions) console.EngineBootSettings {
	if s := newConsoleEngineBoot(root); s != nil {
		return s
	}
	return nil
}

// consoleGateLimitsSeam hands the overview screen the Guardian's ceilings, and
// nothing else (change console-operator-overview task 5.1).
//
// It reads, and it stays a reader. The writer is a separate seam
// (consoleLimitSettingsSeam, limitsettings.go) because the overview must not
// gain the ability to change a number it exists to display.
//
// Turning the automation gate ON remains a §0.7 human decision taken outside any
// browser and neither seam can do it; console-sets-guardian-limits separated the
// ceilings from the switch rather than opening both. What crosses this boundary
// is five float64s and a currency string — not the config service.
//
// A console with no resolvable config file gets no seam, and the overview renders
// the limits as seam 미배선 rather than as zero. The same nil-on-the-concrete-type
// care as consoleSettingsSeam: a typed-nil inside the interface would look wired.
func consoleGateLimitsSeam(root *rootOptions) console.GateLimitsReader {
	svc := configServiceFor(root)
	if svc == nil {
		return nil
	}
	return consoleGateLimits{svc: svc}
}

type consoleGateLimits struct{ svc *config.Service }

// consoleGateLimitsTimeout bounds one read of the config file.
//
// The seam takes no context — it is one method and the console holds nothing else
// of the config service — so the deadline is set here. It matters because of what
// is on the other end: the overview reloads itself every 30 seconds and an
// operator leaves it open all day, and a config read on a wedged filesystem with
// context.Background() behind it would hold an HTTP handler open with no bound at
// all. The overview's whole design is that every failure is a sentence on the
// page; a render that never returns is the one failure it has no words for.
const consoleGateLimitsTimeout = 5 * time.Second

// GateLimits satisfies console.GateLimitsReader.
//
// A read failure is returned rather than swallowed: the overview renders an
// unreadable limit as unmeasured with the error beside it, which is neither zero
// nor unlimited — and both of those would be a screen telling an operator
// something nobody read. A timeout arrives as one of those failures.
func (s consoleGateLimits) GateLimits() (console.GateLimits, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consoleGateLimitsTimeout)
	defer cancel()

	cfg, err := s.svc.Load(ctx)
	if err != nil {
		return console.GateLimits{}, err
	}
	gate := cfg.Engine.AutomationGate
	return console.GateLimits{
		MaxOrderQuantity:   gate.MaxOrderQuantity,
		MaxOrderNotional:   gate.MaxOrderNotional,
		MaxTotalExposure:   gate.MaxTotalExposure,
		MaxDailyLossAmount: gate.MaxDailyLossAmount,
		MaxDailyLossRatio:  gate.MaxDailyLossRatio,
		Currency:           gate.LimitCurrency,
	}, nil
}

// --- the console's one account resolution -----------------------------------------

// consoleBroker is the live client every read seam on this console shares.
//
// The client is built lazily on the first live-data screen. That build resolves
// the account once and the resulting official client is then reused by history,
// positions, orders, and the remaining read seams. Each seam used to build its
// own client, so positions followed by /orders paid for the rate-limited account
// read twice. This type is where both the client identity and that resolution
// "once" live.
//
// It widens nothing. A seam is handed this resolver and nothing else, and what
// crosses into internal/console is still one bound method value per screen; the
// client those method values come from was always reachable from this file, since
// a bound method pins its receiver, and it is reachable from the console neither
// before nor after.
type consoleBroker struct {
	root *rootOptions

	gateOnce   sync.Once
	gate       chan struct{}
	client     verifylive.Broker
	accountRef string
	build      func(context.Context, *rootOptions) (verifylive.Broker, string, error)
}

// newConsoleBroker holds the console's live client, and builds nothing yet.
//
// Lazily because client construction reads credentials and account resolution
// makes a network request. Neither should be a precondition for the console coming
// up when the operator may never open a live-data screen. The first relevant
// render pays that cost; a failure is a sentence on the page.
func newConsoleBroker(root *rootOptions) *consoleBroker {
	return &consoleBroker{root: root, build: buildConsoleAccountBroker}
}

func (c *consoleBroker) lock(ctx context.Context) error {
	c.gateOnce.Do(func() {
		c.gate = make(chan struct{}, 1)
		c.gate <- struct{}{}
	})
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.gate:
		return nil
	}
}

func (c *consoleBroker) unlock() { c.gate <- struct{}{} }

// resolve returns the console's live client, building it on the first call.
//
// The build happens under the lock, so two screens opened in the same second make
// one resolution and the second waits for it instead of starting a second
// /api/v1/accounts read. Waiting is the point: what is being serialised is exactly
// the call whose rate limit costs a run its steps.
//
// A failure is returned rather than remembered: credentials that were not there
// when the console started may be there after `tossctl openapi login`, and what
// stops a failing build from being retried on every render is the console's own
// cache (holdings.go bounds attempts by the TTL, not by successes).
func (c *consoleBroker) resolve() (verifylive.Broker, string, error) {
	if err := c.lock(context.Background()); err != nil {
		return nil, "", err
	}
	defer c.unlock()
	if c.client != nil {
		return c.client, c.accountRef, nil
	}
	broker, accountRef, err := verifyBrokerFactory(c.root)
	if err != nil {
		return nil, "", err
	}
	c.client, c.accountRef = broker, strings.TrimSpace(accountRef)
	return c.client, c.accountRef, nil
}

func (c *consoleBroker) instrumentMetadata(ctx context.Context) (consoleInstrumentMetadata, error) {
	if err := c.lock(ctx); err != nil {
		return nil, err
	}
	defer c.unlock()
	if c.client == nil {
		if c.build == nil {
			return nil, fmt.Errorf("console: context-aware broker builder is not configured")
		}
		broker, accountRef, err := c.build(ctx, c.root)
		if err != nil {
			return nil, err
		}
		c.client, c.accountRef = broker, strings.TrimSpace(accountRef)
	}
	reader, ok := c.client.(consoleInstrumentMetadata)
	if !ok {
		return nil, fmt.Errorf("console: this build's broker (%T) has no official stock metadata read", c.client)
	}
	return reader, nil
}

func (c *consoleBroker) TypedMarketCalendar(ctx context.Context, country, date string) (official.MarketCalendarResponse, error) {
	broker, _, err := c.resolve()
	if err != nil {
		return official.MarketCalendarResponse{}, err
	}
	reader, ok := broker.(typedMarketCalendarReader)
	if !ok {
		return official.MarketCalendarResponse{}, fmt.Errorf("console: this build's broker (%T) has no typed official calendar read", broker)
	}
	return reader.TypedMarketCalendar(ctx, country, date)
}

type consoleMarketScheduleView struct{ reader *consoleMarketScheduleReader }

func (v consoleMarketScheduleView) Read(ctx context.Context) (console.MarketScheduleReading, error) {
	status, err := v.reader.Read(ctx)
	if err != nil {
		return console.MarketScheduleReading{}, err
	}
	return console.MarketScheduleReading{
		SchedulerDesired: status.SchedulerDesired, AutoStartDesired: status.AutoStartDesired,
		SchedulerEffective: status.SchedulerEffective, AutoStartEffective: status.AutoStartEffective,
		Market: status.Market, Session: status.Session, ApplyTiming: status.ApplyTiming,
		CalendarSource: status.CalendarSource, CalendarVersion: status.CalendarVersion,
		CalendarFetchedAt: status.CalendarFetchedAt, DecisionReason: status.DecisionReason,
		NextTransition: status.NextTransition,
	}, nil
}

var _ console.MarketScheduleReader = consoleMarketScheduleView{}

// newConsoleHoldings builds the console's holdings reader.
//
// It is handed the shared resolver rather than the root options, which is what
// makes the account resolution behind the positions screen the same one the orders
// screen uses. The laziness is unchanged — it now lives in consoleBroker.resolve.
func newConsoleHoldings(shared *consoleBroker) console.HoldingsReader {
	return &lazyHoldings{shared: shared}
}

// holdingsFunc is the single capability the console is handed.
type holdingsFunc func(ctx context.Context, symbol string) ([]domain.Position, error)

type lazyHoldings struct {
	shared *consoleBroker
}

// Holdings satisfies console.HoldingsReader.
//
// The method value is taken off the shared client per call and dropped again. This
// type holds no broker and no field at all beyond the resolver, so the Open API
// client's PlaceOrder / CancelOrder / ModifyOrder are not reachable from the value
// that crosses into internal/console.
func (l *lazyHoldings) Holdings(ctx context.Context, symbol string) ([]domain.Position, error) {
	broker, _, err := l.shared.resolve()
	if err != nil {
		return nil, err
	}
	var read holdingsFunc = broker.Holdings
	return read(ctx, symbol)
}

// --- the orders screen's read (change console-orders-screen, task 4.6) ------------
//
// The console declares one method and this is the only thing in the binary that
// satisfies it. Behind that one method are the three broker calls one refresh
// costs, which is where they belong: the console cannot then spend the budget
// three times over, and it cannot report one endpoint's silence as another's zero.

// The two values the `status` query parameter of GET /api/v1/orders takes. It is
// documented `required: true`, and the parameter is not a filter in the way the
// other five are — it selects between two differently-shaped answers:
//
//	OPEN    "모든 대기 중 주문을 전량 반환합니다. limit, cursor 는 무시되며"
//	CLOSED  "limit (기본 20, 최대 100), cursor, from/to 파라미터 모두 적용됩니다"
//
// The first implementation of this screen sent neither and asked for limit=100,
// which is either a rejected request (the live count permanently unmeasured) or
// one page over the account's entire history (a live order past row 100 missing
// from the table AND from the count, rendered as "0건 이상"). Both are the failure
// the screen exists to prevent. They are constants rather than literals so the
// call sites read as a choice between two groups and not as a filter somebody may
// tidy away.
const (
	orderGroupOpen   = "OPEN"
	orderGroupClosed = "CLOSED"
)

// consoleOrdersPageLimit bounds one page of the CLOSED list.
//
// It is a page and not the whole history on purpose — the screen is "what is
// alive and what happened today", and an unbounded walk would be a loop of broker
// calls behind one page load. When the broker says there is more, the console
// renders that count as a floor ("N건 이상") rather than as a number, which is why
// truncating here is honest rather than lossy.
//
// It is deliberately NOT sent with the OPEN request. The API ignores it there, so
// sending it would put a bound on the wire that the answer does not have — and the
// live count's whole claim is that it cannot be short.
const consoleOrdersPageLimit = 100

// consoleOrdersReader is the part of the live client the orders screen needs.
//
// verifyBrokerFactory returns a verifylive.Broker, which does not declare the
// raw order reads — it declares OrdersPageRaw, whose orders are undecoded JSON.
// The concrete client has both, so the path is chosen by asserting for exactly
// the two reads on the client the console has already resolved, rather than by
// building a second *official.Client. That second client would resolve the account
// sequence again, and that resolution is the call that came back 429 three times
// on 2026-07-26 and cost a run three steps (measurements.md M4) — the same reason
// this seam takes consoleBroker rather than rootOptions.
type consoleOrdersReader interface {
	OrdersRaw(ctx context.Context, filter official.OrdersFilter) (official.RawOrderList, error)
	ConditionalOrdersRaw(ctx context.Context, status, symbol, cursor string,
		limit int) (official.RawConditionalOrderList, error)
}

// consoleOrdersSeam builds the console's orders reader.
//
// It is handed the same shared resolver newConsoleHoldings gets, so opening this
// screen after the positions screen costs no second account resolution. The
// laziness is unchanged and now lives in one place — consoleBroker.resolve.
func consoleOrdersSeam(shared *consoleBroker) console.OrdersReader {
	return &lazyOrders{shared: shared}
}

type lazyOrders struct {
	shared *consoleBroker
}

// Orders satisfies console.OrdersReader.
//
// The two read method values are taken off the shared client per call and dropped
// again: this type holds no broker, so nothing reachable from the value the
// console was handed has PlaceOrder on it.
//
// The three calls are sequential and each one's outcome is carried separately. A
// failure on one is NOT returned as an error for the set: the console renders that
// part as unmeasured and refuses to add the counts, and collapsing them into one
// error would take the measured parts down with the missing one.
//
// An error is returned only when there is no reading at all to describe — no
// credentials, or a client that does not have these reads.
func (l *lazyOrders) Orders(ctx context.Context) (console.OrdersReading, error) {
	broker, accountRef, err := l.shared.resolve()
	if err != nil {
		return console.OrdersReading{}, err
	}
	reads, ok := broker.(consoleOrdersReader)
	if !ok {
		return console.OrdersReading{}, fmt.Errorf(
			"console: this build's broker (%T) has no raw order reads; the orders screen needs "+
				"the broker's own decimal strings, because a value that has been through float64 "+
				"cannot say whether the broker sent one at all", broker)
	}
	plain, conditional := reads.OrdersRaw, reads.ConditionalOrdersRaw

	var out console.OrdersReading
	out.AccountRef = accountRef

	// The live half. status=OPEN and nothing else: the broker returns every
	// pending order for it, so this list cannot be short and the count taken from
	// it is a number rather than a floor. That is the only call on this screen
	// which structurally cannot miss a leftover, and missing a leftover is what
	// the screen is for.
	open, err := plain(ctx, official.OrdersFilter{Status: orderGroupOpen})
	if err != nil {
		out.OpenError = err.Error()
	} else {
		out.OpenTruncated = open.HasNext
		out.Open = consoleOrderRecords(open.Orders)
	}

	// The finished half. It is a call of its own rather than a saving, because a
	// cancelled or a rejected order never becomes a trade — so /history, which is
	// built from round trips, never shows it. This is the only place that
	// information exists at all.
	closed, err := plain(ctx, official.OrdersFilter{
		Status: orderGroupClosed,
		Limit:  consoleOrdersPageLimit,
	})
	if err != nil {
		out.ClosedError = err.Error()
	} else {
		out.ClosedTruncated = closed.HasNext
		out.Closed = consoleOrderRecords(closed.Orders)
	}

	// The OPEN group, not a per-order status: openapi says the two vocabularies
	// differ, and this filter's job is to fetch exactly the conditionals that are
	// still watching — the ones filling the live-exposure cap.
	conds, err := conditional(ctx, verifylive.ConditionalStatusOpen, "", "", consoleOrdersPageLimit)
	if err != nil {
		out.ConditionalError = err.Error()
	} else {
		out.ConditionalTruncated = conds.HasNext
		out.Conditional = make([]console.ConditionalRecord, 0, len(conds.Orders))
		for _, o := range conds.Orders {
			out.Conditional = append(out.Conditional, console.ConditionalRecord{
				ID: o.ID, Symbol: o.Symbol, Market: o.Market, Kind: o.Type, Status: o.Status,
				Quantity: o.Quantity, TriggerPrice: o.TriggerPrice, OrderPrice: o.OrderPrice,
				ConditionKind: o.ConditionType, Triggered: o.TriggeredOrderID,
				ExpireDate: o.ExpireDate, CreatedAt: o.CreatedAt,
			})
		}
	}

	return out, nil
}

// consoleOrderRecords carries one page of raw orders across the console boundary.
//
// Every value stays the string the broker sent. parseDecimal anywhere on this path
// would put back the zero the whole raw read exists to keep out: the API sends
// price as null for a market order and the entire execution object as null while
// an order is alive, so a converted reading says every live order filled at
// nothing.
func consoleOrderRecords(orders []official.RawOrder) []console.OrderRecord {
	out := make([]console.OrderRecord, 0, len(orders))
	for _, o := range orders {
		out = append(out, console.OrderRecord{
			ID: o.ID, Symbol: o.Symbol, Side: o.Side, Kind: o.OrderType, Status: o.Status,
			Market: o.Market, Currency: o.Currency,
			Quantity: o.Quantity, Price: o.Price,
			FilledQuantity: o.FilledQuantity, AverageFilledPrice: o.AverageFilledPrice,
			OrderedAt: o.OrderedAt, CanceledAt: o.CanceledAt,
		})
	}
	return out
}

// consoleHandoffPath is where the single-use restart token lives: beside the
// evidence record and the soak pause marker, so an isolated --config-dir profile
// gets its own and the default sits in the data directory with the rest of this
// change's state.
func consoleHandoffPath(recordPath string) string {
	return filepath.Join(filepath.Dir(recordPath), handoff.FileName)
}

// consoleRelaunch re-executes this binary so a NEW process instance starts.
//
// That is the whole product: internal/verifylive's conditional-persistence step
// refuses to certify a conditional from the process that registered it, and it
// judges that on process.instance_id — minted fresh at every startup — rather than
// on the PID, which an exec preserves. The restart is therefore a real process
// boundary for the measurement, and the record proves it.
//
// The port is pinned so the browser comes back to the address it is already on.
// Everything else about the command line is kept: the same subcommand, the same
// --config-dir, the same flags the operator typed.
func consoleRelaunch(out io.Writer) console.Relaunch {
	return func(port int) error {
		path, err := binstamp.SelfPath()
		if err != nil {
			return err
		}
		argv := argvWithPort(os.Args, port)
		fmt.Fprintf(out, "  %s\n\n", strings.Join(argv, " "))
		return reexecSelf(path, argv)
	}
}

// consolePortFlag is the flag argvWithPort rewrites. It is named once.
const consolePortFlag = "--port"

// argvWithPort preserves the command line and pins the loopback port.
//
// A console started without --port took whatever the OS offered, and the browser is
// sitting on it. Re-executing with the original argument list would take another
// free port and strand the tab, so the port this process ended up on is written in
// explicitly — which is the one thing about the restart the operator did not choose
// and should not have to.
func argvWithPort(args []string, port int) []string {
	out := make([]string, 0, len(args)+2)
	skipNext := false
	for i, a := range args {
		switch {
		case skipNext:
			skipNext = false
		case i == 0:
			out = append(out, a)
		case a == consolePortFlag:
			skipNext = true // and its value
		case strings.HasPrefix(a, consolePortFlag+"="):
		default:
			out = append(out, a)
		}
	}
	return append(out, consolePortFlag, strconv.Itoa(port))
}

// consoleVerifyStarter builds the runner the console drives.
//
// It is `runVerifyRun`'s wiring with two differences and no others: the batch
// confirmer is the console's instead of the terminal's, and the run is always a
// resumption. The second follows from the first — a console has no --resume flag
// to forget, and the runner already refuses to re-measure a settled step (it skips
// it, and the plan excludes it with the reason on the page), so continuing the
// record is both the safe default and the only sensible one.
//
// redo is `verify run --redo`'s field reached from a button instead of a flag
// (task 1.7). The console computes the set from the evidence record — never from
// the request — and it changes only which steps the runner will walk: the plan is
// rebuilt and a new expiring string still has to be typed before anything is sent.
//
// This is the one place in the console that does NOT take the shared read client,
// and that is deliberate. It builds its own through verifyBrokerFactory on every
// run, which costs one more /api/v1/accounts read than reusing the console's,
// because:
//
//   - The run is about to place real orders, and the account it names in the
//     evidence record is the one it resolved at run start. Reusing a resolution
//     performed whenever the operator happened to open a read screen — possibly
//     hours earlier, possibly before `tossctl openapi login` replaced the
//     credentials — would produce a record naming an account that this run never
//     confirmed. An unconfirmed account on a record of real orders is worse than
//     one extra read.
//   - The resolution's failure contract differs. A read screen renders it as a
//     sentence and stays up; here it is fatal before a person is asked anything,
//     under verifylive's read retry policy (resolveVerifyAccount).
//   - Reads and a verification are different trust contexts. The read seams are
//     handed a client stripped to bound method values; the runner needs the whole
//     broker, and that asymmetry should not be resolved by giving the read path's
//     client the wider job.
//
// The cost is bounded: one resolution per run started, not per screen and not per
// refresh.
func consoleVerifyStarter(root *rootOptions) console.StartVerify {
	return func(
		ctx context.Context,
		confirm verifylive.BatchConfirmer,
		out io.Writer,
		market string,
		redo []verifylive.StepID,
	) (verifylive.Summary, []verifylive.Entry, error) {
		var empty verifylive.Summary

		executionLock, err := acquireVerifyExecutionLock(root)
		if err != nil {
			return empty, nil, err
		}
		defer executionLock.Release()
		fmt.Fprintf(out, "execution lock   %s (engine · update · verification exclusion)\n",
			executionLock.Path())

		market = verifylive.NormalizeMarket(market)
		// The record is resolved per run rather than captured once, because it is
		// the market that decides which file this run's verdicts belong in.
		recordPath, err := resolveVerifyRecordFor(root, "", market)
		if err != nil {
			return empty, nil, err
		}
		prior, err := verifylive.LoadEntries(recordPath)
		if err != nil {
			return empty, nil, err
		}
		releaseIntent, err := holdVerifyRateBudgetIntent(ctx, out, root)
		if err != nil {
			return empty, nil, err
		}
		defer releaseIntent()
		budgetLease, err := acquireVerifyRateBudget(ctx, out, root)
		if err != nil {
			return empty, nil, err
		}
		defer budgetLease.Release()
		broker, accountRef, err := verifyBrokerFactory(root)
		if err != nil {
			return empty, nil, err
		}

		recorder, err := verifylive.OpenRecorder(recordPath)
		if err != nil {
			return empty, nil, err
		}
		defer recorder.Close()

		runner, err := verifylive.New(verifylive.Options{
			Broker:          broker,
			Recorder:        recorder,
			Confirm:         consoleMutationConfirmer(),
			ConfirmBatch:    confirm,
			ApprovalChannel: verifylive.ApprovalChannelConsoleClick,
			Out:             out,
			AccountRef:      accountRef,
			Market:          market,
			Symbol:          consoleProbeSymbol(ctx, broker, market),
			HoldingSymbol:   firstUsableHoldingIn(ctx, broker, market),
			Offset:          verifylive.DefaultOffset,
			MaxSellQuantity: verifylive.DefaultMaxSellQuantity,
			TTLWait:         verifylive.DefaultTTLWait,
			Redo:            redo,
			Prior:           prior,
		})
		if err != nil {
			return empty, nil, err
		}

		fmt.Fprintf(out, "evidence record  %s\n", recordPath)

		summary, runErr := runner.Run(ctx)
		if runErr != nil && (errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded)) {
			// Ctrl-C on the console. Everything recorded stands, and the summary
			// already names whatever is still live.
			runErr = nil
		}
		return summary, runner.Entries(), runErr
	}
}

// consoleMutationConfirmer is the per-mutation gate the console does not offer.
//
// verifylive.Options requires a non-nil Confirmer and refuses a nil one rather
// than defaulting to something permissive; this is the value that satisfies it.
// The console never sets ConfirmEach, so nothing calls it — and if a later edit
// ever did, it refuses, which is the direction a mistake here has to fail in.
func consoleMutationConfirmer() verifylive.Confirmer {
	return func(verifylive.Mutation) error { return verifylive.ErrNotATerminal }
}

// --- the discovery screen (change add-candidate-discovery, task 5.5) ---------------
//
// The seam below hands /signals a read of the discovery store, and nothing else.
// It lives in this file because console_test.go's
// TestOnlyConsoleGoReachesTheConsolePackage allows exactly one importer of
// internal/console: a second one would be a second place the web confirmer could
// be wired, and reviewing one file is the only way that claim stays checkable.
//
// # It reads and it never scans
//
// candidate.Assess loads what is already stored and calls no source. That is the
// whole arrangement: spec Requirement 7 ranks the account's rate budget engine >
// verification > discovery, and a browser tab reloading every fifteen seconds into
// a second discoverer would compete with `tossctl candidate watch` for the
// undocumented RANKING limit — where one 429 costs the entire source, because the
// rankings have no WTS fallback (D14 decision 2). It would also write a second
// near-duplicate observation per symbol per tick into the series the acceleration
// is differenced from.
//
// # And it opens the store per read rather than holding it
//
// Store.Checkpoint's own documentation names this screen as the long-lived reader
// that stops `wal_checkpoint(TRUNCATE)` from reclaiming anything — and the write-
// ahead log grows beside the table on the filesystem the order ledger writes to
// (D16). A console that held the store open would quietly disable the cleanup the
// watch loop runs every turn. Opening for the length of one render costs a file
// open every refresh and keeps that sweep working.

// consoleSignalsMarkets is what the screen reports on, in the contract's order
// (design.md 결정된 계약값: markets.KR and markets.US, both enabled from the start).
var consoleSignalsMarkets = []string{candidate.MarketKR, candidate.MarketUS}

// consoleSignalsPanelWhy is why the screen cannot name the sources a scan lost.
//
// ScanResult.Missing is the only thing that carries a source id together with the
// reason it did not answer, and it lives for the length of one cycle in the
// process that ran it. The console is not that process and the store persists no
// scan record, so this reading has no names to give — and saying so is the whole
// of the rule task 5.7 states: a degradation nobody can attribute is a display
// nobody can act on, and inventing an attribution would be worse than admitting
// the gap. See openspec/changes/add-candidate-discovery/issues.md.
//
// What IS on the screen is the store's own per-candidate completeness — attempted,
// answered and degraded, as recorded by the last scan that touched each row — so
// the page is not silent about degradation, only about which source caused it.
const consoleSignalsPanelWhy = "이 콘솔 프로세스는 스캔을 돌리지 않는다. " +
	"빠진 원천의 이름과 사유는 스캔 결과에만 있고 저장소에 남지 않으므로, " +
	"`tossctl candidate scan`이나 `tossctl candidate watch`의 출력에서 확인한다. " +
	"아래 후보별 완전성은 저장소가 기록한 마지막 스캔의 값이다."

// consoleSignals is the seam's implementation.
type consoleSignals struct {
	// open resolves the discovery store and hands back the release for it. It is a
	// field rather than a direct call so this package's tests can point it at a
	// temporary one.
	//
	// It takes the caller's context because Open migrates: the schema check, the
	// ladder behind it and every query afterwards belong to the request that asked
	// for the page, and this page reloads itself every fifteen seconds. A request
	// the browser has already abandoned used to keep all of it running under
	// context.Background().
	open func(ctx context.Context) (*candidate.Store, func(), error)
}

// consoleSignalsSeam wires the discovery screen.
//
// What crosses the boundary is a value: verdicts, tallies and a panel report.
// internal/candidate's Store — which can promote, cool and prune — never becomes
// reachable from internal/console, so "the discovery screen changes nothing" is a
// fact about the wiring rather than about the handlers.
func consoleSignalsSeam(root *rootOptions) console.SignalsReader {
	return &consoleSignals{
		open: func(ctx context.Context) (*candidate.Store, func(), error) {
			return candidateStoreFactory(ctx, root)
		},
	}
}

// Signals reads every contract market's assessment as of one instant.
func (s *consoleSignals) Signals(ctx context.Context) (console.SignalsReading, error) {
	store, release, err := s.open(ctx)
	if err != nil {
		return console.SignalsReading{}, fmt.Errorf("candidate: opening the discovery store: %w", err)
	}
	defer release()

	// One instant for every market, taken from the store's clock rather than this
	// process's: every input age on the page is measured against it, and two
	// markets read at two instants would answer the same question differently.
	at := store.Now()
	out := console.SignalsReading{At: at}
	for _, market := range consoleSignalsMarkets {
		out.Markets = append(out.Markets, consoleSignalsMarket(ctx, store, market, at))
	}
	return out, nil
}

// consoleSignalsMarket assesses one market.
//
// A market that could not be read comes back present with Why set rather than
// omitted: a market missing from the page is indistinguishable from a market with
// nothing in it, which is the confusion this whole screen exists to remove, one
// level up from the veto.
func consoleSignalsMarket(ctx context.Context, store *candidate.Store, market string,
	at time.Time) console.SignalsMarket {

	out := console.SignalsMarket{
		Market: market,
		Panel:  console.SignalsPanel{Why: consoleSignalsPanelWhy},
	}
	verdicts, err := candidate.Assess(ctx, store, candidate.AssessOptions{
		Market: market,
		At:     at,
		// The same thresholds `tossctl candidate scan` applies, from the same
		// function. This used to be its own literal, and a literal on each side of
		// the seam is how the two surfaces come to disagree about the same store at
		// the same instant — each one internally consistent, neither one failing.
		// See vetothresholds.go.
		Thresholds: candidateVetoThresholds(),
	})
	if err != nil {
		out.Why = err.Error()
		return out
	}
	out.Verdicts = verdicts

	// The same reducer the scan output uses, rather than the same three
	// constructors called again here. This block used to assemble the tallies by
	// hand: the two agreed, and the whole cost sat in the future — a fourth shadow
	// band added to internal/candidate would have appeared on `tossctl candidate
	// scan` and not on /signals, and a band missing from a page renders as a band
	// nobody crossed rather than as one nobody counted.
	out.Vetoes, out.Crossings, out.Bands = candidate.TallyVerdicts(verdicts)
	// Same reducer as the scan output's, for the same reason as the line above it.
	out.Sightings = candidate.TallySightingSources(verdicts)
	return out
}
