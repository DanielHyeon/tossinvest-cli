# Function Logic Map: `buildReconcileResolveRuntime` (편집 뒤 판)

- Source: `cmd/tossctl/engine_reconcile.go` (239-314), 작업 트리 source sha256 `d5b36622a940…`. 편집 전: `../../pre-edit-ast/cmd-tossctl--buildreconcileresolveruntime.ast.json`(237-319, sha `7192c69246ed`, 분기 14).
- AST evidence: `ast.json`(분기 12 — B12 는 validate 클로저 안; return 13: 242·246·250·253·258·262·266·276·282·297·301·307·309). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file cmd/tossctl/engine_reconcile.go --func buildReconcileResolveRuntime`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 커버리지: 편집 전 `../../coverage/pre-edit-buildReconcileResolveRuntime-blocks.txt`, 편집 뒤 `../../coverage/after-engine_reconcile-blocks.txt`(선택 시험 `-run 'TestA166EngineReconcile|TestEngineReconcileResolve|A195'`).
- 재번호: `renumber.txt`(`../renumber.py` — 옛/새 분기를 (종류, 그 줄의 소스 문장)으로 difflib 정렬. 손 재번호 안 함). B1~B12 는 번호 유지·줄 +2, 옛 B13 range@307·B14 if@308 은 삭제(새 함수 `reconcileResolveRefusedCause` 로 옮김).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 호출자 컨텍스트 | `runEngineReconcileResolve` :139 | 취소면 클라이언트 생성·계좌 조회 오류(B5·B6) |
| `root *rootOptions` | B8 은 nil 을 검사하지만 :254 `root.officialOptions` 가 먼저 역참조한다(nil 이면 panic) | cobra 루트 | B1~B7 오류 반환 |
| 공식 자격증명 | API/Secret 키 비어 있지 않음 | `official.LoadCredentials` | B3·B4 거절 |
| 계좌 신원 | 첫 계좌의 번호·순번 | `reconcileResolveAccountRef` | B7 거절 |
| 장부 | `root.configDir/journal.DBFileName`(빈 경로면 journal 기본) | `journal.Open` | B9 거절 |
| 불변식 1 | 반환 런타임에 주문 쓰기 핸들 없음(읽기 `engine.OfficialReads` 만) | 파일 머리 주석 · `TestEngineReconcileResolveProductionWiringHasNoBrokerMutationOrHTTPDoor` | — |
| 불변식 2 | 해제 가능한 원인 판정은 `reconcileResolveRefusedCause`(reconcile_resolve_causes.go) 한 곳, 클로저는 그 결과를 그대로 돌려줌 | 작업 트리 | 결과를 무시하면 모든 원인이 `Tracker.Resolve` 로 감 — census 가 클로저 모양을 고정 |
| 불변식 3 | 함수 본문 전체가 a166 census 로 고정 — 이 change 가 그 키 하나를 갱신(`../../census-repair/shapes-key-diff.txt`) | shapes.json | 본문 한 글자 변경 → `TestA166WiringCensusBaseline` 실패(`reviewed function shape changed`) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 if@241 | `loadConfig` 오류 | 없음 | return@242 err 그대로 | 편집 무관 |
| B2 if@245 | `resolveOpenAPIPaths` 오류 | 없음 | return@246 | 편집 무관 |
| B3 if@249 | `LoadCredentials` 오류 | 없음 | return@250 `loading official credentials` 감쌈 | 편집 무관 |
| B4 if@252 | creds nil 또는 키 빈 값 | 없음 | return@253 `engine.ErrOfficialCredentialsRequired` | 편집 무관 |
| B5 if@257 | `newRootOfficialClient` 오류(예산 required 실패 포함) | 토큰 교환 시도 가능 | return@258 `opening official client` | 편집 무관 |
| B6 if@261 | `client.Accounts` 오류 | 브로커 읽기 1회 | return@262 `resolving the official account` | 편집 무관 |
| B7 if@265 | 계좌 신원 해석 실패 | 없음 | return@266 | 편집 무관 |
| B8 if@271 | `root != nil` 이고 `configDir` 비어 있지 않음 | `journalPath` 설정(분기 안 return 없음) | — | 편집 무관 |
| B9 if@275 | `journal.Open` 오류 | 없음 | return@276 `opening the engine journal` | 편집 무관 |
| B10 if@280 | `tracker.Restore` 오류 | `j.Close()` | return@282 err 그대로 | 편집 무관 |
| B11 if@285 | 설정 한도 통화가 빈 값 | `currency = engine.DefaultLimitCurrency` | — | 편집 무관 |
| (return@297) | 정상 | 런타임 조립(클로저 localState·validate, 메서드 값 collect·resolve·close) | 런타임, nil | — |
| (return@301, 클로저 localState) | — | `reconcile.LocalStateFromJournal` | 그 결과 | 편집 무관 |
| B12 if@306 (validate 클로저) | `tracker.Blocks()` 가 0 개 | 없음 | return@307 `no active RECONCILE block exists` | 동작 유지 — `TestA195NoActiveBlockIsRefused` |
| (옛 B13 range@307·B14 if@308 — 삭제) | 원인 순회·QM 외 거절 | — | — | `reconcileResolveRefusedCause` 로 옮김(D3) — 그 함수의 원인 표는 `TestA195RefusedCauseTable` |
| (return@309, 클로저) | 블록 1개 이상 | 없음 | `reconcileResolveRefusedCause(blocks)` 의 결과 그대로 — nil 이면 `runtime.resolve` 진행 | `TestA195OperatorReleasesASymbolAttributionFailure` 외 production 런타임 시험 6 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loadConfig`·`resolveOpenAPIPaths`·`official.LoadCredentials` | 설정·자격 | 오류 즉시 반환, 재시도 없음 | AST calls |
| `newRootOfficialClient(ctx, root, *creds, tokenFile, root.officialOptions...)` | 예산 결속 공식 클라이언트 | 오류 감쌈; 호출 모양은 `TestA166EngineReconcileOptionsGuard` 가 AST 로 고정 | AST · a166_reconcile_safety_test.go:25 |
| `client.Accounts` | 계좌 신원 | 브로커 읽기, 오류 감쌈 | AST |
| `journal.Open` · `tracker.Restore` | 장부·활성 차단 투영(`ActiveReconcileStates` → 설정 계좌만 → `blockFromReconcileState`) | Restore 실패면 장부 닫고 반환 | AST · `../../cited-ast/Tracker.Restore.ast.json` B4 if@979 |
| `tracker.Blocks`(클로저 안) | validate 시점의 활성 블록 | 오류 없음 | `../../cited-ast/Tracker.Blocks.ast.json` |
| `reconcileResolveRefusedCause`(클로저 안, 새 함수) | 허용 원인 표 {QM, 종목 범위 AF} 밖의 첫 블록을 이름으로 거절 | 오류 반환만, side effect 없음 | cmd/tossctl/reconcile_resolve_causes.go |
| `tracker.Resolve`(메서드 값) | 원장 원자 해제(`ReleaseReconciles`, 블록별 `ExpectCause`) | 한 행이라도 불일치면 전체 롤백 | `../../cited-ast/Tracker.Resolve.ast.json` |
| live config | `cfg.Engine.AutomationGate.LimitCurrency` 만 읽음(B11) | 쓰기 없음 | AST |

## State mutations and fallbacks

- 이 함수 자체의 쓰기: 없음(장부 열기·투영 복원만). 해제 쓰기는 반환한 `resolve` 를 `runEngineReconcileResolve` B23 if@206 가 부를 때만.
- fallback 둘: B8 거짓이면 journal 기본 경로, B11 참이면 기본 통화.
- 클로저(B12 + 새 함수 호출)는 `runEngineReconcileResolve` 가 잠금(B7)·스냅숏 3(B11~B19)·비교(B21 `diff.BlocksEntry()`) **뒤**에 부른다(B22 if@203).

## Safety conclusion

- Edit boundary(실제): 클로저의 원인 순회 8줄을 호출 1줄로 바꿈(`../../census-repair/shapes-key-diff.txt` 의 옛/새 값 diff 가 그 전부). B1~B12·return@297 의 배선과 `newRootOfficialClient` 호출 모양 무변경. census 키 하나 갱신(D6, Manager 승인).
- High-risk impact: yes — 대사 해제(안전 불변식 5). 허용 목록이 넓어지면 모르는 원인의 차단이 운영자 한 명령으로 풀린다.
