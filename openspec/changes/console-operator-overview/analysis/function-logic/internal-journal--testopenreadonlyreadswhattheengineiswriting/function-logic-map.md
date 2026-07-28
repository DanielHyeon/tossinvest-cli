# Function Logic Map: `TestOpenReadOnlyReadsWhatTheEngineIsWriting`

- Source: `internal/journal/readonly_test.go`
- AST evidence: `ast.json` (revision `base` — base 쪽 hunk에만 걸린다)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 테스트 — **본문 무변경**이다. 이 change는 이 함수 **뒤**(base L249 이후)에
`insertAttemptWithBrokerOrder`와 새 테스트 둘을 삽입했고, `@@ -246,3 +250,83 @@` hunk가
함수 끝과 인접해 evidence가 요구됐다. base(`137cc8d`) L222-248과 HEAD L226-252는
**바이트 동일**(함수 구간 sha256 `87715e37d816eb93…` 일치, 본 세션 확인)이며,
줄 번호 4칸 이동은 앞쪽 allowlist 4줄 삽입 때문이다.

이 테스트가 지키는 사실: 콘솔은 **돌고 있는 엔진의** journal을 연다. WAL이 그것을 가능하게
하고, read-only 커넥션이 `-shm`/`-wal` 쌍에 접근하는 것이 "아무것도 만들지 않는다"의
문서화된 예외다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 쓰기 handle | `openTestJournalAt` | 테스트 헬퍼 | `t.Fatalf` |
| 읽기 handle | `openTestReadOnly` (`mode=ro`) | 동상 | 동상 |
| 관측 대상 | `LivePositionExits(ctx, "acct-1")` | `account_views.go`(무변경) | 행 수·id 단언 |

불변식: read-only handle은 **startup 스냅샷에 고정되지 않는다** — handle을 연 뒤에 쓰인 행이
다음 질의에 보인다. 콘솔이 요청마다 다시 읽기 때문이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, base L230) | 첫 `LivePositionExits`가 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B2 (if, base L233) | 행이 1개가 아니거나 id 불일치 | 없음 | `t.Fatalf` — writer의 행을 못 봤다 | 자체 실행 |
| B3 (if, base L242) | 두 번째 `LivePositionExits`가 오류 | 없음 | `t.Fatalf` | 자체 실행 |
| B4 (if, base L245) | 행이 2개가 아님 | 없음 | `t.Fatalf` — 나중 쓰기가 보이지 않는다 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournalAt` / `openTestReadOnly` | 쓰기·읽기 두 handle | 실패 시 `t.Fatalf` | ast.json calls |
| `insertDecision` / `insertPosition` / `insertPosition2` | writer 쪽 행 생성 | 동상 | ast.json calls |
| `ro.LivePositionExits` | read-only 투영 | 오류 그대로 단언 | ast.json calls |

## State mutations and fallbacks

- 테스트 전용 `t.TempDir()` 안의 journal만 만든다. 실계좌·네트워크 무접촉.

## Safety conclusion

- Safe edit boundary: **본문 0줄 변경**. 인접 삽입만.
- High-risk impact: **yes** — 원장 read-only 계약(WAL 동시 읽기·스냅샷 비고정)을 재는
  가드다. 다만 이번 change의 편집은 이 함수에 0줄이고, 같은 파일에 새 테스트 2건과 헬퍼
  1건이 뒤에 붙었을 뿐이다. `readOnlyTables` 확대의 영향은 이 테스트에도 적용되지만
  (`openTestJournalAt`은 최신 스키마로 열므로 `mutation_attempts`가 있다) 통과 조건은 그대로다.
