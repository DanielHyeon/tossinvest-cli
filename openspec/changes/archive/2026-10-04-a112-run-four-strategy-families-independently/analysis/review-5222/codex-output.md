BLOCK — 활성화 없는 입력의 마지막 권한 경계가 넓어졌고, 위험 적재 중 원장 오류가 범위 거절로 바뀔 수 있습니다.

코드·부모 커밋 비교에 따른 판정입니다. 사본 생성은 `mktemp: Read-only file system`으로 실패하여 시험·변이를 실행하지 못했습니다. 생성된 사본·삭제 대상 없음. `~/.codex`, 운영 원장, 다른 리뷰 파일은 읽지 않았습니다.

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | **P0** | **4-family 활성화 없는 두 범위 입력도 마지막 발급 권한 수집을 통과합니다.** 개수 관문 삭제가 활성화 여부로 제한되지 않았습니다. | 현재 [권한 수집:268–286](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_account_first_leg_authority.go:268). 새 [시험:193–199](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/a112_first_leg_owner_scope_seal_test.go:193)은 `activation`을 설정하지 않은 두 범위 쌍에서 winner의 수집 성공을 요구합니다. 부모 `7ab8cd12`의 같은 생산 파일:230–233은 이 입력을 거절합니다. | 비활성 시장에서는 기존 개수 관문을 유지하고, 활성 시장에서만 제거하십시오. OFF 입력의 이전 거절을 회귀 시험으로 고정하십시오. |
| 2 | **P1** | **J4 분류가 오류 발생 원점까지 이어지지 않습니다.** 위험 적재기의 원장 읽기 오류도 일반적인 범위 권한 부재로 변환됩니다. 다른 범위가 준비되면 주기가 계속될 수 있습니다. | 원장 오류 반환 [riskbucket:363–389](/tmp/claude-1000/a112-review-80ae96a5/internal/riskbucket/production_snapshot_authority.go:363) → `err`를 버리고 미준비 범위 저장 [위험 적재:205–220](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_risk_authority.go:205) → 다른 준비 범위가 시장을 Ready로 만듦 [:228–250](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_risk_authority.go:228) → `strategyScopeRefusal` 생성 [권한 수집:280–282](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_account_first_leg_authority.go:280) → 다음 범위 진행 [전달:69–72](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_market_handoff_delivery.go:69). | 적재 단계부터 예상된 범위 결손과 원장·무결성 오류를 구분하고 원인을 보존하십시오. 원장 오류가 하나라도 발생한 파도는 발급 전에 중단하는 종단 시험이 필요합니다. |
| 3 | **P2** | **“같은 파도 둘째 범위는 항상 `BUCKET_USAGE_STALE`”는 구현보다 강한 주장입니다.** 실제 검사는 버전 동일성이 아닌 사용량 하한 비교입니다. | [사용량 검사:46–57](/tmp/claude-1000/a112-review-80ae96a5/internal/journal/risk_bucket_usage.go:46)는 `snapshot < ledger`일 때만 거절합니다. 예약 버전은 발급 직전에 다시 읽습니다([수집:329–333](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_account_first_leg_authority.go:329)). 기존 사용량은 해제된 owner에 따라 감소할 수 있습니다([합산:516–524](/tmp/claude-1000/a112-review-80ae96a5/internal/riskbucket/production_snapshot_authority.go:516)). | 파도 순차 시험의 전제와 일반적인 공유 버킷 안전성을 구분하십시오. 기존 예약 해제가 끼는 경우도 검증하십시오. |

**#1의 범위:** 정상 상류에서는 비활성 다중 범위를 여전히 차단합니다([handoff:37–39](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_dispatch_handoff.go:37), [계좌 수집:160–162](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_account_first_leg_authority.go:160)). 따라서 현재 생산에서 실제 주문이 새로 나간다고 입증한 것은 아닙니다. 다만 마지막 권한 경계에서 **비활성 입력의 수락 집합이 확대**됐으므로, 요청하신 기준에 따라 P0입니다. 해당 시험도 원장 발급 완료가 아닌 발급 권한 객체 반환을 확인합니다.

**#2의 구체 경로:** 첫 범위의 위험 적재 중 일시적인 DB 읽기 실패 → 둘째 범위 적재 성공 → 시장 Ready → 첫 범위는 타입 거절로 건너뜀 → 둘째 범위 발급 가능. 코드상 도출이며 미실행입니다. 현재 J4 시험은 전달 함수에 이미 분류된 오류를 주입하므로, 이 **적재 오류 → 권한 부재** 변환을 검증하지 않습니다([시험:88–108](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/a112_scope_refusal_census_test.go:88)).

**#3의 반례 조건:** 공유 버킷 스냅숏 사용량이 100인 상태에서 기존 예약 100이 해제되고 첫 범위가 40을 예약하면, 둘째 범위는 `100 ≥ 40`이라 stale 검사를 통과할 수 있습니다. 다른 조건과 한도도 충족한다는 전제입니다. 이는 이중 소비 증명이 아니라 **항상 거절한다는 논증의 반례**입니다.

필수 항목별 판정:

- **① J4 — BLOCK.** 위조 identity 오류와 admission/Gateway 직접 오류의 중단 경로는 유지됩니다. 그러나 적재 오류의 범위 거절 오분류가 남습니다(#2). 반대 방향인 정상 범위 결손의 건너뛰기·반환 기록은 구현돼 있습니다.
- **② CAS — 조건부 적합.** 원장 트랜잭션의 사용량 재검사와 `usage + held + new` 합산은 유지됩니다([한도 검사:675–698](/tmp/claude-1000/a112-review-80ae96a5/internal/journal/reservations.go:675)). 이중 소비 경로는 확인하지 못했습니다. “둘째 항상 거절”은 성립하지 않습니다(#3).
- **③ 거울 다리 — 코드상 적합.** 실제 행만 복사하고 동일성을 비교하며([복사:150–165](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/a112_owner_scope_trading_test.go:150)), 둘째 파도의 위험 수집·발급 전에 실행됩니다([시험:316–320](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/a112_owner_scope_trading_test.go:316)). admission은 실제 journal을 계속 사용합니다. 승인된 스키마 핀 우회 외에 원장 거절을 다리가 발급으로 바꾸는 경로는 찾지 못했습니다.
- **④ M20 — 생산 구성에 한해 구조 동등 인정.** 범위마다 같은 고정 manifest digest를 전달하고([위험 적재:209–213](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_risk_authority.go:209)), 파일 digest를 검증한 뒤 그 manifest의 `Generation`을 사용합니다([생산 적재:153–169](/tmp/claude-1000/a112-review-80ae96a5/internal/riskbucket/production_snapshot_authority.go:153)). 파일 교체는 다른 세대의 정상 수락 대신 digest 불일치로 거절됩니다.
- **⑤ OFF = upstream — BLOCK.** 정상 상류 제한은 유지되지만 마지막 권한 경계의 비활성 입력 수락이 확대됐습니다(#1).
- **⑥ worker 승격 — 주문 검사 대체 없음.** “하나라도 통과”는 승격 판단입니다. dispatch가 범위마다 보호·진입 관문을 다시 검사하고([dispatch:96](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_dispatch_cycle.go:96), [:143](/tmp/claude-1000/a112-review-80ae96a5/internal/app/engine/strategy_dispatch_cycle.go:143)), Gateway도 전송 직전 재검사합니다([Gateway:658–715](/tmp/claude-1000/a112-review-80ae96a5/internal/execgw/gateway.go:658)). R2~R5는 승인된 이월로 유지했습니다.

Recommendation: BLOCK — 비활성 시장의 마지막 개수 관문과 적재 오류의 J4 분류를 수리하고, CAS 논증을 정정한 뒤 격리 종단 시험으로 재검증하십시오.