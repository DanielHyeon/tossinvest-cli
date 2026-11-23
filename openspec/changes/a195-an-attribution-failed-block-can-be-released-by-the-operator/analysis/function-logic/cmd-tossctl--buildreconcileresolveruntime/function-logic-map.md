# Function Logic Map: `buildReconcileResolveRuntime` (pre-edit 판)

- Source: `cmd/tossctl/engine_reconcile.go` (237-319), base `064f59fd9`, source sha256 `7192c69246ed…`.
- AST evidence: `ast.json`(분기 14 — 그중 B12~B14 는 validate 클로저 안; return 14). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file cmd/tossctl/engine_reconcile.go --func buildReconcileResolveRuntime`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 커버리지: `../../coverage/pre-edit-buildReconcileResolveRuntime-blocks.txt`(선택 시험 `./cmd/tossctl -run 'TestA166EngineReconcile|TestEngineReconcileResolve'`, covermode=set).
- 이 판은 **편집 전** 지도다. 편집 뒤(tasks 3.3)에 AST·지도를 다시 뽑고 옛/새 ast 를 정렬해 재번호한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 호출자 컨텍스트 | `runEngineReconcileResolve` :139 | 취소면 클라이언트 생성·계좌 조회 오류(B5·B6) |
| `root *rootOptions` | B8 은 nil 을 검사하지만 :254 `root.officialOptions` 가 먼저 역참조한다(nil 이면 panic) | cobra 루트 | B1~B7 오류 반환 |
| 공식 자격증명 | API/Secret 키 비어 있지 않음 | `official.LoadCredentials` | B3·B4 거절 |
| 계좌 신원 | 첫 계좌의 번호·순번 | `reconcileResolveAccountRef` | B7 거절 |
| 장부 | `root.configDir/journal.DBFileName`(빈 경로면 journal 기본) | `journal.Open` | B9 거절 |
| 불변식 1 | 반환 런타임에 주문 쓰기 핸들 없음(읽기 `engine.OfficialReads` 만) | 파일 머리 주석 · `TestEngineReconcileResolveProductionWiringHasNoBrokerMutationOrHTTPDoor` | — |
| 불변식 2 | 해제 가능한 원인 판정은 validate 클로저(B12~B14) 한 곳 | 현재 HEAD | 판정이 빠지면 모든 원인이 `Tracker.Resolve` 로 감 |
| 불변식 3 | 함수 본문 전체가 a166 census 로 고정 | `internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` | 본문 한 글자 변경 → `TestA166WiringCensusBaseline` 실패(`reviewed function shape changed`) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 if@239 | `loadConfig` 오류 | 없음 | return@240 err 그대로 | 편집 무관 |
| B2 if@243 | `resolveOpenAPIPaths` 오류 | 없음 | return@244 | 편집 무관 |
| B3 if@247 | `LoadCredentials` 오류 | 없음 | return@248 `loading official credentials` 감쌈 | 편집 무관 |
| B4 if@250 | creds nil 또는 키 빈 값 | 없음 | return@251 `engine.ErrOfficialCredentialsRequired` | 편집 무관 |
| B5 if@255 | `newRootOfficialClient` 오류(예산 required 실패 포함) | 토큰 교환 시도 가능 | return@256 `opening official client` | 편집 무관 |
| B6 if@259 | `client.Accounts` 오류 | 브로커 읽기 1회 | return@260 `resolving the official account` | 편집 무관 |
| B7 if@263 | 계좌 신원 해석 실패 | 없음 | return@264 | 편집 무관 |
| B8 if@269 | `root != nil` 이고 `configDir` 비어 있지 않음 | `journalPath` 설정(분기 안 return 없음) | — | 편집 무관 |
| B9 if@273 | `journal.Open` 오류 | 없음 | return@274 `opening the engine journal` | 편집 무관 |
| B10 if@278 | `tracker.Restore` 오류 | `j.Close()` | return@280 err 그대로 | 편집 무관 |
| B11 if@283 | 설정 한도 통화가 빈 값 | `currency = engine.DefaultLimitCurrency` | — | 편집 무관 |
| (return@295) | 정상 | 런타임 조립(클로저 localState·validate, 메서드 값 collect·resolve·close) | 런타임, nil | — |
| (return@299, 클로저 localState) | — | `reconcile.LocalStateFromJournal` | 그 결과 | 편집 무관 |
| B12 if@304 (validate 클로저) | `tracker.Blocks()` 가 0 개 | 없음 | return@305 `no active RECONCILE block exists` | 편집 인접 — 동작 유지(design §3 의 7) |
| B13 range@307 (validate 클로저) | 활성 블록 순회(`Blocks()` 는 정렬됨) | 없음 | — | **편집 대상** — 새 원인 판정 함수로 옮김(D3) |
| B14 if@308 (validate 클로저) | `block.Cause != QUANTITY_MISMATCH` | 없음 | return@309 `active cause %s on %s requires cause-specific recovery evidence`(첫 거절 블록 하나만 말함) | **편집 대상** — 허용 = {QM, 종목 범위 AF}(D1·D5) |
| (return@314, 클로저) | 모든 블록이 QM | 없음 | nil → `runtime.resolve` 진행 | 편집 대상 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loadConfig`·`resolveOpenAPIPaths`·`official.LoadCredentials` | 설정·자격 | 오류 즉시 반환, 재시도 없음 | AST calls |
| `newRootOfficialClient(ctx, root, *creds, tokenFile, root.officialOptions...)` | 예산 결속 공식 클라이언트 | 오류 감쌈; 호출 모양은 `TestA166EngineReconcileOptionsGuard` 가 AST 로 고정 | AST · a166_reconcile_safety_test.go:25 |
| `client.Accounts` | 계좌 신원 | 브로커 읽기, 오류 감쌈 | AST |
| `journal.Open` · `tracker.Restore` | 장부·활성 차단 투영(`ActiveReconcileStates` → 설정 계좌만 → `blockFromReconcileState`) | Restore 실패면 장부 닫고 반환 | AST · `../../cited-ast/Tracker.Restore.ast.json` B4 if@979 |
| `tracker.Blocks`(클로저 안) | validate 시점의 활성 블록 | 오류 없음 | `../../cited-ast/Tracker.Blocks.ast.json` |
| `tracker.Resolve`(메서드 값) | 원장 원자 해제(`ReleaseReconciles`, 블록별 `ExpectCause`) | 한 행이라도 불일치면 전체 롤백 | `../../cited-ast/Tracker.Resolve.ast.json` |
| live config | `cfg.Engine.AutomationGate.LimitCurrency` 만 읽음(B11) | 쓰기 없음 | AST |

## State mutations and fallbacks

- 이 함수 자체의 쓰기: 없음(장부 열기·투영 복원만). 해제 쓰기는 반환한 `resolve` 를 `runEngineReconcileResolve` B23 if@204 가 부를 때만.
- fallback 둘: B8 거짓이면 journal 기본 경로, B11 참이면 기본 통화.
- 클로저 B12~B14 는 `runEngineReconcileResolve` 가 잠금(B7)·스냅숏 3(B11~B19)·비교(B21 `diff.BlocksEntry()`) **뒤**에 부른다(B22 if@201).

## Safety conclusion

- Safe edit boundary: validate 클로저(302~315)의 원인 판정만. B1~B11·return@295 의 배선과 `newRootOfficialClient` 호출 모양은 바꾸지 않는다. 단 본문 어느 글자를 바꿔도 a166 census(shapes.json) 갱신이 같이 필요하다 — 편집 전 Manager 결정 대상(tasks 1.4).
- High-risk impact: yes — 대사 해제(안전 불변식 5). 허용 목록이 넓어지면 모르는 원인의 차단이 운영자 한 명령으로 풀린다.
