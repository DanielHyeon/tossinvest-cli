# Function Logic Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :286–359, 분기 7 · 반환 5 · 호출 18, source_sha256 `fbdfd9e0218b…`, 추출 커밋 `55963f29`(25라운드 수리 뒤 재추출). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존. 26라운드 수리 두 로트(`d8769cfb` · `b910173a`) 뒤 재추출 — 이 함수 본문 · 분기 수 불변(같은 파일의 다른 함수 편집으로 줄 이동 · 파일 해시만 바뀜, 분기 좌표는 `ast.json` 이 정본).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ③ — `fbc6df5f`): `defer n.mu.Unlock()` 을 없애고 각 반환 앞 · `deliver` 앞에서 명시적으로 푼다 — 잠금은 claim 과 그 판정(B1~B6)만 덮는다. `deliver` 의 판정을 넷째 반환값으로 올린다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.mu` | claim ~ 판정만 | 알림기 | 구조 핀 `a096PinLockedBetween(claimAndDeliver, ClaimAlertForDelivery)`(변이 L15 · L18) |

## 25라운드 수리 (`55963f29`)

- 반납 결과 분기에 새 갈래: `SettleAlreadySettled` · `SettleLeaseLost` 만 선점(lost), 그 밖(`SettleNotFound` · 모르는 결과)은 근거 확정 뒤 세대를 읽고 `BlockUnlessClearedSince`(승격 없음 — a124 N6). codex r25 P0.
- 아래 분기 표의 줄 번호는 단위 ③ 착지(`fbc6df5f`) 기준이다 — 현재 좌표는 `branch-test-map.md`(AST 기준).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:295) | claim 실패 → 잠금 안 무조건 래치 · 해제 후 반환 | — | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` |
| B2 | if (:308) | 로그 | — | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` |
| B3 | if (:311) | 래치 | — | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` |
| B4 | switch (:318) | claim 결과 분기 | — | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092ARecordDoesNotWaitForAnotherSendersTransport` |
| B5 | case (:319) | 이미 정착 → 해제 후 반환 | — | `TestNotifierIsConcurrencySafe`, `TestOneConditionIsOneSend` |
| B6 | case (:329) | 남의 임차 → 해제 후 `logClaimHeld`(INFO) | — | `TestAHeldRowIsNotWhispered`, `TestALeaseLineCarriesItsOwnName` |
| B7 | if (:348) | lost → 판정 없음 | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| 종단 | — | 잠금 해제 → `logClaimStolen` · `deliver`(잠금 밖) | `sent,true,verdict,nil` | `TestA092ARecordDoesNotWaitForAnotherSendersTransport` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ClaimAlertForDelivery` | 기록 + 임차(잠금 안) | 오류는 B1 | AST |
| `n.deliver` | 전송 · 정산(잠금 밖) | (sent, lost, verdict) | AST |

## State mutations and fallbacks

- 잠금 안: claim(원장 쓰기). 잠금 밖: 전송 · 정산. B1 래치는 잠금 안이라 승인의 셈~해제와 겹치지 않는다.

## Safety conclusion

- Safe edit boundary: 분기 B1~B7 의미 불변, 잠금 범위만.
- High-risk impact: yes — 잠금 범위(변이 L01 이 행동 시험 셋에서 잡힘).
