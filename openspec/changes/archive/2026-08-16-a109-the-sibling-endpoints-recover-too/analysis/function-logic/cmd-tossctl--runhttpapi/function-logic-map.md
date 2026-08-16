# Function Logic Map: `runHTTPAPI`

- Source: `cmd/tossctl/httpapi.go` (121-217)
- AST evidence: `ast.json` — AST 분기 16 · return 14 · defer 6
  (source_sha256 는 ast.json 이 정본, a109 §2-fix F3 편집 후 생성)
- Risk scan: `risk-pattern-report.md`
- **a109 §2-fix 편집 대상: 인자 하나**. publisher goroutine 이 `publishHTTPAPISnapshots`
  에 `strategyRuntime` 을 넘긴다(:182). 분기·defer·수명 구조는 **하나도 바뀌지 않는다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 가능 | cobra | nil 이면 `context.Background()` (B1) |
| 경계·TLS 옵션 | 유효한 boundary | `httpAPIBoundary` | 실패는 fatal (B2·B4) |
| 조회 자원 | 열 수 있어야 한다 | `openHTTPAPIResources` | 실패는 fatal (B5), 성공하면 defer Close |
| 전략 reader | **nil·sentinel·live 전부 유효** | `strategyRuntimeReaderFor` | 부재·실패는 화면 값으로만 나타난다 (a108 D4-2·a109 D4) |

**불변식 1**: 이 데몬은 **조회 전용**이다. 전략 endpoint 하나 때문에 죽는 경로는 없다
(a108 D4-2). `strategyRuntimeReaderFor` 는 언제나 값을 주고 여기서 fatal 을 만들지 않는다.

**불변식 2**: publisher goroutine 의 수명은 `publisherCtx` 다. `defer` 두 개(:184 의
cancel + `<-publisherDone`)가 서버 종료보다 **나중에** 등록돼 먼저 돌고, 그래서 루프가
멈춘 뒤에 stream 이 닫힌다(:158).

**불변식 3 (a109 §2-fix F3)**: publisher 는 재부착의 상시 구동원이므로 그 goroutine 이
전략 reader 를 **알아야 한다**. 인자로 넘기는 이유는 `httpAPISnapshotCache` 에 source 를
되돌려주는 새 결합을 만들지 않기 위해서다(issues.md T2-12).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:123) | ctx 가 nil | 기본 ctx 사용 | — | 기존 httpapi 명령 테스트 |
| B2 (:130) | boundary 해석 실패 | 없음 | 오류 (:131) | 기존 |
| B3 (:134) | 직접 TLS | 인증서 로드 | — | 기존 |
| B4 (:138) | 인증서 로드 실패 | 없음 | 오류 (:139) | 기존 |
| B5 (:143) | 자원 열기 실패 | 없음 | 오류 (:144) | 기존 |
| B6 (:151) | 스냅샷 캐시 생성 실패 | 없음 | 오류 (:152) | 기존 |
| B7 (:155) | stream 생성 실패 | 없음 | 오류 (:156) | 기존 |
| B8 (:160) | mutation route 조립 실패 | 없음 | 오류 (:161) | 기존 |
| B9 (:166) | router 생성 실패 | 없음 | 오류 (:167) | 기존 |
| B10 (:171) | server 생성 실패 | 없음 | 오류 (:172) | 기존 |
| B11 (:174) | 직접 TLS | `server.TLSConfig` 설정 | — | 기존 |
| B12 (:191) | 직접 TLS (serve goroutine) | ListenAndServeTLS | 채널로 오류 | 기존 |
| B13 (:199) | `select` — serve 오류 vs ctx 종료 | 없음 | 아래 둘 | 기존 |
| B14 (:201) | `http.ErrServerClosed` | 없음 | nil (:202) | 기존 |
| B15 (:208) | Shutdown 실패 | 없음 | 오류 (:209) | 기존 |
| B16 (:212) | 종료 후 serve 오류 | 없음 | 오류 (:213) | 기존 |

defer 6: `stop`(:127) · `resources.Close`(:146) · `stream.Close`(:158) ·
`close(publisherDone)`(:181, goroutine 자기 스택) · publisher 정지(:184) ·
`cancel`(:207, ctx.Done 가지 안).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyRuntimeReaderFor` | 전략 reader 자리를 만든다(재부착 wrapper) | **오류를 내지 않는다** — 부재·실패는 값이다 | AST · :147 |
| `newHTTPAPISnapshotCache` | 집계 캐시 | 실패는 fatal (B6) | AST · :150 |
| `httpapi.NewStream` / `NewRouter` / `NewServer` | 표면 조립 | 실패는 fatal (B7·B9·B10) | AST · :154·163·170 |
| `publishHTTPAPISnapshots` | 발행 루프 **+ 재부착 상시 구동원** | 반환은 ctx 종료·Publish 실패뿐 | AST · :182 |

## State mutations and fallbacks

- 프로세스 밖 상태를 바꾸지 않는다 — 조회 데몬이다.
- fallback 없음: 조립 실패는 전부 fatal 이고, **전략 endpoint 만** 예외다(값으로 흡수).

## Safety conclusion

- Safe edit boundary: publisher goroutine 의 **인자 목록** 한 줄.
- High-risk impact: **no** — 주문·손절 경로가 아니다. 다만 defer 순서는 stream·publisher
  수명 계약이므로 건드리지 않는다.
- 금지: 전략 endpoint 실패를 fatal 로 되돌리는 것(a108 D4-2), publisher 정지 defer(:184)를
  stream.Close(:158)보다 **먼저** 등록하는 것(닫힌 stream 에 발행하게 된다).
