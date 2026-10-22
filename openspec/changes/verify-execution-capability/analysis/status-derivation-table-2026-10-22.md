# 주문 status 파생 표 — task 2.1 (2026-10-22)

> 근거 순서: 코드(`internal/brokerstate/derive.go`) → AST(`analysis/ast/*.ast.json`, 열거
> `analysis/ast/enumerate-output.md`, `python3 -I enumerate.py` 재생성) → 이 문서. 분기·반환 줄 번호는 AST 열거에서
> 옮겼고 손으로 고르지 않았다. AST `source_sha256` 은 작성 시점 워킹트리와 일치(열거 출력 `STALE 0`).
> 관측 열은 이 change 의 `measurements.md` 만 인용한다. 이 문서는 실계좌 능력·자동 주문 권한을 승급하지 않는다.

## 1. 결론

- 상태 매핑 함수는 `brokerstate.Derive` 하나다(생산 호출자: `internal/filldetect/detect.go`
  `viewOf` 경유 2곳, `internal/console/orders.go:536` 표시용). 체결 감지 경로는
  `Derive` → `JournalLedger.Apply`(`FailClosed` 전달) → `Detector.pollLocked` B7(`detect.go:324`
  `if applied.FailClosed`) → `blockSymbol` → `Gate.Block(ReasonBrokerStateUnknown)` 이고, 생산 배선은
  `cmd/tossctl/engine.go:780` `Gate: ectx.Entry` 다.
- **모르는 값은 fail-closed 다.** `Derive` 의 `switch status` 마지막 갈래 B50(`derive.go:506` `default: // 37`)
  → 반환 `derive.go:507` `d.unknown(ReasonUnknownStatus, …)`. `unknown` 은 `derive.go:591`
  `d.FailClosed = true` 를 세운다. 빈 status 는 `derive.go:385` `ReasonMissingStatus`.
- **실계좌 미관측 두 값도 fail-closed 다.** `CANCEL_REJECTED` → `derive.go:493`
  `ReasonCancelRejectedRecord`, `REPLACE_REJECTED` → `derive.go:499` `ReasonReplaceRejectedRecord`.
  사용자 결정 2026-10-22 로 이 둘은 「실제 관측 없음」으로 수용했다. 별도 레코드의 형태·원주문 링크는
  측정되지 않은 채이며, 그 대가가 이 막힘이다(관측 전까지 자동 귀속하지 않는다).
- `Derive` 반환 35개 중 정상 판정(`settle`)은 8개(399·417·438·451·453·465·467·484)와
  위임 3개(`pendingMutation` 421·424, `deriveGroupLabel` 504)뿐이고 나머지 24개는 전부 `unknown`(fail-closed)이다.

## 2. 파생 표 (문서화된 10개 + 레거시 그룹 라벨 + 그 밖의 값)

`Derive` 의 1–4행 구조 검증(비유한 수량 356 · 수량 ≤0 359 · 음수 체결 362 · 체결>수량 365)이
status 보다 먼저 걸리고, 그것도 전부 `unknown` 이다.

| raw status | 정상 결과(`settle` 반환 줄) | fail-closed 결과(반환 줄 · reason) | 실계좌 관측 |
|---|---|---|---|
| `""`(없음) | — | 385 `missing_status` | — |
| `PENDING` | 399 `OPEN_UNFILLED` | 390 `open_with_cancel_timestamp` · 393 `open_with_successor` · 396 `pending_with_fill` | 관측(M25 US 접수 후, M32 KR 접수·취소 직후). M31 목록 walk 에선 미관측 |
| `PARTIAL_FILLED` | 417 `OPEN_PARTIALLY_FILLED` | 405 · 408 · 411 `partial_filled_without_fill` · 414 `partial_filled_with_full_fill` | **미관측** |
| `PENDING_CANCEL` | `pendingMutation` 527 → `CANCEL_PENDING` | 518 · 521 · 524 `pending_mutation_with_full_fill` | 관측(M25 US `after_cancel=PENDING_CANCEL`) |
| `PENDING_REPLACE` | `pendingMutation` 527 → `REPLACE_PENDING` | 518 · 521 · 524 | 관측(M28 US `original_status=PENDING_REPLACE`) |
| `FILLED` | 438 `FILLED` | 429 `filled_with_remainder` · 432 `replaced_with_full_fill` · 435 `cancelled_with_full_fill` | 관측(M31 65건) |
| `CANCELED` | 451 `CANCELLED` · 453 `CANCELLED_PARTIALLY_FILLED` | 444 `cancelled_with_successor` · 448 `cancelled_with_full_fill` | 관측(M31 119건) |
| `REJECTED` | 465 `REJECTED` · 467 `REJECTED_PARTIALLY_FILLED` | 459 `rejected_with_successor` · 462 `rejected_with_full_fill` | 관측(M31 13건) |
| `REPLACED` | 484 `REPLACED`(후속 주문 id 있을 때만) | 473 `replaced_with_full_fill` · 480 `replaced_without_successor` | **미관측** |
| `CANCEL_REJECTED` | — | 493 `cancel_rejected_record` | **미관측**(M31 `listed=false`) — 2026-10-22 「실제 관측 없음」 수용 |
| `REPLACE_REJECTED` | — | 499 `replace_rejected_record` | **미관측**(M31 `listed=false`) — 2026-10-22 「실제 관측 없음」 수용 |
| `OPEN`/`CLOSED`(요청 그룹 라벨, status 아님) | `deriveGroupLabel` 547 · 549 · 557 · 562 · 564 · 566 | 541 · 544 · 554 · 559 · 568 `closed_without_fill_or_cancel` · 571 `closed_with_unexplained_remainder` | 레거시 호환(fixture·옛 저널 행) |
| 그 밖의 값(예: `EXPIRED`, `PARTIALLY_FILLED`, `체결`) | — | 507 `unknown_status` | — |

대소문자·앞뒤 공백은 `strings.ToUpper(strings.TrimSpace(...))` 로 정규화한 뒤 비교한다(`derive.go:382`).

## 3. 고정하는 시험

| 시험 | 고정하는 것 |
|---|---|
| `internal/brokerstate/derive_status10_test.go` `TestDeriveOrderStatusTable` | 10개 값 전 행, `CANCEL_REJECTED`/`REPLACE_REJECTED` → `UNKNOWN`, `EXPIRED` → `unknown_status`, `UNKNOWN` 이면 `FailClosed=true`·`Terminal=false` |
| `internal/brokerstate/derive_test.go` `TestDerivePriorityTable` | 빈 값 · 미문서 값(`PARTIALLY_FILLED`) · WTS 한글 값(`체결`) → `UNKNOWN` |
| `internal/filldetect/unknown_status_test.go` `TestUnobservedStatusFailsClosedAtTheEntryGate` (2026-10-22 추가) | `EXPIRED`·`CANCEL_REJECTED`·`REPLACE_REJECTED` 가 폴링 1회에 진입 게이트를 `ReasonBrokerStateUnknown` 으로 막고 체결 이벤트 0건 |

추가 시험의 반증(2026-10-22, HEAD `4cd7af19d` 의 분리 worktree, 변이 후 원복 확인):
`derive.go:507` 을 `settle(StateCancelledPartiallyFilled)` 로 바꾸면 `undocumented_status` 가 FAIL,
`derive.go:493` 앞에 `return d.settle(StateCancelled)` 을 넣으면 `cancel_rejected_record` 가 FAIL.

## 4. 같은 status 문자열을 읽는 다른 소비자 (판정 경로 아님)

- `internal/console/orders.go:536` — 같은 `Derive` 를 표시용으로 부르고 `FailClosed` 를 `Unresolved` 로 노출한다.
- `internal/manualbuy/summary.go:116` `terminalStatuses` — 수동 매수 금액 요약. `CANCEL_REJECTED`·`REPLACE_REJECTED` 를
  종결로 세지만, 그 밖의 모르는 값은 비종결로 떨어져 주문 금액 전체를 센다(과대 집계 방향). 주문·진입 판정은 하지 않는다.
- 완료 이력(WTS `orders completed`)의 한글 status(`체결완료`·`취소`·`실패`)는 이 enum 이 아니다(M46). `Derive` 에 넣으면 `unknown_status` 다.

## 5. 남은 미관측(이 문서가 닫지 않는 것)

`PARTIAL_FILLED`·`REPLACED`·`CANCEL_REJECTED`·`REPLACE_REJECTED` 의 실제 payload 형태. 앞 둘은 정상 판정 행이
있으나 실계좌 fixture 가 없다. 뒤 둘은 관측될 때까지 막힘이 정답이고, 관측되면 이 표와 2a 귀속 규칙을 다시 본다.
