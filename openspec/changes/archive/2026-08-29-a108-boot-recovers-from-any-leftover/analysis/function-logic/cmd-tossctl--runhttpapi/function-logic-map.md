# Function Logic Map: `runHTTPAPI`

- Source: `cmd/tossctl/httpapi.go` (121-217)
- AST evidence: `ast.json` — AST branches 16 · returns 14 · defers 6
  (source_sha256 `37c2ed8990337096cb98973cba78f1d0e9cc61abade8df115106d754761ef2df` 는
  파일 해시이며 도구가 채운다, **gstack Fix 라운드 편집 후 재생성**)
- Risk scan: `risk-pattern-report.md`
- 편집 이력 세 라운드:
  1. 원 라운드 — `Dial` 실패의 `return fmt.Errorf("…unavailable: %w")` 를 지웠다.
     그 줄이 httpapi 컨테이너를 `Restarting (1)` 로 돌게 만든 것이다.
  2. Fix 라운드(6.8①) — `return fmt.Errorf("httpapi: inspect strategy runtime
     projection: %w")` 도 지웠다. **전략 endpoint 때문에 죽는 경로가 0 이 됐다.**
  3. **gstack 라운드 — 전략 블록 자체가 이 함수에서 나갔다.** 판정은
     `strategyRuntimeReaderFor` 로 옮겼고, 이 함수는 그 결과를 받아 두 곳에 넘긴다.
     분기 22 → 16, return 은 14 로 불변이다(사라진 여섯은 전부 return 없는 분기였다).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 허용 | cobra | nil 이면 배경 ctx (B1) |
| 종료 시그널 | SIGINT/SIGTERM | `signal.NotifyContext` | ctx 종료 → graceful shutdown (B13·B15·B16) |
| `--port/--bind/--allowed-cidr/--public-url` | 사설 경계만 | `httpAPIBoundary` | 공인 bind·CIDR·TLS 모호는 거절 (B2) |
| 직접 TLS 인증서 | `--tls-cert`/`--tls-key` 둘 다 | `networkboundary.LoadServerCertificate` | 로드 실패는 거절 (B3·B4) |
| 원장·성과·최적화 저장소 | 없어도 되는 것과 아닌 것이 갈린다 | `openHTTPAPIResources` | `reader.validate()` 실패는 거절 (B5) |
| **전략 endpoint 상태** | 부재·조사불가·dial실패·성공 넷 다 정상 입력 | **`strategyRuntimeReaderFor`** | **거절 없음.** 넷이 reader 값 셋(nil · sentinel · client)으로 접히고, 구별은 stderr 문구와 화면 refusal code 가 진다 |
| 스냅샷 캐시·스트림·라우터·서버 | 전부 필요 | `httpapi.*` | 실패는 거절 (B6~B10) |

**관통 불변식:** 조회 전용 데몬은 **전략 endpoint 때문에** 뜨지 못하는 일이 없다 —
잔재든 환경 이상이든. 뜨지 못하는 이유로 남는 것은 경계 설정 오류와 자기 저장소
(원장·성과·최적화) 열기 실패뿐이다.

「삼키지 않는다」는 요구는 **경고 문구**가 진다: 부재는 조용하고, 조사 불가와 dial
실패는 각각 다른 note 를 찍는다. 세 상태를 stderr 에서 구별할 수 있다.

**gstack 라운드가 더한 불변식:** 강등의 **값**도 구별된다. nil 은 「이 배포는 전략
화면을 안 쓴다」(`NOT_CONFIGURED`)이므로 붙지 못한 endpoint 를 nil 로 접지 않는다 —
dial 실패는 항상-에러 sentinel 로 와서 `RUNTIME_UNAVAILABLE` 로 그려진다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ctx == nil` | 없음 | 없음 | 간접 |
| B2 | 경계 구성 실패 | 없음 | 원인 그대로 | `TestHTTPAPIBoundaryRejectsPublicAndAmbiguousTLSConfiguration` |
| B3 | 직접 TLS 모드 | 인증서 로드 | 없음 | `TestHTTPAPIDefaultsStayLoopbackNoTokenReadOnlyAndRequireTLS` |
| B4 | 인증서 로드 실패 | 없음 | 원인 그대로 | 미고정 |
| B5 | 리소스 열기 실패 | 없음 | 원인 그대로 | 미고정 |
| B6 | 스냅샷 캐시 실패 | 없음 | 원인 그대로 | 미고정 |
| B7 | 스트림 생성 실패 | 없음 | 원인 그대로 | 미고정 |
| B8 | mutation 라우트 구성 실패 | 없음 | 원인 그대로 | `TestHTTPAPIMutationsStayAbsentUntilBothSecurityArtifactsAreConfigured` |
| B9 | 라우터 생성 실패 | 없음 | 원인 그대로 | 미고정 |
| B10 | 서버 생성 실패 | 없음 | 원인 그대로 | 미고정 |
| B11 | 직접 TLS 면 TLSConfig 설정 | 서버 상태 | 없음 | 미고정 |
| B12 | 직접 TLS 면 ListenAndServeTLS | listener | 없음 | 미고정 |
| B13 | serve 종료 vs ctx 종료 | 없음 | 분기 | `TestADeadDescriptorDoesNotStopTheDaemon` (배너 관측 → ctx 취소 경로) |
| B14 | `http.ErrServerClosed` | 없음 | nil | 미고정 |
| B15 | Shutdown 실패 | 없음 | 원인 그대로 | 미고정 |
| B16 | serve 오류가 ErrServerClosed 아님 | 없음 | 원인 그대로 | 미고정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httpAPIBoundary` | 사설 경계·TLS 모드 확정 | 실패는 거절 | AST · httpapi_test.go |
| `openHTTPAPIResources` | 원장·성과·최적화·reader | `reader.validate()` 가 최종 관문. **경고 writer 를 받는다**(gstack A3) | AST |
| **`strategyRuntimeReaderFor`** | 전략 endpoint 판정 전부 | **거절 없음** — 넷이 값 셋으로 접힌다 | AST · a108 daemon 테스트 |
| `httpapi.NewRouter` | 라우팅 | `StrategyRuntime` 은 nil 도 sentinel 도 받는다 | AST |
| `server.ListenAndServe(TLS)` | 서비스 | `ErrServerClosed` 는 정상 종료 | AST |

## State mutations and fallbacks

- 이 함수는 **디스크를 쓰지 않는다.** 특히 잔재 descriptor 를 **치우지 않는다** — 잔재
  정리는 엔진(겹1)의 일이고, 조회 전용 데몬이 남의 control 디렉터리를 치우기 시작하면
  그것이 새 사고 유형이 된다. `TestAnAbsentDescriptorAndADeadOneBootTheSame` 가 고정한다.
- 강등의 결과 상태는 이제 **둘**이다: `nil`(부재·해석 실패·조사 불가)과 항상-에러
  sentinel(dial 실패). 같은 값이 `resources.reader.strategyRuntime` 와
  `httpapi.NewRouter` 두 곳에 들어간다.
- 그 값이 화면에서 무엇이 되는지는 `httpAPIReader.Snapshot` 의 FLM 이 진다
  (`cmd-tossctl--httpapireader.snapshot/`): nil 은 dormant(NOT_CONFIGURED)이고,
  읽지 못하는 reader 는 unavailable(RUNTIME_UNAVAILABLE)이다.
- 재-dial 은 없다. descriptor 부재로 뜬 데몬도 지금 그렇게 남으며, 대칭을 깨는 쪽이
  새 비대칭을 만든다(design D4, 선언된 생략).

## Safety conclusion

- Safe edit boundary: 전략 블록의 **추출**과 `openHTTPAPIResources` 의 인자 하나.
  B1~B5·B13~B16 의 판정과 순서는 그대로다. 추출은 판정을 옮긴 것이지 바꾼 것이
  아니며, 옮긴 뒤 그 판정을 직접 부르는 핀
  (`TestADialFailureRendersUnavailableRatherThanNotConfigured`)이 생겼다.
- High-risk impact: **no** (이 함수 자체는). `tossctl httpapi` 는 조회 전용이고
  `mutating` 주석이 없다. 다만 사고 당시 이 데몬의 crash loop 가 운영자의 유일한 화면을
  없앴으므로, 가용성 관점에서 a108 의 세 겹 중 하나로 다룬다.
