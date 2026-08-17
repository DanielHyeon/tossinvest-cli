## ADDED Requirements

### Requirement: breakout-retest lane은 closed-bar pure evaluator다
KR/US breakout-retest lane은 immutable official evidence와 versioned config를 입력받아 state transition, proposal 또는 typed refusal을 결정적으로 반환해야 하며 (SHALL), broker, writable journal, Guardian, scheduler/activation writer와 system clock을 직접 호출해서는 안 된다 (MUST NOT). Breakout, retest와 reclaim transition은 official regular-session의 완전히 닫힌 1-minute bar만 사용해야 하고 (SHALL), live quote는 ARMED 이후 spread/drift/freshness veto에만 사용할 수 있다 (SHALL).

#### Scenario: intrabar breakout touch
- **WHEN** 아직 닫히지 않은 bar의 high가 resistance를 넘었지만 close가 확정되지 않았다
- **THEN** `BREAKOUT_CLOSED`로 전이하지 않고 proposal과 broker call은 0건이다

#### Scenario: deterministic replay
- **WHEN** 같은 config와 같은 ordered evidence snapshot을 서로 다른 process/timezone에서 평가한다
- **THEN** state, refusal/proposal, prices, quantity와 lineage digest가 byte-equivalent다

### Requirement: breakout setup은 봉인된 상태기계와 terminal lifecycle을 따른다
Setup ID는 canonical `(market,symbol,session_id,calendar_version,opening_range_first_bar_id,opening_range_last_bar_id,lane_id,lane_version,config_digest)`의 SHA-256이어야 하며 (SHALL), bar revision은 setup ID가 아니라 immutable snapshot/evidence digest에 포함되어야 한다 (SHALL). Canonical byte preimage는 UTF-8 `tossos.breakout.setup.v1` domain과 위 9개 field를 정확히 그 순서로 NUL byte (`0x00`) 하나씩 사이에 두고 join하며 terminal delimiter를 붙이지 않아야 한다 (SHALL). 각 field는 non-empty canonical value여야 하고 leading/trailing whitespace와 NUL을 거부해야 하며 (SHALL), hash 전에 trim, case fold, Unicode normalization 또는 JSON serialization을 수행해서는 안 된다 (MUST NOT). Setup ID 표현은 `sha256:` 뒤 64자리 lowercase hexadecimal이어야 한다 (SHALL). Setup은 `DISCOVERED → RANGE_LOCKED → BREAKOUT_CLOSED → RETEST_WAIT → RECLAIMED → ARMED → PROPOSED`의 forward transition만 허용해야 하며 (SHALL), `INVALIDATED`, `TIMED_OUT`, `CONSUMED`를 terminal state로 가져야 한다 (SHALL). Transition precondition을 건너뛰거나 같은 setup ID의 terminal state를 부활시키거나 첫 breakout touch에서 proposal을 만들면 안 된다 (MUST NOT). 같은 session의 pre-terminal correction은 같은 setup의 새 snapshot을 replay하지만 terminal 이후 또는 `PROPOSED`와 `CONSUMED` 사이 correction은 state/proposal seal/first-leg authority를 부활·교체·재발행할 수 없고 (MUST NOT), `CORRECTION_AFTER_PROPOSAL` 진단만 append할 수 있다 (SHALL). 새 regular session은 새 setup ID를 만들어야 한다 (SHALL).

#### Scenario: canonical setup ID known vector
- **WHEN** fields are `KR`, `005930`, `KRX:2026-08-18`, `krx-calendar-v1`, `bar-20260818-0900`, `bar-20260818-0914`, `kr_short_breakout_retest_v1`, `v1`, `sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`
- **THEN** setup ID is exactly `sha256:d2d5e2da006b841e45a1d2991624c5316b520a3900bb780653c79f455eeef04c`
- **AND** adding a terminal NUL, changing field order/case, trimming a field, or JSON-encoding the field list produces no accepted setup identity

#### Scenario: 정상 breakout-retest-reclaim
- **WHEN** opening range가 lock되고 buffered breakout close, 허용 범위 retest와 closed-bar reclaim이 순서대로 증명된다
- **THEN** setup은 각 중간 state를 정확히 한 번 거쳐 ARMED가 되고 final quote veto까지 통과한 뒤에만 PROPOSED가 된다

#### Scenario: terminal setup 재평가
- **WHEN** TIMED_OUT setup에 이후 강한 reclaim bar가 추가된다
- **THEN** 같은 setup ID는 TIMED_OUT을 유지하고 새 proposal을 만들지 않는다

#### Scenario: corrected opening-range bar
- **WHEN** 같은 session의 opening-range bar correction이 terminal 전후에 higher revision으로 도착한다
- **THEN** setup ID는 유지되고 새 snapshot digest만 생성되며 terminal 전에는 결정적으로 replay하고 terminal 뒤에는 terminal/proposal authority를 부활시키지 않는다

#### Scenario: correction between PROPOSED and CONSUMED
- **WHEN** first-leg proposal seal이 발행됐지만 shared admission이 아직 CONSUMED를 기록하기 전에 같은 setup bar correction이 도착한다
- **THEN** 기존 proposal identity만 유지되고 새 proposal seal/first-leg authority는 0건이며 correction 진단만 append된다

#### Scenario: 다음 regular session
- **WHEN** 같은 symbol이 새 official session/calendar identity에서 opening range를 시작한다
- **THEN** 이전 setup과 다른 setup ID를 만들고 이전 terminal state를 상속하지 않는다

### Requirement: breakout v1 threshold는 server-owned experimental config로 versioning된다
v1 config는 regular-session opening range 15분, breakout close buffer `max(1 tick, 100,000 ppm × ATR)`, retest tolerance `100,000..250,000 ppm × ATR`, timeout KR 8/US 10 closed 1-minute bars, RVOL minimum 1,500,000 ppm, upper-wick/range maximum 350,000 ppm, positive finite `max_quote_age_ms`, `max_spread_ppm`, `max_entry_drift_ppm`을 포함해야 한다 (SHALL). Evaluation authority time은 immutable snapshot의 `evaluated_at`이고 quote age는 `evaluated_at-source_observed_at`이어야 하며 (SHALL), future source time 또는 `source_observed_at > received_at > evaluated_at` 순서 위반은 typed refusal이어야 한다 (SHALL). Buy-side spread는 `ceil((ask-bid)×1_000_000 / floor((ask+bid)/2))`, entry drift는 `ceil(abs(ask-proposed_entry)×1_000_000 / proposed_entry)`로 계산해야 하고 (SHALL), 모든 가격/분모는 positive integer minor units여야 한다 (SHALL). Quote age/spread/absolute entry drift는 한계와 같을 때 허용하고 한 단위라도 크면 각각 `QUOTE_STALE`, `SPREAD_TOO_WIDE`, `ENTRY_DRIFT_EXCEEDED` refusal이어야 한다 (SHALL). RVOL 1,200,000/2,000,000/2,500,000 ppm counterfactual 결과도 evidence에 기록해야 하며 (SHALL), config/version/digest 변경은 진행 중 setup을 소급 재해석해서는 안 된다 (MUST NOT).

#### Scenario: breakout close buffer 미달
- **WHEN** closed bar가 resistance를 넘지만 초과폭이 max(1 tick, 0.10 ATR)보다 작다
- **THEN** breakout은 확정되지 않고 threshold/config provenance와 typed refusal이 기록된다

#### Scenario: 시장별 timeout
- **WHEN** KR setup은 breakout 뒤 8개, US setup은 10개의 closed 1-minute bar 안에 유효 reclaim을 만들지 못한다
- **THEN** 각 setup은 정확한 시장 한계에서 TIMED_OUT이 되고 proposal은 0건이다

#### Scenario: upper wick veto
- **WHEN** breakout bar upper-wick/range가 350,000 ppm을 초과한다
- **THEN** setup은 proposal을 만들지 않고 wick veto와 measured PPM을 기록한다

### Requirement: breakout invalidation과 first-touch 금지는 fail-closed다
Lane은 invalidation level 아래 close, volume-expanded failed reclaim, timeout, stale/wide-spread quote 또는 excessive entry drift를 typed invalidation/refusal로 처리해야 한다 (SHALL). Failed reclaim 이후 더 낮은 가격으로 평균단가를 낮추거나 stop을 entry에서 멀어지게 하는 proposal을 만들면 안 되며 (MUST NOT), first-touch breakout order는 금지된다 (MUST NOT).

#### Scenario: failed reclaim과 volume expansion
- **WHEN** retest 뒤 reclaim이 실패하고 volume expansion evidence가 threshold를 넘는다
- **THEN** setup은 INVALIDATED가 되고 이후 averaging-down leg와 broker request는 0건이다

#### Scenario: final quote drift
- **WHEN** ARMED setup의 current official quote가 config의 freshness 또는 entry drift 한계를 위반한다
- **THEN** proposal은 typed quote refusal로 종료되고 closed-bar state를 성공으로 위조하지 않는다

### Requirement: breakout evidence는 strict append-only snapshot으로 재생 가능하다
Breakout evidence는 market/symbol/session/calendar, setup ID, ordered closed bar IDs/revisions, opening range, resistance, ATR, RVOL/counterfactuals, wick/body PPM, transitions, quote freshness/spread, entry/stop/target/cost assumptions, source timestamps와 config/evidence digests를 canonical integer minor/PPM units로 포함해야 한다 (SHALL). Toss rank/volume/flow source는 candidate discovery evidence로만 취급하고 breakout/session/tradability/order authority로 사용해서는 안 되며 (MUST NOT), transition과 final veto는 current official adapter evidence로 검증해야 한다 (SHALL). Unknown field/enum, float, secret-like field, duplicate/out-of-order/unbounded bar, unfinished/future bar 또는 digest mismatch는 snapshot seal 전에 거부해야 한다 (SHALL). Correction은 기존 row를 덮지 않고 additive revision으로 저장해야 한다 (SHALL).

#### Scenario: bar correction
- **WHEN** official source가 이미 기록된 closed bar의 correction을 제공한다
- **THEN** higher revision과 new digest가 append되고 이전 snapshot replay 결과는 변하지 않으며 새 snapshot만 correction을 본다

#### Scenario: malformed payload
- **WHEN** payload가 float price, unknown state 또는 duplicate bar identity를 포함한다
- **THEN** strict decoder가 evidence를 거부하고 setup transition과 proposal은 0건이다

#### Scenario: Toss rank만 존재
- **WHEN** candidate가 Toss ranking evidence를 가지지만 current official closed bars 또는 session/tradability evidence가 없다
- **THEN** lane은 breakout을 추정하지 않고 typed evidence refusal과 proposal 0건을 반환한다

#### Scenario: US lossless official source가 증명되지 않음
- **WHEN** US 1-minute source가 raw decimal/timestamp 보존, pagination, USD currency, regular-session/calendar identity와 rate-budget semantics를 M-B receipt로 증명하지 못했다
- **THEN** US breakout production adapter와 exact 8-lane activation authority는 OFF/UNOBSERVED를 유지하고 WTS, float-adapted candle 또는 test fixture로 대체하지 않으며 exposure-raising broker request는 0건이다

### Requirement: M-B0/M-B1 measurement는 production evidence authority와 분리된다
M-B0 official seam은 A112 source measurement만을 위한 unused capability여야 하며 (SHALL), 기존 KR `RawMinuteCandles`, 일반 retry/refresh transport, token cache writer, account discovery, WTS/hybrid, float/time adapter 또는 product runtime caller를 변경·호출해서는 안 된다 (MUST NOT). Factory는 exact `*official.Client` 하나만 받아 같은 instance의 sealed authority origin/transport를 다시 검증해야 하고 (SHALL), caller-supplied origin token, 별도 provider 또는 cross-client result/trace를 결합해서는 안 된다 (MUST NOT). Unix seam은 cached token path를 descriptor-relative no-follow로 열고 opened FD를 current-UID regular 0600으로 `Fstat`해야 하며 (SHALL), unsupported platform은 request 0건 HOLD여야 한다 (SHALL). Existing 60-second validity skew를 통과한 cached token만 사용하고 (SHALL), cloned transport는 `Proxy=nil`, `DisableCompression=true`, redirect/non-GET refusal, `Accept-Encoding: identity`와 default User-Agent suppression을 강제해야 한다 (SHALL). On-wire application header는 `Authorization`, optional `Accept`, `Accept-Encoding: identity`뿐이어야 하고 protocol-owned Host만 추가로 허용하며 (SHALL), account/cookie/client-secret/User-Agent header는 0개여야 한다 (SHALL). Caller deadline은 15초 이하여야 하고, `Content-Length` precheck와 2-MiB-plus-one read limiter를 모두 적용하며, non-empty/non-identity Content-Encoding을 거부해야 한다 (SHALL). Injected/process clock은 cached-token expiry와 deadline enforcement에만 사용할 수 있고 (SHALL), 그 값을 official evidence로 직렬화하거나 candle finality/source observation time을 추론하는 데 사용해서는 안 된다 (MUST NOT). Exact candle `US/AAPL/1m/count=200/adjusted=false`, raw AAPL orderbook와 explicit-date US calendar descriptor, response status, raw cursor JSON와 decoded UTF-8 value bytes, allow-listed raw headers `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After`를 같은 request에 결속해 decode 전 opaque evidence로 보존해야 한다 (SHALL). Success의 Limit/Remaining/Reset은 각각 정확히 한 valid value, Retry-After는 0개여야 하며 (SHALL), 429는 exactly-one valid Retry-After diagnostic만 허용하고 evidence를 mint하지 않아야 한다 (SHALL). Duplicate/missing/malformed/negative header, cache miss/expiry, unsafe ownership/mode, deadline missing/over-limit, redirect, non-GET, 401/403/429/5xx, body overflow, malformed envelope/cursor/rate field는 exchange, refresh, retry, fallback 또는 evidence mint 없이 HOLD여야 한다 (SHALL).

M-B1 collector는 설치되지 않는 `tools/a112-mb-us-source` one-shot tool이어야 하며 (SHALL), required explicit cached-token path, secure `/tmp` receipt root, exact `YYYY-MM-DD` session date와 initial cursor 외에는 시장·symbol·interval·count·adjusted 값을 입력받아서는 안 된다 (MUST NOT). It constructs the default-origin official client with empty credentials and no account/config resolver, and only the M-B0 direct cache-only seam may access the token (SHALL); invalid cache must not fall through to OAuth (MUST NOT). Collector는 root descriptor와 openat/no-follow/Fstat/O_EXCL/directory-fsync를 사용해 0700 root와 owner-only 0600 sentinel/payload capability를 network 전에 검증하고 (SHALL), unsupported platform은 request 0건 HOLD해야 한다 (SHALL). Candle 최대 4페이지와 orderbook/calendar 각 1회, request당 15초 및 전체 120초 deadline에서만 동작하고 정확히 120초부터 HOLD하며 15초가 남지 않으면 새 request를 시작해서는 안 된다 (MUST NOT). Identity, payload write, seal과 success 전후에 cancellation/deadline을 다시 검사해야 한다 (SHALL). 직전 decoded cursor UTF-8 value bytes를 trim/casefold/normalization 없이 URL percent-encoding만 한 번 적용해 전달하고 raw JSON과 exact canonical query를 별도 보존해야 하며 (SHALL), terminal은 raw JSON null만 허용해야 한다 (SHALL). Candle 다음 page는 직전 same-endpoint unique headers의 parsed remaining이 1 이상일 때만 호출할 수 있고 (SHALL), loop/absent/empty/non-string cursor를 만나면 HOLD해야 한다 (SHALL). 4페이지 cap 소진은 HOLD가 아니라 sealed manifest 안의 `candle-crawl.json`(`schema a112-mb-us-candle-crawl:v1`, `pages`, `terminal` = `null`|`cap_exhausted`, cap 소진 시 `last_cursor_sha256`)에 기록하고 orderbook/calendar/post identity/seal로 계속 진행해야 하며 (SHALL), 그 뒤에도 추가 candle request를 보내서는 안 된다 (MUST NOT). Candle PASS 기준은 terminal null이 아니라 explicit regular-session full coverage여야 한다 (SHALL): session-date의 explicit-date calendar body가 정한 regular session의 모든 1분봉이 raw body에 결손 없이 존재하고, 그 시작 이전 봉이 최소 1개 관측되며, USD·decimal string·cursor continuity·unique rate headers가 함께 성립해야 한다 (SHALL). Raw null terminal이 관측되면 그것만 terminal로 기록하고 (SHALL), terminal null 미관측은 그 자체로 HOLD 사유가 아니다. External M-B run 전에는 같은 OpenAPI credential의 모든 보유자가 token cache를 공유하거나 정지 상태임을 사람이 확인·기록해야 한다 (SHALL). Raw JSON은 모든 depth의 duplicate key, invalid UTF-8/surrogate 및 secret-like key/value를 fail-closed한 뒤에만 저장할 수 있다 (SHALL). Pinned root/run FD는 seal 진입·종료에 current-UID exact0700 directory로, payload는 current-UID regular exact0600으로 다시 검증해야 한다 (SHALL). 각 payload는 최대 64 MiB의 전체 bytes와 동일 FD의 device/inode/size/read length를 결속해 재해시하고, strictly parsed deterministic self-excluding manifest가 exact entry set과 digest를 포함하며 manifest fsync 뒤 original manifest bytes와 payload를 다시 검증해야 한다 (SHALL). Unsafe FUSE mode, unexpected entry, mode/owner/content/manifest drift, appended tail 또는 overflow를 POSIX-safe로 간주하거나 repo receipt로 fallback해서는 안 된다 (MUST NOT). Network 전후 exact GOOS-bound source/tool/module/Go/Git/base/worktree/compiled-closure/build/executable identity hashes가 같아야 한다 (SHALL). Git은 vetted absolute binary, canonical root와 sanitized no-`GIT_*` environment에서만 실행해야 하고 (SHALL), source identity read는 no-follow regular file로 256 MiB 초과를 거부해야 한다 (SHALL). Rebuild는 caller GO/CC/CXX/CGO 환경을 상속하지 않고 offline private cache에서 `go mod verify`와 machine-readable `go list -json -deps`를 실행해 exact compiled Go/Cgo/embed closure를 증명하며, allowlist 밖 input을 거부하고 running executable SHA를 재현해야 한다 (SHALL). Identity-unknown `go run`, caller-claimed command, ad-hoc binary 또는 drift는 HOLD여야 한다 (SHALL). M-B0 exported-symbol references are allowed only in their definitions, M-B0 tests and the exact tool; ordinary pre-existing official-package imports are unaffected (SHALL). Completion은 M-B PASS·L1 acceptance·closed-bar authority·activation·dispatch·deployment 또는 exposure를 만들지 않는다 (MUST NOT).

M-B1은 required explicit absolute `--go-binary`를 받아야 하고 (SHALL), 그 leaf를 no-follow regular file로 resolve/hash/revalidate한 동일 binary만 `go env`, `go mod verify`, `go list`와 prescribed rebuild에 사용해야 한다 (SHALL). `runtime.GOROOT`, ambient `GOROOT`, `PATH` 또는 inherited `GO*`를 locator/build authority로 사용해서는 안 된다 (MUST NOT). Exact prescribed `-trimpath` collector가 offline identity preflight를 통과하고 recording reader call 0을 유지하는 통합 테스트 전에는 external measurement를 다시 실행해서는 안 된다 (MUST NOT).

Go와 Git command는 검증 뒤 caller-controlled original pathname을 다시 열어 실행해서는 안 되고 (MUST NOT), opened no-follow regular FD의 exact bytes와 digest가 같은 freshly created current-UID 0700 `/tmp` directory의 O_EXCL owner-only executable snapshot capability로만 실행해야 한다 (SHALL). Snapshot file/directory fsync, mode/owner/type/digest와 unexpected-entry 검증을 통과해야 하며 (SHALL), pathname swap/restore 또는 snapshot drift는 reader call 0에서 HOLD해야 한다 (SHALL). Complete machine-derived tracked dependency closure는 frozen base, exact diff와 per-input SHA evidence로 결속하고 (SHALL), untracked M-B0/tool production input은 exact GOOS allowlist 밖이면 거부해야 한다 (SHALL).

Private Go executable은 unbound original GOROOT를 사용해서는 안 된다 (MUST NOT). Selected `<go-root>/bin/go` distribution을 root-FD에서 descriptor-relative/no-follow로 순회해 symlink/special file을 거부하고, 512 MiB/50,000-entry cap 안의 각 regular input을 pre/post device/inode/size/read length/SHA로 결속한 deterministic owner-only private-root manifest를 만들어야 한다 (SHALL). O_EXCL copy와 모든 file/directory fsync 뒤 재검증된 private root만 internally set `GOROOT`로 사용할 수 있고 (SHALL), 각 Go command 전후 entry/mode/owner/content drift를 검사하며 모든 success/HOLD path에서 cleanup해야 한다 (SHALL).

Secure Go-distribution snapshot과 prescribed identity/rebuild는 동일한 120-second collector deadline 안에서 완료되어야 한다 (SHALL). Fixed finite worker pool은 pre-enumerated descriptor-validated inventory만 parallel copy할 수 있고 (SHALL), per-file no-follow/pre-post metadata/digest/fsync, deterministic directory fsync, sorted manifest, first-error cancellation과 complete cleanup을 보존해야 한다 (SHALL). Deadline 밖 준비나 durability 완화는 금지한다 (MUST NOT).

#### Scenario: cached token이 만료됨
- **WHEN** M-B0가 없거나 만료됐거나 owner/mode가 잘못된 token cache를 읽는다
- **THEN** OAuth exchange, token/cache write, account discovery와 data GET은 0건이고 typed HOLD만 반환한다

#### Scenario: bounded pagination이 cap을 소진함
- **WHEN** M-B1이 고정 4페이지 cap 안에서 `nextBefore == null`을 관측하지 못했지만 cursor loop나 malformed cursor는 없다
- **THEN** 추가 candle request, retry, sleep/backoff와 fallback은 0건이고 receipt는 `cap_exhausted`를 기록하며 orderbook/calendar/post identity/seal은 계속 진행되고, PASS 여부는 regular-session full coverage 판정에 달린다

#### Scenario: cursor loop 또는 malformed cursor
- **WHEN** M-B1이 이미 본 cursor를 다시 받거나 absent/empty/non-string cursor를 받는다
- **THEN** 추가 request, retry, sleep/backoff와 fallback은 0건이고 M-B는 HOLD를 유지한다

#### Scenario: regular session이 완전히 덮이지 않음
- **WHEN** receipt의 raw candle body가 session-date의 explicit-date calendar regular session 안에서 1분봉 하나라도 빠뜨리거나 세션 시작 이전 봉을 하나도 담지 못한다
- **THEN** M-B는 HOLD를 유지하고 L1은 시작하지 못한다

#### Scenario: measurement receipt가 PASS함
- **WHEN** independent reviewer와 Manager가 exact raw US body, cursor continuity, explicit regular-session full coverage(calendar body join), USD, decimal string, raw quote, unique rate headers, redaction, modes와 self-excluding manifest를 모두 검증한다
- **THEN** L1 implementation만 시작할 수 있고 US closed-bar authority, L1 acceptance, lane activation과 broker request는 여전히 0건이다

#### Scenario: candle finality field가 없음
- **WHEN** official candle page가 regular-session timestamp와 raw OHLCV를 반환하지만 publisher `closed/finalized` field 또는 server-observed timestamp를 제공하지 않는다
- **THEN** M-B receipt는 그 bar를 closed authority로 승격하지 않고 L1 dual-cutoff/correction decoder가 별도 acceptance를 통과할 때까지 US production evidence를 만들지 않는다

### Requirement: breakout 수량은 비용 포함 risk로 제안되고 q_final을 넘지 않는다
Quote seal은 bid/ask/last integer minor units, instrument currency, source/received timestamp와 digest를 포함해야 한다 (SHALL). FX seal은 account/instrument currency, canonical direction `ACCOUNT_MINOR_TO_INSTRUMENT_MINOR`, positive integer `rate_num/rate_den`, as-of/fresh-until과 digest를 포함해야 한다 (SHALL). Account-currency capacity `x`는 `floor(x×rate_num/rate_den)` instrument minor units로 변환하고, account-currency per-share cost `c`는 `ceil(c×rate_num/rate_den)`으로 변환해야 한다 (SHALL); 역방향 provider quote는 adapter가 numerator/denominator를 뒤집어 이 canonical direction으로 seal해야 한다 (SHALL). Buy-side `worst_entry = max(proposed_entry, ask) + entry_slippage_minor`이어야 한다 (SHALL). 모든 risk budget/notional/cost를 instrument minor currency로 정렬한 뒤 Lane은 `risk_per_share = entry-stop + entry_slippage + exit_slippage + round_trip_costs`와 `q_candidate = floor(min(risk_budget/risk_per_share, notional_cap/worst_entry))`를 overflow-safe exact arithmetic으로 계산해야 한다 (SHALL). Capacity conversion은 아래로, per-share cost/risk는 위로 보수적으로 round해야 하며 quantity를 exact rational floor보다 키워서는 안 된다 (MUST NOT). Stop missing/non-protective, target below break-even 또는 Guardian minimum RR, missing/stale/mismatched cost/FX/price, non-positive risk, overflow와 q_candidate 0은 typed refusal이어야 한다 (SHALL). Guardian/bucket admission은 q_candidate를 q_final로 같거나 작게 축소할 수 있으며 lane 또는 runtime은 이를 다시 늘리면 안 된다 (MUST NOT).

#### Scenario: cost가 수량을 축소
- **WHEN** slippage와 round-trip costs를 포함한 risk-per-share가 순수 entry-stop 폭보다 크다
- **THEN** q_candidate는 비용 포함 값으로 floor되고 비용 제외 수량을 사용하지 않는다

#### Scenario: q_final cap
- **WHEN** breakout q_candidate가 10주이고 shared Guardian/bucket q_final이 3주다
- **THEN** dispatch intent는 최대 3주이며 proposal lineage에 q_candidate 10과 q_final 3의 provenance가 보존된다

### Requirement: breakout v1 production 권위는 first-leg 하나로 제한된다
Breakout v1은 하나의 PROPOSED setup에서 existing shared dispatch가 소비할 first-leg proposal 최대 하나만 발행해야 한다 (SHALL). Proposal replay, duplicate bar delivery, correction 또는 restart가 동일 setup/bar에서 두 번째 first-leg 권위를 만들면 안 되며 (MUST NOT), a066 scale-in lifecycle과 별도 activation 승인이 완료되기 전 추가 exposure leg는 0건이어야 한다 (MUST NOT).

#### Scenario: duplicate evaluation
- **WHEN** 같은 setup, terminal bar, snapshot과 config가 retry/restart로 반복 평가된다
- **THEN** proposal identity는 멱등이고 shared journal admission에서 first-leg authority 최대 하나만 존재한다
