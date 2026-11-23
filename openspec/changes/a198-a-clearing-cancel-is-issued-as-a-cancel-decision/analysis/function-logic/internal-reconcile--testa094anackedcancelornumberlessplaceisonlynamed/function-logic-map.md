# Function Logic Map: `TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed`

- Source: `internal/reconcile/a094_acked_boot_test.go:182-215`(post-edit)
- AST evidence: `ast.json` (**post-edit** — source_sha256 `26cff774…`, 분기 5). Pre-edit 판은 HEAD `a7639378e`(sha `3b783a36…`, 분기 3 — 아래 표의 번호).
- **재번호(pre → post)**: B1·B2 동일 · **새 B3 `tc.want == CONFIRMED` · 새 B4 알림·번호 단언** · 옛 B3 → B5.
- Risk scan: `risk-pattern-report.md` (적중 0)
- 편집 범위(design §4-29, D9 로 뒤집힘): 표 행마다 기대(상태·읽기·알림)를 싣는다 — 취소 행은 CONFIRMED·읽기 0·알림 0, 정정(AMEND) 행과 번호 없는 PLACE 행은 ACKED·읽기 0·명명 알림 1(종전 단언 유지).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 표 행 | 오늘 `{"cancel", KindCancel, "O-target"}` 하나 | 시험 | — |
| `a094CrashAfterAck` | Prepare → MarkDispatchStarted → MarkAcked(번호) 뒤 닫음 | 시험 픽스처 | Fatal |
| `a094Recover` | 실 `Recovery.Run`, 읽기 계수 가짜 | 시험 픽스처 | Run 오류면 Fatal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L182 | 표 순회 | 하위 시험 | — | 자기 자신 |
| B2 L196 | 상태 ≠ ACKED 또는 읽기 ≠ 0 | — | Fatal | 자기 자신(D9 뒤 취소 행이 여기서 실패 — 뒤집힘) |
| B3 L200 | 알림 종목 ≠ AAPL | — | Error | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a094CrashAfterAck` · `a094Recover` · `LookupAttempt` · `a094AssertNamed` | 픽스처·기동·판독·알림 단언 | Fatal | AST |

## State mutations and fallbacks

- 임시 원장 파일만. 실 브로커 0.

## Safety conclusion

- Edit boundary: 표 행에 기대 상태를 더하고 취소 행만 CONFIRMED·알림 0 으로 뒤집음(D9). AMEND·번호 없는 PLACE 행은 종전 단언(ACKED·명명 알림 1·종목) 그대로 — 시험 이름이 말하는 범위를 실제로 채움.
