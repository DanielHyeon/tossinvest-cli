package main

// a203_replay.go 는 a203 「돌파 KR 과거 재생」 하위 명령 replay — 시간대 거절 · 세션 목록 · 수집 · 재생 순서
// (openspec/changes/a203-breakout-kr-is-replayed-on-past-sessions design §2 · §3 · §6 · §8).
//
//	replay sessions --config-dir DIR [--out FILE] [--dry-run]            공식 달력으로 세션 목록(tasks 1.1)
//	replay collect  --sample control|main --sessions FILE --config-dir DIR [--dry-run] [--max-gets N]
//	replay report   --sample control|main --sessions FILE --out FILE      캐시만 읽음(전송 0)
//
// 전송은 a203_cache.go 의 a203Get 한 곳뿐이고 공식 GET 넷(1분봉 · 일봉 · 랭킹 · 달력)만 부름. 주문 · 서명 · 엔진 전달 ·
// 운영 설정 쓰기 0 — a140 static_test.go 세 시험이 이 파일에도 그대로 걸림(design §2).

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudgetruntime"
	apppaths "github.com/JungHoonGhae/tossinvest-cli/internal/app/paths"
	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutsignal"
	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialbars"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

const (
	a203Anchor          = "2026-11-26" // 기준일(이 날 전 세션만 — design §6)
	a203MainSessions    = 100
	a203Warmup          = 21 // 가장 이른 주 표본 세션 앞 21 세션(R 창 d−2 … d−21 · a204 d−1)
	a203PollAfterClose  = 30 * time.Minute
	a203MaxPages        = 4
	a203Gap             = time.Second
	a203DefaultMaxGets  = 13_600 // H3 승인 상한(a203 + a204 추가분, 2026-11-27 사용자)
	a203DefaultA140     = "openspec/changes/archive/2026-10-22-a140-breakout-signals-are-measured-in-shadow/analysis/shadow-ledger/ledger.jsonl"
	a203EvidenceName    = "a203-evidence.sqlite"
	a203SampleMain      = "main"
	a203SampleControl   = "control"
	a203RoleWarmup      = "warmup"
	a203UniverseUC      = "uc"
	a203UniverseA140    = "a140"
	a203UniverseA204    = "a204"
	a203SessionsVersion = "a203-sessions/v1"
)

var (
	a203ControlDates = []string{"2026-10-16", "2026-10-19", "2026-10-20", "2026-10-21", "2026-10-22"}
	a203Seoul        = a203Location(marketclock.MarketKR)
	a203NewYork      = a203Location(marketclock.MarketUS)
	// a203Exit 은 M9 의 종료 코드 2 를 냄 — 시험이 바꿔 끼움.
	a203Exit = os.Exit
)

func a203Location(m marketclock.Market) *time.Location {
	loc, err := m.Location()
	if err != nil {
		panic(err) // time/tzdata 가 main.go 에 박혀 있어 도달 불가
	}
	return loc
}

// a203Refused 는 결정 M3 의 시간대 거절 — 달력 없이 시각만으로 판정함(휴장일에도 거절하는 쪽으로 틀림).
//
//	평일(KST) 08:00 ≤ t < 20:00 — KR 정규장 + NXT 프리 · 애프터마켓(엔진이 KR 손절 가격을 조회하는 시간)
//	평일(ET)  09:30 ≤ t < 16:00 — 미국 정규장(서머타임은 America/New_York 이 반영). 금요일 장은 토요일 KST 새벽이라
//	                              「토요일 미국 정규장 마감 전 거절」이 여기서 저절로 나옴.
func a203Refused(at time.Time) (bool, string) {
	weekday := func(t time.Time) bool { return t.Weekday() != time.Saturday && t.Weekday() != time.Sunday }
	minutes := func(t time.Time) int { return t.Hour()*60 + t.Minute() }
	if kr := at.In(a203Seoul); weekday(kr) && minutes(kr) >= 8*60 && minutes(kr) < 20*60 {
		return true, "weekday 08:00-20:00 KST (KR regular + NXT)"
	}
	if us := at.In(a203NewYork); weekday(us) && minutes(us) >= 9*60+30 && minutes(us) < 16*60 {
		return true, "US regular session 09:30-16:00 ET"
	}
	return false, ""
}

// a203NextChange 는 at 뒤로 거절 여부가 처음 바뀌는 분(8 일 안, 없으면 0 값).
func a203NextChange(at time.Time) time.Time {
	now, _ := a203Refused(at)
	t := at.Truncate(time.Minute)
	for i := 0; i < 8*24*60; i++ {
		t = t.Add(time.Minute)
		if refused, _ := a203Refused(t); refused != now {
			return t
		}
	}
	return time.Time{}
}

// a203Options 는 한 실행의 인자.
type a203Options struct {
	command, stateDir, configDir, sessionsFile, sample, out, a140Ledger string
	dryRun                                                              bool
	maxGets                                                             int
	anchor                                                              string
	mainSessions, warmup                                                int
	now                                                                 func() time.Time
	sleep                                                               func(context.Context, time.Duration) error
	openLive                                                            func(context.Context, string) (context.Context, a203Live, func() error, error)
	stdout, stderr                                                      io.Writer
}

func replayCommand(ctx context.Context, args []string, stdout io.Writer) error {
	opts, err := parseA203Options(args, stdout)
	if err == nil {
		err = runA203(ctx, opts)
	}
	if errors.Is(err, errA203Population) {
		fmt.Fprintln(os.Stderr, "a140 breakout signal shadow: replay: "+err.Error())
		a203Exit(2)
	}
	return err
}

func parseA203Options(args []string, stdout io.Writer) (a203Options, error) {
	if len(args) == 0 {
		return a203Options{}, errors.New(`replay needs a step: "sessions", "collect" or "report"`)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return a203Options{}, err
	}
	o := a203Options{command: args[0], anchor: a203Anchor, mainSessions: a203MainSessions, warmup: a203Warmup,
		now: func() time.Time { return time.Now().UTC() }, sleep: a203Sleep, openLive: openA203Live, stdout: stdout, stderr: os.Stderr}
	flags := flag.NewFlagSet("replay "+args[0], flag.ContinueOnError)
	flags.StringVar(&o.stateDir, "state-dir", filepath.Join(home, ".local", "state", "tossos-a203"), "state directory outside the repo (cache · ledger · P(a))")
	flags.StringVar(&o.configDir, "config-dir", "", "required to send: the tossctl profile (no default — name it)")
	flags.StringVar(&o.sessionsFile, "sessions", "", "the committed session list (analysis/sample/sessions.txt); collect/report refuse a different list")
	flags.StringVar(&o.sample, "sample", "", `"control" (5 sessions, tasks 3) or "main" (100 sessions, tasks 4)`)
	flags.StringVar(&o.out, "out", "", "sessions: where to write the list (default <state-dir>/sessions.txt); report: the report file")
	flags.StringVar(&o.a140Ledger, "a140-ledger", a203DefaultA140, "the a140 ledger (lot params · control 1 universe)")
	flags.BoolVar(&o.dryRun, "dry-run", false, "send nothing: print the plan, the GET estimate and the next time-window change")
	flags.IntVar(&o.maxGets, "max-gets", a203DefaultMaxGets, "per-run GET cap (stop and checkpoint when reached)")
	if err := flags.Parse(args[1:]); err != nil {
		return a203Options{}, err
	}
	switch o.command {
	case "sessions":
	case "collect", "report":
		if o.sample != a203SampleMain && o.sample != a203SampleControl {
			return a203Options{}, fmt.Errorf(`--sample must be "control" or "main", got %q`, o.sample)
		}
		if o.sessionsFile == "" {
			return a203Options{}, errors.New("--sessions is required (the committed list from `replay sessions`)")
		}
		if o.command == "report" && o.out == "" {
			return a203Options{}, errors.New("--out is required for report")
		}
	default:
		return a203Options{}, fmt.Errorf(`unknown replay step %q; use "sessions", "collect" or "report"`, o.command)
	}
	if o.command != "report" && !o.dryRun && strings.TrimSpace(o.configDir) == "" {
		return a203Options{}, errors.New("--config-dir is required to send (or pass --dry-run)")
	}
	if o.maxGets < 1 {
		return a203Options{}, errors.New("--max-gets must be positive")
	}
	return o, nil
}

func a203Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// openA203Live 는 a140 run 과 같은 연결 — 계정 공유 예산(a166) 팩토리 클라이언트 + 보조 우선순위. 설정 파일은 읽지 않음.
func openA203Live(ctx context.Context, configDir string) (context.Context, a203Live, func() error, error) {
	credentialFile, tokenFile, err := apppaths.OpenAPI(configDir)
	if err != nil {
		return ctx, nil, nil, fmt.Errorf("resolving the Open API paths: %w", err)
	}
	credentials, err := official.LoadCredentials(os.Getenv, credentialFile)
	if err != nil {
		return ctx, nil, nil, fmt.Errorf("reading the Open API credentials: %w", err)
	}
	if credentials == nil {
		return ctx, nil, nil, errors.New("no Open API credentials in that profile")
	}
	mode, modeSet := os.LookupEnv("TOSSOS_ACCOUNT_BUDGET_MODE")
	manifest, manifestSet := os.LookupEnv("TOSSOS_ACCOUNT_BUDGET_MANIFEST")
	budgetConfig, err := accountbudgetruntime.ParseLaunchConfig(mode, modeSet, manifest, manifestSet)
	if err != nil {
		return ctx, nil, nil, err
	}
	budgetConfig.CleanupError = func(error) { fmt.Fprintln(os.Stderr, "budget cleanup failed") }
	factory, err := accountbudgetruntime.Open(ctx, budgetConfig)
	if err != nil {
		return ctx, nil, nil, err
	}
	client, err := factory.NewClient(*credentials, tokenFile)
	if err != nil {
		return ctx, nil, nil, errors.Join(err, factory.Close())
	}
	return accountbudget.WithPriority(ctx, accountbudget.Auxiliary), client, factory.Close, nil
}

func (o a203Options) cacheRoot() string { return filepath.Join(o.stateDir, "cache") }

func (o a203Options) offline() *a203Fetcher {
	return newA203Fetcher(nil, o.cacheRoot(), o.now, o.sleep, a203Gap, o.maxGets)
}

func runA203(ctx context.Context, o a203Options) error {
	// M9 — 어느 단계든 사전등록 모집단부터 확인(다르면 전송 0 · 종료 코드 2).
	pa, err := a203ReadPA(filepath.Join(o.stateDir, a203PAFile))
	if err != nil {
		return err
	}
	if o.command == "report" {
		return a203ReportCommand(ctx, o, pa)
	}
	if o.dryRun {
		return a203Plan(ctx, o, pa)
	}
	ctx, live, closeLive, err := o.openLive(ctx, o.configDir)
	if err != nil {
		return err
	}
	f := newA203Fetcher(live, o.cacheRoot(), o.now, o.sleep, a203Gap, o.maxGets)
	run := a203RunRecord{Command: o.command, Sample: o.sample, Start: o.now()}
	if o.command == "sessions" {
		err = a203SessionsCommand(ctx, o, f)
	} else {
		err = a203Collect(ctx, o, f, pa)
	}
	err = errors.Join(err, closeLive())
	run.End, run.Gets, run.ByKind, run.CacheHits = o.now(), f.gets, f.byKind, f.hits
	if err != nil {
		run.Stop = err.Error()
	}
	if logErr := a203AppendRun(o.stateDir, run); logErr != nil {
		err = errors.Join(err, logErr)
	}
	var stop *a203StopError
	if errors.As(err, &stop) {
		// 체크포인트 = 멈춘 자리 · 사유. 이어하기는 같은 명령을 다시 돌리면 됨(캐시에 있는 것은 GET 0).
		if cpErr := a203WriteFile(filepath.Join(o.stateDir, "checkpoint.json"), run); cpErr != nil {
			err = errors.Join(err, cpErr)
		}
	}
	fmt.Fprintf(o.stderr, "a203: %s %s — GET %d %v · cache hits %d\n", o.command, o.sample, f.gets, f.byKind, f.hits)
	return err
}

// a203RunRecord 는 실행 한 번(runs.jsonl 한 줄 · 체크포인트).
type a203RunRecord struct {
	Command   string         `json:"command"`
	Sample    string         `json:"sample,omitempty"`
	Start     time.Time      `json:"start"`
	End       time.Time      `json:"end"`
	Gets      int            `json:"gets"`
	ByKind    map[string]int `json:"gets_by_kind"`
	CacheHits int            `json:"cache_hits"`
	Stop      string         `json:"stop,omitempty"`
}

func a203AppendRun(stateDir string, run a203RunRecord) error {
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return err
	}
	raw, err := json.Marshal(run)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(stateDir, "runs.jsonl"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(append(raw, '\n'))
	return errors.Join(err, f.Sync(), f.Close())
}

func a203ReadRuns(stateDir string) ([]a203RunRecord, error) {
	f, err := os.Open(filepath.Join(stateDir, "runs.jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var runs []a203RunRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var run a203RunRecord
		if err := json.Unmarshal(scanner.Bytes(), &run); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, scanner.Err()
}

// ---- 세션 목록 (design §6 · tasks 1.1) ----

// a203Session 은 공식 달력의 KR 정규장 세션 하나.
type a203Session struct {
	Date, Role  string
	Open, Close time.Time
	payload     official.MarketCalendarResponse
}

func (s a203Session) id() string { return "KRX:" + s.Date }

// snapshot 은 그 세션을 today 로 둔 달력과 폴 시각(마감 + 30 분 — a140 라벨 유예와 같음). 달력을 받은 시각은 과거 폴 시각보다
// 늦으므로(officialbars 가 시계 역행으로 거절함) 재생의 「받은 시각」을 폴 시각으로 둠 — 달력 판본(digest)은 받은 시각과 무관.
func (s a203Session) snapshot() (scheduler.CalendarSnapshot, time.Time, error) {
	pollAt := s.Close.Add(a203PollAfterClose)
	cal, err := scheduler.AdaptOfficialCalendar(marketclock.MarketKR, s.payload, pollAt)
	return cal, pollAt, err
}

func (s a203Session) info(cal scheduler.CalendarSnapshot) sessionInfo {
	return sessionInfo{ID: s.id(), CalendarVersion: cal.Version, Open: s.Open, Close: s.Close}
}

// a203DiscoverSessions 는 기준일 전날부터 공식 달력의 전 영업일 사슬을 거슬러 감 — 세션마다 그 날짜의 달력 1 회.
// 역할: 대조 5 → control, 나머지 가장 최근 100 → main, 그 앞 21 → warmup. 대조가 주 표본 창 안에 없으면 거절.
func a203DiscoverSessions(ctx context.Context, f *a203Fetcher, o a203Options) ([]a203Session, error) {
	controls := map[string]bool{}
	for _, d := range a203ControlDates {
		controls[d] = true
	}
	anchor, err := time.Parse("2006-01-02", o.anchor)
	if err != nil {
		return nil, err
	}
	cursor := anchor.AddDate(0, 0, -1).Format("2006-01-02")
	var found []a203Session
	mains, ctrls, warms := 0, 0, 0
	for guard := 0; mains < o.mainSessions || ctrls < len(controls) || warms < o.warmup; guard++ {
		if guard > 4*(o.mainSessions+len(controls)+o.warmup)+30 {
			return nil, errors.New("the calendar chain did not reach the required sessions")
		}
		payload, err := f.calendar(ctx, cursor)
		if err != nil {
			return nil, err
		}
		cal, err := scheduler.AdaptOfficialCalendar(marketclock.MarketKR, payload, anchor)
		if err != nil {
			return nil, fmt.Errorf("the calendar of %s: %w", cursor, err)
		}
		if cal.Today.Date == cursor && cal.Today.Regular != nil {
			s := a203Session{Date: cursor, Open: cal.Today.Regular.Open, Close: cal.Today.Regular.Close, payload: payload}
			switch {
			case controls[cursor]:
				s.Role, ctrls = a203SampleControl, ctrls+1
			case mains < o.mainSessions:
				s.Role, mains = a203SampleMain, mains+1
			case ctrls < len(controls):
				return nil, fmt.Errorf("the %d main sessions end at %s before every control session was seen", o.mainSessions, cursor)
			default:
				s.Role, warms = a203RoleWarmup, warms+1
			}
			found = append(found, s)
		}
		previous := cal.PreviousBusinessDay.Date
		if previous == "" || previous >= cursor {
			return nil, fmt.Errorf("the calendar of %s gives the previous business day %q", cursor, previous)
		}
		cursor = previous
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Date < found[j].Date })
	return found, nil
}

// a203RenderSessions 는 커밋할 세션 목록(종목 · 가격 없음).
func a203RenderSessions(o a203Options, sessions []a203Session) []byte {
	counts := map[string]int{}
	for _, s := range sessions {
		counts[s.Role]++
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — anchor %s (sessions strictly before it), official KR calendar (TypedMarketCalendar), ascending\n", a203SessionsVersion, o.anchor)
	fmt.Fprintf(&b, "# main %d · control %d · warmup %d (R window d-2..d-21 and the a204 d-1 of the earliest main session)\n",
		counts[a203SampleMain], counts[a203SampleControl], counts[a203RoleWarmup])
	b.WriteString("# date role regular-open regular-close (KST)\n")
	for _, s := range sessions {
		fmt.Fprintf(&b, "%s %s %s %s\n", s.Date, s.Role, s.Open.In(a203Seoul).Format("15:04"), s.Close.In(a203Seoul).Format("15:04"))
	}
	return []byte(b.String())
}

func a203SessionsCommand(ctx context.Context, o a203Options, f *a203Fetcher) error {
	sessions, err := a203DiscoverSessions(ctx, f, o)
	if err != nil {
		return err
	}
	raw := a203RenderSessions(o, sessions)
	out := o.out
	if out == "" {
		out = filepath.Join(o.stateDir, "sessions.txt")
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, raw, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(o.stdout, "a203 sessions: %d written to %s · sha256 %s\n", len(sessions), out, a203SHA(raw))
	return nil
}

// a203LoadSessions 는 커밋된 목록을 캐시 달력에서 다시 만들어 바이트로 대조함(사전등록 목록과 다른 세션으로 수집 · 재생 금지).
func a203LoadSessions(ctx context.Context, o a203Options) ([]a203Session, string, error) {
	committed, err := os.ReadFile(o.sessionsFile)
	if err != nil {
		return nil, "", err
	}
	sessions, err := a203DiscoverSessions(ctx, o.offline(), o)
	if err != nil {
		return nil, "", fmt.Errorf("rebuilding the session list from the cached calendar (run `replay sessions` first): %w", err)
	}
	if rebuilt := a203RenderSessions(o, sessions); string(rebuilt) != string(committed) {
		return nil, "", fmt.Errorf("%s (sha256 %s) is not the list the cached calendar gives (sha256 %s)", o.sessionsFile, a203SHA(committed), a203SHA(rebuilt))
	}
	return sessions, a203SHA(committed), nil
}

// ---- 수집 대상 (design §5 · §6 · §8) ----

// a203Job 은 종목-세션 하나. Universe = uc(U-C) · a140(대조 1) · a204(d−1 1분봉 추가분 — 재생하지 않고 수집만).
type a203Job struct {
	Index    int // sessions 안 번호
	Symbol   string
	Share    string
	Universe string
}

// a203Inputs 는 수집 · 재생이 함께 쓰는 입력.
type a203Inputs struct {
	sessions    []a203Session
	sessionsSHA string
	params      shadowParams
	a140U       map[string][]string
	a140Rows    []ledgerRow
	pa, pb      []string
	population  []string
	daily       map[string]map[string]a203Day
	dailyMiss   int
	jobs        []a203Job
	universes   map[string][]a203Pick // 세션 날짜 → U-C 선정
	eligible    map[string]int
}

// a203A140Lot 은 a140 원장에서 판 설정(design §1 — 현재 설정 파일을 다시 읽지 않음)과 세션별 종목군.
func a203A140Lot(path string) (shadowParams, map[string][]string, []ledgerRow, error) {
	rows, err := readLedger(path)
	if err != nil {
		return shadowParams{}, nil, nil, err
	}
	universes := map[string][]string{}
	var params *shadowParams
	for _, row := range rows {
		if row.Kind != kindUniverse {
			continue
		}
		if row.Params == nil || params != nil && *row.Params != *params {
			return shadowParams{}, nil, nil, fmt.Errorf("%s: the a140 universe rows do not carry one lot of params", path)
		}
		params = row.Params
		universes[strings.TrimPrefix(row.Session, "KRX:")] = row.Symbols
	}
	if params == nil {
		return shadowParams{}, nil, nil, fmt.Errorf("%s: no a140 universe row", path)
	}
	return *params, universes, rows, nil
}

// a203Prepare 는 목록 · 판 설정 · 모집단 · 일봉 · 종목군 · 수집 대상을 만듦. needAll 이면 일봉이 하나라도 없을 때 거절.
func a203Prepare(ctx context.Context, o a203Options, f *a203Fetcher, pa []string) (a203Inputs, error) {
	var in a203Inputs
	var err error
	if in.sessions, in.sessionsSHA, err = a203LoadSessions(ctx, o); err != nil {
		return in, err
	}
	if in.params, in.a140U, in.a140Rows, err = a203A140Lot(o.a140Ledger); err != nil {
		return in, err
	}
	in.pa = pa
	ranking, err := f.ranking(ctx)
	if err != nil {
		return in, fmt.Errorf("P(b): %w", err)
	}
	for _, item := range ranking.Items {
		in.pb = append(in.pb, item.Symbol)
	}
	in.population = a203Union(pa, in.pb)
	in.daily = map[string]map[string]a203Day{}
	for _, code := range in.population {
		chart, err := f.daily(ctx, code)
		if errors.Is(err, errA203NotCached) {
			in.dailyMiss++
			continue
		}
		if err != nil {
			return in, err
		}
		in.daily[code] = a203DailyIndex(chart)
	}
	if in.dailyMiss > 0 {
		return in, nil // 호출자가 판단(dry-run 은 추정, report 는 거절)
	}
	in.universes, in.eligible = map[string][]a203Pick{}, map[string]int{}
	seen := map[[2]string]bool{}
	add := func(job a203Job) {
		key := [2]string{in.sessions[job.Index].Date, job.Symbol + "\x00" + job.Universe}
		if !seen[key] {
			seen[key] = true
			in.jobs = append(in.jobs, job)
		}
	}
	for i, s := range in.sessions {
		if s.Role != o.sample {
			continue
		}
		if i < 1+a203RWindow {
			return in, fmt.Errorf("session %s has fewer than %d sessions before it in the list", s.Date, 1+a203RWindow)
		}
		window := make([]string, 0, 1+a203RWindow)
		for k := 1; k <= 1+a203RWindow; k++ {
			window = append(window, in.sessions[i-k].Date)
		}
		picks, eligible := a203SelectUC(window, in.population, in.daily, in.params.NotionalCapKRW)
		in.universes[s.Date], in.eligible[s.Date] = picks, eligible
		if o.sample == a203SampleControl {
			for _, code := range in.a140U[s.Date] {
				add(a203Job{Index: i, Symbol: code, Share: a203ShareA140, Universe: a203UniverseA140})
			}
		}
		for _, p := range picks {
			add(a203Job{Index: i, Symbol: p.Symbol, Share: p.Share, Universe: a203UniverseUC})
		}
		// a204 추가분(design §8 · M10): U-C 종목군 전원의 d−1 1분봉. 대조 1(a140 종목군)은 넣지 않음.
		for _, p := range picks {
			add(a203Job{Index: i - 1, Symbol: p.Symbol, Universe: a203UniverseA204})
		}
	}
	if o.sample == a203SampleControl {
		for _, d := range a203ControlDates {
			if len(in.a140U[d]) == 0 {
				return in, fmt.Errorf("the a140 ledger has no universe for the control session %s", d)
			}
		}
	}
	return in, nil
}

// a203PollInput 은 종목-세션 하나의 officialbars 폴 인자(수집 · 재생 · dry-run 이 같은 요청 열을 만들게 한 곳에서).
func a203PollInput(s a203Session, symbol string) (officialbars.BarPollInput, error) {
	cal, pollAt, err := s.snapshot()
	if err != nil {
		return officialbars.BarPollInput{}, err
	}
	return officialbars.BarPollInput{Market: marketclock.MarketKR, Symbol: symbol, Calendar: cal, PollAt: pollAt,
		LowerBound: s.Open, MaxPages: a203MaxPages}, nil
}

// a203UniverseFile 은 a204 어댑터가 읽는 종목군 파일(analysis/cache-format.md).
type a203UniverseFile struct {
	Format      string                `json:"format"`
	Sample      string                `json:"sample"`
	SessionsSHA string                `json:"sessions_sha256"`
	Universe    map[string][]a203Pick `json:"universe"`
	Prior       map[string]string     `json:"prior_session"`
}

func a203Collect(ctx context.Context, o a203Options, f *a203Fetcher, pa []string) error {
	in, err := a203Prepare(ctx, o, f, pa)
	if err != nil {
		return err
	}
	if in.dailyMiss > 0 {
		return fmt.Errorf("%d daily series are missing after collection", in.dailyMiss) // 라이브 수집에서는 도달 불가 방어
	}
	uf := a203UniverseFile{Format: "a203-universe/v1", Sample: o.sample, SessionsSHA: in.sessionsSHA, Universe: in.universes, Prior: map[string]string{}}
	for i, s := range in.sessions {
		if s.Role == o.sample {
			uf.Prior[s.Date] = in.sessions[i-1].Date
		}
	}
	if err := a203WriteFile(filepath.Join(o.stateDir, "universe", "uc-"+o.sample+".json"), uf); err != nil {
		return err
	}
	refused := 0
	for n, job := range in.jobs {
		input, err := a203PollInput(in.sessions[job.Index], job.Symbol)
		if err != nil {
			return err
		}
		result, err := officialbars.PollClosedBars(ctx, a203MinuteReader{f}, a203NullStore{}, input)
		if f.stop != nil {
			return f.stop
		}
		var refusal *officialbars.PollRefusal
		if errors.As(err, &refusal) && refusal.Reason == officialbars.RefusalReaderError {
			// 리더 쪽 오류(캐시 파일을 못 읽음 · 요청 문법 거절)는 받은 페이지의 결함이 아님 — 결손으로 세고 넘어가면 이어하기마다
			// 같은 자리가 영구 결손이 되므로 오류처럼 실행을 멈춤(design §8 「오류는 멈춤」).
			return err
		}
		if err != nil {
			refused++ // 받은 페이지가 계약에 어긋남 — 페이지는 이미 캐시됨, 재생에서 결손으로 셈
			fmt.Fprintf(o.stderr, "a203: %s %s refused by the bar producer: %v\n", in.sessions[job.Index].Date, job.Universe, err)
			continue
		}
		if result.RateGated {
			f.stop = &a203StopError{Reason: "the broker reported its rate allowance spent", Key: in.sessions[job.Index].Date}
			return f.stop
		}
		if (n+1)%100 == 0 {
			fmt.Fprintf(o.stderr, "a203: %d/%d symbol-sessions · GET %d\n", n+1, len(in.jobs), f.gets)
		}
	}
	fmt.Fprintf(o.stdout, "a203 collect %s: %d symbol-sessions (refused by the producer %d) · GET %d · cache hits %d\n",
		o.sample, len(in.jobs), refused, f.gets, f.hits)
	return nil
}

// a203Plan 은 --dry-run — 전송 0. 세션 목록 · 종목군 크기 · GET 추정 · 다음 시간대 경계를 찍음.
func a203Plan(ctx context.Context, o a203Options, pa []string) error {
	w := o.stdout
	off := o.offline()
	now := o.now()
	low, high := 0, 0
	fmt.Fprintf(w, "a203 dry-run %s %s — sends 0 · P(a) %d (sha256 ok)\n", o.command, o.sample, len(pa))
	if o.command == "sessions" {
		_, err := a203DiscoverSessions(ctx, off, o)
		need := o.mainSessions + len(a203ControlDates) + o.warmup
		switch {
		case err == nil:
			fmt.Fprintf(w, "- calendar: all %d sessions cached · GET 0\n", need)
		case errors.Is(err, errA203NotCached):
			low, high = max(1, need-off.hits), max(1, need-off.hits)+5
			fmt.Fprintf(w, "- calendar: %d cached · about %d more GET (one per session, holidays add one each)\n", off.hits, low)
		default:
			return err
		}
	} else {
		in, err := a203Prepare(ctx, o, off, pa)
		switch {
		case errors.Is(err, errA203NotCached): // 랭킹 없음
			low, high = 1+len(pa), 1+len(pa)+100
			fmt.Fprintf(w, "- P(b) ranking: not cached · GET 1, then daily for P = %d..%d symbols\n", len(pa), len(pa)+100)
		case err != nil:
			return err
		}
		if err == nil {
			fmt.Fprintf(w, "- sessions %s (sha256 %s): %d listed\n", o.sessionsFile, in.sessionsSHA, len(in.sessions))
			fmt.Fprintf(w, "- P(b) %d · P %d · daily cached %d · daily missing %d (GET %d)\n", len(in.pb), len(in.population),
				len(in.daily), in.dailyMiss, in.dailyMiss)
			low, high = in.dailyMiss, in.dailyMiss
		}
		perSession := []int{0}
		if err == nil && in.dailyMiss == 0 {
			byUniverse := map[string][2]int{}
			counted := map[[2]string]bool{} // 같은 종목-세션은 요청 열이 같음 — 한 번만 셈
			for _, job := range in.jobs {
				key := [2]string{in.sessions[job.Index].Date, job.Symbol}
				if counted[key] {
					continue
				}
				counted[key] = true
				input, err := a203PollInput(in.sessions[job.Index], job.Symbol)
				if err != nil {
					return err
				}
				result, pollErr := officialbars.PollClosedBars(ctx, a203MinuteReader{off}, a203NullStore{}, input)
				if pollErr == nil || !errors.Is(pollErr, errA203NotCached) {
					continue // 이미 다 받음(또는 받은 페이지가 거절됨 — 다시 받지 않음)
				}
				cached := result.Pages - 1
				span := byUniverse[job.Universe]
				span[0] += max(1, 2-cached)
				span[1] += max(1, 3-cached)
				byUniverse[job.Universe] = span
			}
			for _, u := range []string{a203UniverseA140, a203UniverseUC, a203UniverseA204} {
				span := byUniverse[u]
				low, high = low+span[0], high+span[1]
				fmt.Fprintf(w, "- minute pages %s: GET %d..%d\n", u, span[0], span[1])
			}
			for _, s := range in.sessions {
				if s.Role == o.sample {
					perSession = append(perSession, len(in.universes[s.Date]))
				}
			}
			fmt.Fprintf(w, "- universe per session (U-C): %v (sessions %d)\n", perSession[1:], len(perSession)-1)
		} else {
			sessions := o.mainSessions
			if o.sample == a203SampleControl {
				sessions = len(a203ControlDates)
			}
			// 종목군이 아직 없으면 design §8 식으로 추정(τ = 회전율 — 0 · 0.3 · 1).
			minute := [2]int{sessions * 30 * 2, sessions * 30 * 3}
			if o.sample == a203SampleControl {
				minute[0], minute[1] = minute[0]+sessions*30*2, minute[1]+sessions*30*3
			}
			low, high = low+minute[0]+30*2, high+minute[1]+(30+(sessions-1)*30)*3
			fmt.Fprintf(w, "- universe: unknown until the daily series are cached · minute GET %d..%d + a204 d-1 (30 + %d×30×τ)×2..3\n",
				minute[0], minute[1], sessions-1)
		}
	}
	refusedNow, why := a203Refused(now)
	change := a203NextChange(now)
	fmt.Fprintf(w, "- GET estimate %d..%d (≥ %s at 1 s apart) · per-run cap %d\n", low, high, time.Duration(high)*a203Gap, o.maxGets)
	if refusedNow {
		fmt.Fprintf(w, "- now %s is REFUSED (%s) · the window opens at %s\n", now.In(a203Seoul).Format(time.RFC3339), why, change.In(a203Seoul).Format(time.RFC3339))
	} else {
		fmt.Fprintf(w, "- now %s is allowed · first refusal at %s", now.In(a203Seoul).Format(time.RFC3339), change.In(a203Seoul).Format(time.RFC3339))
		if finish := now.Add(time.Duration(high) * a203Gap); !change.IsZero() && finish.After(change) {
			fmt.Fprintf(w, " — the upper estimate ends after it (%s): the run stops there and resumes on the next run", finish.In(a203Seoul).Format(time.RFC3339))
		}
		fmt.Fprintln(w)
	}
	return nil
}

// ---- 재생 (design §3) ----

// a203Outcome 은 종목-세션 하나의 재생 결과.
type a203Outcome struct {
	Session, Symbol, Share, Universe string
	Gap                              string // 결손 사유(봉 수집 실패 · 앞부분 부족 · GAP 로 끝남) — 비면 결손 아님
	Furthest                         string // 가장 멀리 간 단계(퍼널)
	Terminal                         string // ARMED 전 종결: ARMED · INVALIDATED · TIMED_OUT · 마지막 단계
	Refusals                         map[string]int
	EvalErrors                       int
	Last                             *evaluation
	LastQuote                        quoteSnap
	Trades                           map[string]*a203Trade
	Attempts                         map[string][]a203Attempt
	SeriesDigest                     string
	PageDigests                      []string
}

// a203ReplaySymbol 은 design §3 의 2 ~ 6: 캐시 리더 → PollClosedBars → SealBarSeries → contiguousPrefix → 단계 진행 → ARMED 시도 · 라벨.
func a203ReplaySymbol(ctx context.Context, reader officialbars.CandleReader, store *strategyevidence.Store, s a203Session, job a203Job,
	params shadowParams, now func() time.Time) a203Outcome {
	out := a203Outcome{Session: s.id(), Symbol: job.Symbol, Share: job.Share, Universe: job.Universe, Refusals: map[string]int{},
		Trades: map[string]*a203Trade{}, Attempts: map[string][]a203Attempt{}}
	input, err := a203PollInput(s, job.Symbol)
	if err != nil {
		out.Gap = "calendar: " + err.Error()
		return out
	}
	if _, err := officialbars.PollClosedBars(ctx, reader, store, input); err != nil {
		out.Gap = "bars: " + err.Error()
		return out
	}
	// 평가 시각은 과거(폴 시각), 적재 컷오프는 지금 — 봉은 오늘 적재됐으므로 과거 컷오프로는 아무것도 안 보임(재생 컷오프는 읽는 쪽 몫).
	series, err := store.SealBarSeries(ctx, strategyevidence.BarSeriesQuery{Market: "KR", Symbol: job.Symbol, SessionID: s.id(),
		IntervalMS: strategyevidence.ClosedBar1mIntervalMS, EvaluationAt: input.PollAt, IngestionCutoff: now(), RegularSessionOnly: true})
	if err != nil {
		out.Gap = "series: " + err.Error()
		return out
	}
	out.SeriesDigest = series.Digest
	pages := map[string]bool{}
	for _, record := range series.Bars {
		if d := record.Payload.SourceResponseDigest; !pages[d] {
			pages[d] = true
			out.PageDigests = append(out.PageDigests, d)
		}
	}
	sort.Strings(out.PageDigests)
	sess := s.info(input.Calendar)
	prefix, _, gapped := contiguousPrefix(series, s.Open)
	if len(prefix.Bars) <= breakoutsignal.OpeningRangeBars {
		out.Gap = fmt.Sprintf("bars: %d contiguous bars from the open (gap %v)", len(prefix.Bars), gapped)
		return out
	}
	ask0 := prefix.Bars[0].Payload.OpenMinor
	closeMS := uint64(s.Close.UnixMilli())
	for n := breakoutsignal.OpeningRangeBars + 1; n <= len(prefix.Bars); n++ {
		head := prefix
		head.Bars = prefix.Bars[:n]
		// 평가 시각 = 봉 n−1 이 열린 뒤 2 분 5 초 — a140 run 과 같음(design §3-4): 분마다 5 초에 step 하고(shadow.go:554),
		// 생산자는 가장 새 봉을 적재하지 않으므로(producer.go:357) 봉 k 는 k+2 분 step 에서 처음 평가됨.
		atMS := prefix.Bars[n-1].Payload.OpenAtMS + 2*minuteMS + 5_000
		if atMS >= closeMS {
			break // a140 은 마감 뒤 새 setup 을 평가하지 않음(shadow.go:396) — 마감 3 분 전 이후에 연 봉은 단계 판정에 안 씀
		}
		// ARMED 전 단계는 호가 값과 무관(C2) — 초기 대리 호가는 첫 봉 시가.
		quote := a203ProxyQuote(ask0, atMS, "init")
		ev, err := evaluateSetup(head, sess, job.Symbol, params, quote, atMS)
		if err != nil {
			out.EvalErrors++
			continue
		}
		if phaseRank[ev.Phase] > phaseRank[out.Furthest] {
			out.Furthest = ev.Phase
		}
		if out.Last == nil || ev.Phase != out.Last.Phase || ev.Refusal != out.Last.Refusal {
			if ev.Refusal != "" {
				out.Refusals[ev.Refusal]++ // a140 처럼 단계 · 거절이 바뀐 순간만 셈
			}
		}
		last := ev
		out.Last, out.LastQuote = &last, quote
		switch ev.Phase {
		case "ARMED", "PROPOSED":
			out.Terminal = "ARMED"
			for _, assumption := range a203Assumptions {
				trade, attempts := a203ArmedAttempts(prefix, series, ev.ArmedIndex, sess, job.Symbol, params, assumption)
				if trade != nil {
					trade.Share = job.Share
					out.Trades[assumption] = trade
				}
				out.Attempts[assumption] = attempts
			}
			if out.Trades[a203AssumptionP0] != nil {
				out.Furthest = "PROPOSED"
			}
			return out
		case "INVALIDATED", "TIMED_OUT":
			out.Terminal = ev.Phase
			return out
		}
	}
	if out.Last != nil {
		out.Terminal = out.Last.Phase
	}
	if gapped {
		out.Gap = "GAP: " + string(breakoutsignal.RefusalMinuteGap) + " before the setup ended"
	}
	return out
}

// a203Policy 는 판 정책 — a140 정책 + 호가 가정 + 종목군 규칙(P(a) · P(b) 지문) + 표본 목록(design §3-7).
type a203Policy struct {
	Version     string          `json:"version"`
	A140        effectivePolicy `json:"a140_policy"`
	Assumptions []string        `json:"quote_assumptions"`
	Primary     string          `json:"primary_assumption"`
	Universe    struct {
		Rule      string `json:"rule"`
		PA        string `json:"pa_sha256"`
		PACount   int    `json:"pa_count"`
		PB        string `json:"pb_sha256"`
		PBCount   int    `json:"pb_count"`
		PBRequest any    `json:"pb_request"`
		PerShare  int    `json:"per_share"`
		MinClose  int    `json:"min_close_krw"`
		RSessions int    `json:"r_sessions"`
	} `json:"universe"`
	Sessions       string `json:"sessions_sha256"`
	PollAfterClose int64  `json:"poll_after_close_ns"`
	MaxPages       int    `json:"max_pages"`
	Attempts       int    `json:"armed_attempts"`
	Verdict        string `json:"verdict_rule"`
}

func a203PolicyFor(in a203Inputs) (a203Policy, string) {
	cfg := defaultProbeConfig()
	cfg.Params = in.params
	p := a203Policy{Version: "a203-replay/v1", A140: effectivePolicyFor(cfg, sessionSourceCalendar), Assumptions: a203Assumptions,
		Primary: a203AssumptionP0, Sessions: in.sessionsSHA, PollAfterClose: int64(a203PollAfterClose), MaxPages: a203MaxPages,
		Attempts: a203Attempts, Verdict: "design §7: n>=30 · sessions>=20 · gaps<=20% · all labeled; mean>0 · bootstrap(203,2000) q.025>0 · ex-best>0"}
	p.Universe.Rule, p.Universe.PACount, p.Universe.PBCount = "U-C", len(in.pa), len(in.pb)
	p.Universe.PA, _ = a203ListFingerprint(in.pa)
	p.Universe.PB, _ = a203ListFingerprint(in.pb)
	p.Universe.PBRequest, p.Universe.PerShare, p.Universe.MinClose, p.Universe.RSessions = a203PBRanking, a203PerShare, a203MinCloseKRW, a203RWindow
	raw, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return p, "a203-replay/v1@sha256:" + a203SHA(raw)
}

// a203Row 는 원장 한 줄 — a140 행과 같은 칸 + replay 칸(design §3-7).
type a203Row struct {
	ledgerRow
	A203Policy *a203Policy    `json:"a203_policy,omitempty"`
	Replay     *a203RowReplay `json:"replay,omitempty"`
}

type a203RowReplay struct {
	Universe     string   `json:"universe,omitempty"`
	Share        string   `json:"share,omitempty"`
	Assumption   string   `json:"assumption,omitempty"`
	Terminal     string   `json:"terminal,omitempty"`
	SeriesDigest string   `json:"series_digest,omitempty"`
	PageDigests  []string `json:"page_digests,omitempty"`
	NetReturn    float64  `json:"net_return,omitempty"`
}

func a203OutcomeRows(o a203Outcome, lot string) []a203Row {
	base := func(row ledgerRow, r a203RowReplay) a203Row {
		row.Schema, row.Session, row.Lot = "a203-replay/v1", o.Session, lot
		r.Universe, r.Share = o.Universe, o.Share
		return a203Row{ledgerRow: row, Replay: &r}
	}
	var rows []a203Row
	if o.Gap != "" {
		rows = append(rows, base(ledgerRow{Kind: kindGap, Symbol: o.Symbol, Stage: "replay", Reason: o.Gap}, a203RowReplay{Terminal: o.Terminal}))
	}
	if o.Last != nil {
		rows = append(rows, base(signalRow(o.Symbol, *o.Last, o.LastQuote), a203RowReplay{Terminal: o.Terminal, SeriesDigest: o.SeriesDigest, PageDigests: o.PageDigests}))
	}
	for _, assumption := range a203Assumptions {
		for _, a := range o.Attempts[assumption] {
			row := signalRow(o.Symbol, a.Eval, a.Quote)
			row.Attempt, row.Reason = a.N, a.Err
			rows = append(rows, base(row, a203RowReplay{Assumption: assumption}))
		}
		if t := o.Trades[assumption]; t != nil && t.Labeled {
			rows = append(rows, base(ledgerRow{Kind: kindLabel, Symbol: o.Symbol, SetupID: t.SetupID, Outcome: t.Label.Outcome, ExitMinor: t.Label.ExitMinor,
				ExitOpenAt: time.UnixMilli(int64(t.Label.ExitOpenMS)).UTC().Format(time.RFC3339), RPPM: t.Label.RPPM, LabelFromMS: t.LabelFromMS},
				a203RowReplay{Assumption: assumption, NetReturn: t.Return}))
		}
	}
	return rows
}

// a203ReportCommand 는 캐시만 읽어 재생 · 원장 · 보고서를 만듦(전송 0 — fetcher 에 live 가 없음).
func a203ReportCommand(ctx context.Context, o a203Options, pa []string) error {
	in, err := a203Prepare(ctx, o, o.offline(), pa)
	if err != nil {
		return err
	}
	if in.dailyMiss > 0 {
		return fmt.Errorf("%d daily series of P are not cached — finish `replay collect --sample %s` first", in.dailyMiss, o.sample)
	}
	evidencePath := filepath.Join(o.stateDir, a203EvidenceName)
	if err := checkEvidencePath(evidencePath, productionEvidenceCandidates("")); err != nil {
		return err
	}
	if err := os.MkdirAll(o.stateDir, 0o700); err != nil {
		return err
	}
	store, err := strategyevidence.Open(ctx, strategyevidence.Options{Path: evidencePath})
	if err != nil {
		return fmt.Errorf("opening the replay evidence store: %w", err)
	}
	defer store.Close()
	policy, lot := a203PolicyFor(in)
	reader := a203MinuteReader{o.offline()}
	var outcomes []a203Outcome
	for _, job := range in.jobs {
		if job.Universe == a203UniverseA204 {
			continue // a204 추가분은 수집만(a204 가 읽음)
		}
		// 적재 컷오프는 저장소 시계(시스템 시계)와 같은 시계여야 방금 적재한 봉이 보임.
		outcomes = append(outcomes, a203ReplaySymbol(ctx, reader, store, in.sessions[job.Index], job, in.params, func() time.Time { return time.Now().UTC() }))
	}
	ledgerPath := filepath.Join(o.stateDir, "ledger", o.sample+"-"+o.now().UTC().Format("20060102T150405Z")+".jsonl")
	var lines []string
	header := a203Row{ledgerRow: ledgerRow{Schema: "a203-replay/v1", Kind: "lot", At: o.now().UTC().Format(time.RFC3339), Lot: lot}, A203Policy: &policy}
	for _, row := range append([]a203Row{header}, a203FlatRows(outcomes, lot)...) {
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		lines = append(lines, string(raw))
	}
	ledgerBytes := []byte(strings.Join(lines, "\n") + "\n")
	if err := os.MkdirAll(filepath.Dir(ledgerPath), 0o700); err != nil {
		return err
	}
	lf, err := os.OpenFile(ledgerPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := lf.Write(ledgerBytes); err != nil {
		lf.Close()
		return err
	}
	if err := errors.Join(lf.Sync(), lf.Close()); err != nil {
		return err
	}
	cacheSHA, cacheFiles, err := a203CacheFingerprint(o.cacheRoot())
	if err != nil {
		return err
	}
	runs, err := a203ReadRuns(o.stateDir)
	if err != nil {
		return err
	}
	costPPM, costErr := costRatePPM(costs.DefaultModel())
	text := a203RenderReport(a203ReportInput{sample: o.sample, in: in, outcomes: outcomes, lot: lot, cacheSHA: cacheSHA, cacheFiles: cacheFiles,
		ledgerPath: filepath.ToSlash(strings.TrimPrefix(ledgerPath, o.stateDir+string(filepath.Separator))), ledgerSHA: a203SHA(ledgerBytes),
		ledgerLines: len(lines), ledgerLen: len(ledgerBytes), runs: runs,
		costPPMNow: costPPM, costErr: costErr})
	if err := os.WriteFile(o.out, []byte(text), 0o644); err != nil {
		return err
	}
	_, err = io.WriteString(o.stdout, text)
	if mismatches := a203ControlMismatches(in, outcomes); len(mismatches) > 0 {
		// 종목 코드가 든 불일치 목록은 저장소 밖에만 둠(보고서에는 수만).
		path := filepath.Join(o.stateDir, "control1-mismatches.txt")
		err = errors.Join(err, a203WriteBytes(path, []byte(strings.Join(mismatches, "\n")+"\n")))
		fmt.Fprintf(o.stderr, "a203: control 1 mismatches listed in %s\n", path)
	}
	return err
}

func a203FlatRows(outcomes []a203Outcome, lot string) []a203Row {
	var rows []a203Row
	for _, o := range outcomes {
		rows = append(rows, a203OutcomeRows(o, lot)...)
	}
	return rows
}
