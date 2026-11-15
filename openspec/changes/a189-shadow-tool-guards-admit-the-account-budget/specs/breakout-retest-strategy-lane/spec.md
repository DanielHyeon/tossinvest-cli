## MODIFIED Requirements

### Requirement: breakout 그림자 신호 측정은 주문 권위가 없다
breakout 신호의 그림자 측정은 브로커에 읽기 요청만 보내야 하며 (SHALL), 주문 생성 · 정정 · 취소, 토글 · 레인 활성화 · 운영 모드 쓰기,
엔진 저널 쓰기를 해서는 안 된다 (MUST NOT). 측정의 evidence 쓰기는 자기 전용 evidence 저장소에만 해야 하고 (SHALL), 엔진이 생산 제안 재생에 쓰는
`evidence.db` 와 같은 파일(심볼릭 링크 해소 뒤)이면 시작을 거부해야 한다 (SHALL). 측정은 서명 매니페스트를 읽거나 만들지 않아야 하고
(MUST NOT), 외부 판단 API(Jev)를 호출해서는 안 된다 (MUST NOT). ARMED 에 닿은 setup 은 처음 본 분부터 최대 5 분 동안 분마다 새 호가로 다시 판정해야 하며 (SHALL), 처음 PROPOSED, 5 번째 거절, 또는 처음 본 뒤 5 분이 지남(호가 읽기가 실패해도) 중 먼저 오는 것에서 그 setup 의 판정을 끝내고 시도마다 호가 나이 · 스프레드 · 진입 괴리 실측값을 원장에 남겨야 한다 (SHALL).
측정이 만든 「냈을 제안」은 어떤 발주 경로 · 조정자 · Guardian 에도
전달되어서는 안 된다 (MUST NOT).
측정 도구는 계좌 공유 요청 예산에 보조 등급 참여자로 연결될 수 있다. 그 연결은 예산 시작과 보조 우선순위 표시, 그 예산에 묶인 공식 클라이언트 생성으로 한정되어야 하며 (SHALL), 측정 도구의 정적 검사는 예산 패키지에서 쓰는 이름을 패키지별 허용 목록으로 고정해야 한다 (SHALL). 예산 참여는 읽기 요청의 전송 허가와 시점만 정하며, 주문 · 계좌 · 보유 · 매수 가능 표면을 열어서는 안 된다 (MUST NOT).

#### Scenario: 생산 evidence DB 경로를 준다
- **WHEN** 측정 도구의 evidence 저장소 경로가 엔진 저널 옆 `evidence.db` 를 가리킨다(직접 또는 심볼릭 링크)
- **THEN** 도구는 브로커 호출 없이 시작을 거부하고 사유를 말한다

#### Scenario: 측정 도구의 의존 금지
- **WHEN** 측정 도구 패키지의 import 와 호출 이름을 정적으로 검사한다
- **THEN** 주문 · 계좌 · 보유 · 매수 가능 · 토글 · 활성화 쓰기 이름이 0 건이다

#### Scenario: ARMED 의 첫 호가가 넓다
- **WHEN** ARMED 첫 분의 새 호가가 SPREAD_TOO_WIDE 이고 둘째 분의 새 호가는 한계 안이다
- **THEN** 둘째 분에 PROPOSED 가 판정되고 두 시도의 실측값이 원장에 남으며, 그 setup 은 다시 평가되지 않는다

#### Scenario: ARMED 에 닿은 setup
- **WHEN** 한 종목의 setup 이 ARMED 에 닿아 새 호가로 PROPOSED 가 판정된다
- **THEN** 그 결과는 원장 행으로만 남고 Guardian 판정 · 원장 예약 · 브로커 주문은 0 건이다

#### Scenario: 측정 도구가 계좌 공유 요청 예산에 참여한다
- **WHEN** 측정 도구 패키지가 계좌 공유 요청 예산 패키지를 import 하고 그 이름을 정적으로 검사한다
- **THEN** 허용되는 이름은 예산 시작과 보조 우선순위 표시뿐이고, 그 목록 밖 이름 · 허용 목록 밖 import · 주문 · 계좌 · 보유 · 매수 가능 · 토글 · 활성화 쓰기 이름은 여전히 검사를 실패시킨다

### Requirement: M-B0/M-B1 measurement는 production evidence authority와 분리된다
M-B0 official seam은 A112 source measurement만을 위한 unused capability여야 하며 (SHALL), 기존 KR `RawMinuteCandles`, 일반 retry/refresh transport, token cache writer, account discovery, WTS/hybrid, float/time adapter 또는 product runtime caller를 변경·호출해서는 안 된다 (MUST NOT). Factory는 exact `*official.Client` 하나만 받아 같은 instance의 sealed authority origin/transport를 다시 검증해야 하고 (SHALL), caller-supplied origin token, 별도 provider 또는 cross-client result/trace를 결합해서는 안 된다 (MUST NOT). Unix seam은 cached token path를 descriptor-relative no-follow로 열고 opened FD를 current-UID regular 0600으로 `Fstat`해야 하며 (SHALL), unsupported platform은 request 0건 HOLD여야 한다 (SHALL). Existing 60-second validity skew를 통과한 cached token만 사용하고 (SHALL), cloned transport는 `Proxy=nil`, `DisableCompression=true`, redirect/non-GET refusal, `Accept-Encoding: identity`와 default User-Agent suppression을 강제해야 한다 (SHALL). On-wire application header는 `Authorization`, optional `Accept`, `Accept-Encoding: identity`뿐이어야 하고 protocol-owned Host만 추가로 허용하며 (SHALL), account/cookie/client-secret/User-Agent header는 0개여야 한다 (SHALL). Caller deadline은 15초 이하여야 하고, `Content-Length` precheck와 2-MiB-plus-one read limiter를 모두 적용하며, non-empty/non-identity Content-Encoding을 거부해야 한다 (SHALL). Injected/process clock은 cached-token expiry와 deadline enforcement에만 사용할 수 있고 (SHALL), 그 값을 official evidence로 직렬화하거나 candle finality/source observation time을 추론하는 데 사용해서는 안 된다 (MUST NOT). Exact candle `US/AAPL/1m/count=200/adjusted=false`, raw AAPL orderbook와 explicit-date US calendar descriptor, response status, raw cursor JSON와 decoded UTF-8 value bytes, allow-listed raw headers `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After`를 같은 request에 결속해 decode 전 opaque evidence로 보존해야 한다 (SHALL). Success의 Limit/Remaining/Reset은 각각 정확히 한 valid value, Retry-After는 0개여야 하며 (SHALL), 429는 exactly-one valid Retry-After diagnostic만 허용하고 evidence를 mint하지 않아야 한다 (SHALL). Duplicate/missing/malformed/negative header, cache miss/expiry, unsafe ownership/mode, deadline missing/over-limit, redirect, non-GET, 401/403/429/5xx, body overflow, malformed envelope/cursor/rate field는 exchange, refresh, retry, fallback 또는 evidence mint 없이 HOLD여야 한다 (SHALL).

M-B1 collector는 설치되지 않는 `tools/a112-mb-us-source` one-shot tool이어야 하며 (SHALL), required explicit cached-token path, secure `/tmp` receipt root, exact `YYYY-MM-DD` session date와 initial cursor 외에는 시장·symbol·interval·count·adjusted 값을 입력받아서는 안 된다 (MUST NOT). It constructs the default-origin official client with empty credentials and no account/config resolver, and only the M-B0 direct cache-only seam may access the token (SHALL); invalid cache must not fall through to OAuth (MUST NOT). Collector는 root descriptor와 openat/no-follow/Fstat/O_EXCL/directory-fsync를 사용해 0700 root와 owner-only 0600 sentinel/payload capability를 network 전에 검증하고 (SHALL), unsupported platform은 request 0건 HOLD해야 한다 (SHALL). Candle 최대 4페이지와 orderbook/calendar 각 1회, request당 15초 및 전체 120초 deadline에서만 동작하고 정확히 120초부터 HOLD하며 15초가 남지 않으면 새 request를 시작해서는 안 된다 (MUST NOT). Identity, payload write, seal과 success 전후에 cancellation/deadline을 다시 검사해야 한다 (SHALL). 직전 decoded cursor UTF-8 value bytes를 trim/casefold/normalization 없이 URL percent-encoding만 한 번 적용해 전달하고 raw JSON과 exact canonical query를 별도 보존해야 하며 (SHALL), terminal은 raw JSON null만 허용해야 한다 (SHALL). Candle 다음 page는 직전 same-endpoint unique headers의 parsed remaining이 1 이상일 때만 호출할 수 있고 (SHALL), loop/absent/empty/non-string cursor를 만나면 HOLD해야 한다 (SHALL). 4페이지 cap 소진은 HOLD가 아니라 sealed manifest 안의 `candle-crawl.json`(`schema a112-mb-us-candle-crawl:v1`, `pages`, `terminal` = `null`|`cap_exhausted`, cap 소진 시 `last_cursor_sha256`)에 기록하고 orderbook/calendar/post identity/seal로 계속 진행해야 하며 (SHALL), 그 뒤에도 추가 candle request를 보내서는 안 된다 (MUST NOT). Candle PASS 기준은 terminal null이 아니라 explicit regular-session full coverage여야 한다 (SHALL): session-date의 explicit-date calendar body가 정한 regular session의 모든 1분봉이 raw body에 결손 없이 존재하고, 그 시작 이전 봉이 최소 1개 관측되며, USD·decimal string·cursor continuity·unique rate headers가 함께 성립해야 한다 (SHALL). Raw null terminal이 관측되면 그것만 terminal로 기록하고 (SHALL), terminal null 미관측은 그 자체로 HOLD 사유가 아니다. External M-B run 전에는 같은 OpenAPI credential의 모든 보유자가 token cache를 공유하거나 정지 상태임을 사람이 확인·기록해야 한다 (SHALL). Raw JSON은 모든 depth의 duplicate key, invalid UTF-8/surrogate 및 secret-like key/value를 fail-closed한 뒤에만 저장할 수 있다 (SHALL). Pinned root/run FD는 seal 진입·종료에 current-UID exact0700 directory로, payload는 current-UID regular exact0600으로 다시 검증해야 한다 (SHALL). 각 payload는 최대 64 MiB의 전체 bytes와 동일 FD의 device/inode/size/read length를 결속해 재해시하고, strictly parsed deterministic self-excluding manifest가 exact entry set과 digest를 포함하며 manifest fsync 뒤 original manifest bytes와 payload를 다시 검증해야 한다 (SHALL). Unsafe FUSE mode, unexpected entry, mode/owner/content/manifest drift, appended tail 또는 overflow를 POSIX-safe로 간주하거나 repo receipt로 fallback해서는 안 된다 (MUST NOT). Network 전후 exact GOOS-bound source/tool/module/Go/Git/base/worktree/compiled-closure/build/executable identity hashes가 같아야 한다 (SHALL). Git은 vetted absolute binary, canonical root와 sanitized no-`GIT_*` environment에서만 실행해야 하고 (SHALL), source identity read는 no-follow regular file로 256 MiB 초과를 거부해야 한다 (SHALL). Rebuild는 caller GO/CC/CXX/CGO 환경을 상속하지 않고 offline private cache에서 `go mod verify`와 machine-readable `go list -json -deps`를 실행해 exact compiled Go/Cgo/embed closure를 증명하며, allowlist 밖 input을 거부하고 running executable SHA를 재현해야 한다 (SHALL). Identity-unknown `go run`, caller-claimed command, ad-hoc binary 또는 drift는 HOLD여야 한다 (SHALL). M-B0 exported-symbol references are allowed only in their definitions, M-B0 tests, the exact tool and the single account-budget runtime test file `internal/accountbudgetruntime/v2_token_test.go` that proves the M-B0 cloned transport is counted by the shared request budget; ordinary pre-existing official-package imports are unaffected (SHALL). Completion은 M-B PASS·L1 acceptance·closed-bar authority·activation·dispatch·deployment 또는 exposure를 만들지 않는다 (MUST NOT).

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
