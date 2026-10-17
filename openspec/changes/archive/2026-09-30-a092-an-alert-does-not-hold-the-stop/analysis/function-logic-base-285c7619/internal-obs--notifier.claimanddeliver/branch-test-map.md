# Branch Test Map: `Notifier.claimAndDeliver`

Source: `internal/obs/notifier.go` (238-277). AST 기준 분기 4 / 이탈 3 /
defers 1 / go_statements 0.

**a092가 편집하는 함수**이므로 RED 열이 실제 의무를 담는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:245` **claim 실패 → 진입 차단** | `internal/obs/a097_claim_failure_blocks_entry_test.go TestAClaimThatFailsBlocksNewEntries:61` | yes (a097) | yes |
| B2 | `:258` `n.Log != nil` → 오류 줄이 남는다 | 같은 파일의 claim 실패 테스트가 로그를 검사한다 | yes (a097) | yes |
| B3 | `:261` `n.Gate != nil` → 래치 | `TestAClaimThatFailsBlocksNewEntries:61` | yes (a097) | yes |
| B4 | `:266` **`!owed` → 발송 억제, 기록은 남는다** | `a096_one_send_per_condition_test.go TestSuppressingTheSendKeepsTheRecord:302`, `TestTheSameConditionIsRemindedOncePerWindow:93` | yes (a096) | yes |
| — `:276` | owed → `deliver` 호출 | `TestOneConditionIsOneSend:67`, `TestConcurrentObservationsOfOneConditionSendOnce:207` | yes (a096) | yes |

**a097의 주석이 B1의 이력을 적어 뒀다**(`a097_claim_failure_blocks_entry_test.go:13-15`):
*"claimAndDeliver's error branch (B1@227) has count=0. Nothing in this repository
has ever executed it."* — 그 측정이 a097의 RED를 만들었고 지금은 GREEN이다.
행 번호가 227에서 245로 옮긴 것은 a097이 그 사이에 코드를 더했기 때문이다.

## a092가 이 함수에 대해 지는 RED

| RED | 무엇을 관측하나 | 이 표의 어느 행 |
|---|---|---|
| **R17-1** | 관측 사이클이 전송 응답을 기다리지 않는다 | `:276` 행이 **동기 호출이 아니게** 된다 |
| **R17-2** | 이 함수가 배달 잠금을 잡지 않는다 | `:241` 잠금 범위 |
| **R17-3** | claim × 배달 루프 인터리빙에서 이중 발송이 없다 | `:276`과 배달 루프 사이 |
| **R17-9** | outbox 기록 실패는 **여전히 동기**로 래치한다 | **B1 — 회귀 방지** |

**R17-9는 지키는 테스트다.** B1의 현재 동작이 17판에서도 유지되는지 보는 것이며,
`TestAClaimThatFailsBlocksNewEntries`가 이미 그것을 관측하고 있으므로
**17판 구현 뒤에도 이 테스트가 초록이어야 한다.**

`TestConcurrentObservationsOfOneConditionSendOnce:207`은 현재 `n.mu`로 성립한다.
17판이 잠금을 좁힌 뒤에도 이 테스트가 초록이어야 하고, 그때의 근거는 뮤텍스가
아니라 CAS다. **테스트를 고치지 않고 통과해야 성질이 보존된 것이다** —
고쳐서 통과시키면 a096 라운드 1로 되돌아간 것이다. §8 GREEN의 조건으로 남긴다.
