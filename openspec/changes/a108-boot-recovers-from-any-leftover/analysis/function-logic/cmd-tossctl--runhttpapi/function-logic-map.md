# Function Logic Map: `runHTTPAPI`

- Source: `cmd/tossctl/httpapi.go` (120-251)
- AST evidence: `ast.json` — AST branches 22 · returns 14 · defers 6
  (source_sha256 `d2bf9f341beb169e869baa722f79a20f233447419806c0832ed36665c7e7ab42`,
  **Fix 라운드 6.8 편집 후 재생성**)
- Risk scan: `risk-pattern-report.md`
- 편집 대상: **B9·B10** (겹3). 기준(base) 판은 분기 21개·return 16개였고, 사라진 return 이
  `Dial` 실패의 `return fmt.Errorf("…unavailable: %w")` 다 — httpapi 컨테이너가
  `Restarting (1)` 로 돌게 만든 줄.
- **Fix 라운드(6.8①, design D4-2)**: **B11 도 강등이 됐다.** 분기 수는 22 로 불변이고
  return 이 15 → 14 로 하나 더 줄었다 — 사라진 것이
  `return fmt.Errorf("httpapi: inspect strategy runtime projection: %w")` 다.
  이제 이 함수에서 전략 endpoint 때문에 죽는 경로는 **하나도 없다**.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 허용 | cobra | nil 이면 배경 ctx (B1) |
| 종료 시그널 | SIGINT/SIGTERM | `signal.NotifyContext` | ctx 종료 → graceful shutdown (B19·B21·B22) |
| `--port/--bind/--allowed-cidr/--public-url` | 사설 경계만 | `httpAPIBoundary` | 공인 bind·CIDR·TLS 모호는 거절 (B2) |
| 직접 TLS 인증서 | `--tls-cert`/`--tls-key` 둘 다 | `networkboundary.LoadServerCertificate` | 로드 실패는 거절 (B3·B4) |
| 원장·성과·최적화 저장소 | 없어도 되는 것과 아닌 것이 갈린다 | `openHTTPAPIResources` | `reader.validate()` 실패는 거절 (B5) |
| 엔진 디렉터리 | `--config-dir` 를 따른다 | `engineJournalDir` | 해석 실패면 전략 화면 없이 진행 (B6 의 else 없음) |
| **projection descriptor 부재** | 정상 상태 | `os.Stat` | **강등** — 설계된 동작 (B7 의 else + B11 이 아닌 경로) |
| **projection dial 실패** | **정상 상태 (a108)** | `strategyprojectionrpc.Dial` | **강등 + 경고 로그 (B9)** |
| **descriptor 조사 불가** | 환경 이상 | `os.Stat` 의 비-NotExist 오류 | **경고 + 강등 (B11)** — 콘솔 패리티(console.go:419-421). 원 D4 의 fatal 결정을 A2 가 반증했다: ENOTDIR 볼륨 오배치가 crash loop 를 재생산한다 |
| 스냅샷 캐시·스트림·라우터·서버 | 전부 필요 | `httpapi.*` | 실패는 거절 (B12~B16) |

**관통 불변식(a108 Fix 라운드 이후):** 조회 전용 데몬은 **전략 endpoint 때문에** 뜨지
못하는 일이 없다 — 잔재든 환경 이상이든. 뜨지 못하는 이유로 남는 것은 경계 설정
오류와 자기 저장소(원장·성과·최적화) 열기 실패뿐이다.

「삼키지 않는다」는 요구는 **경고 문구**가 진다: 부재는 조용하고, dial 실패와 조사
불가는 각각 다른 note 를 찍는다. 세 상태를 stderr 에서 구별할 수 있다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ctx == nil` | 없음 | 없음 | 간접 |
| B2 | 경계 구성 실패 | 없음 | 원인 그대로 | `TestHTTPAPIBoundaryRejectsPublicAndAmbiguousTLSConfiguration` |
| B3 | 직접 TLS 모드 | 인증서 로드 | 없음 | `TestHTTPAPIDefaultsStayLoopbackNoTokenReadOnlyAndRequireTLS` |
| B4 | 인증서 로드 실패 | 없음 | 원인 그대로 | 미고정 |
| B5 | 리소스 열기 실패 | 없음 | 원인 그대로 | 미고정 |
| B6 | 엔진 디렉터리를 해석했다 | 없음 | 없음 | `TestTheDaemonReadsTheDescriptorBesideItsOwnJournal` |
| B7 | descriptor 가 stat 된다 | 없음 | 없음 | `TestADeadDescriptorDoesNotStopTheDaemon` |
| B8 | (B7 의 else 절 헤드) | 없음 | 없음 | `TestAnAbsentDescriptorAndADeadOneBootTheSame` |
| B9 | **dial 실패 → 강등** | **stderr 경고**, `strategyRuntime` 은 nil 로 남는다 | **없음** | `TestADeadDescriptorDoesNotStopTheDaemon` |
| B10 | dial 성공 | `strategyRuntime = client` | 없음 | 미고정(실 engine 필요 — 겹1 T1 범위) |
| B11 | stat 오류가 NotExist 가 아니다 | **stderr 경고**, `strategyRuntime` 은 nil 로 남는다 | **없음 — 강등 (6.8① 이 뒤집었다)** | `TestAnUninspectableDescriptorDegradesLikeTheConsole` |
| B12 | 스냅샷 캐시 실패 | 없음 | 원인 그대로 | 미고정 |
| B13 | 스트림 생성 실패 | 없음 | 원인 그대로 | 미고정 |
| B14 | mutation 라우트 구성 실패 | 없음 | 원인 그대로 | `TestHTTPAPIMutationsStayAbsentUntilBothSecurityArtifactsAreConfigured` |
| B15 | 라우터 생성 실패 | 없음 | 원인 그대로 | 미고정 |
| B16 | 서버 생성 실패 | 없음 | 원인 그대로 | 미고정 |
| B17 | 직접 TLS 면 TLSConfig 설정 | 서버 상태 | 없음 | 미고정 |
| B18 | 직접 TLS 면 ListenAndServeTLS | listener | 없음 | 미고정 |
| B19 | serve 종료 vs ctx 종료 | 없음 | 분기 | `TestADeadDescriptorDoesNotStopTheDaemon` (ctx 만료 경로) |
| B20 | `http.ErrServerClosed` | 없음 | nil | 미고정 |
| B21 | Shutdown 실패 | 없음 | 원인 그대로 | 미고정 |
| B22 | serve 오류가 ErrServerClosed 아님 | 없음 | 원인 그대로 | 미고정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `httpAPIBoundary` | 사설 경계·TLS 모드 확정 | 실패는 거절 | AST · httpapi_test.go |
| `openHTTPAPIResources` | 원장·성과·최적화·reader | `reader.validate()` 가 최종 관문 | AST |
| `engineJournalDir` | 자기 프로파일의 엔진 디렉터리 | 실패면 전략 화면 없이 진행 | AST · engine.go:172 |
| `strategyprojectionrpc.DescriptorPath` | descriptor 경로 | 순수 함수 | AST |
| `os.Stat(descriptor)` | 부재/조사불가 구분 | **NotExist 만 정상 부재** | AST |
| **`strategyprojectionrpc.Dial`** | 조회 클라이언트 | **실패는 강등 (a108 D4-2)**. S3(주인 없는 socket 파일)에서 실패로 오는 것은 T1-fix 의 connect probe 몫이다 — 그때까지 `TestASocketFileWithNoOwnerDegradesTheDaemon` 이 RED | AST · a108 T2 테스트 |
| `httpapi.NewRouter` | 라우팅 | `StrategyRuntime` 은 nil 을 받는다 | AST |
| `server.ListenAndServe(TLS)` | 서비스 | `ErrServerClosed` 는 정상 종료 | AST |

## State mutations and fallbacks

- 이 함수는 **디스크를 쓰지 않는다.** 특히 잔재 descriptor 를 **치우지 않는다** — 잔재
  정리는 엔진(겹1)의 일이고, 조회 전용 데몬이 남의 control 디렉터리를 치우기 시작하면
  그것이 새 사고 유형이 된다. `TestAnAbsentDescriptorAndADeadOneBootTheSame` 가 고정한다.
- 강등의 결과 상태는 `strategyRuntime == nil` 하나이며 `httpapi.NewRouter` 와
  `resources.reader.strategyRuntime` 두 곳에 같은 nil 이 들어간다 — descriptor 부재
  경로와 **문자 그대로 같은 상태**다. 6.8① 이후에는 조사 불가 경로도 여기로 합류한다:
  세 입력(부재·dial 실패·stat 오류)이 한 상태로 모이고, 구별은 stderr 문구가 진다.
- 그 nil 이 화면에서 무엇이 되는지는 `httpAPIReader.Snapshot` 의 FLM 이 진다
  (`cmd-tossctl--httpapireader.snapshot/`): nil 은 dormant(NOT_CONFIGURED)이고,
  붙었다가 죽은 reader 는 unavailable(RUNTIME_UNAVAILABLE)이다.
- 재-dial 은 없다. descriptor 부재로 뜬 데몬도 지금 그렇게 남으며, 대칭을 깨는 쪽이
  새 비대칭을 만든다(design D4, 선언된 생략).

## Safety conclusion

- Safe edit boundary: B9·B10·**B11**. B6·B7·B8 의 판정과 순서는 그대로다.
  B11 의 판정 반전은 「환경 이상을 삼킨다」가 아니라 「환경 이상을 **말하되 죽지는
  않는다**」이며, 이 데몬이 조회 전용(`mutating` 주석 없음)이라 성립한다.
- High-risk impact: **no** (이 함수 자체는). `tossctl httpapi` 는 조회 전용이고
  `mutating` 주석이 없다. 다만 사고 당시 이 데몬의 crash loop 가 운영자의 유일한 화면을
  없앴으므로, 가용성 관점에서 a108 의 세 겹 중 하나로 다룬다.
