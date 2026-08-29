# Change: attest-covers-supervised-mutations

## Why

인터록 5절(`att.Verify(..., RequiredEndpoints())`)이 요구하는 **8개** endpoint 중
mutation 2개는 attestation에 들어갈 **경로 자체가 없다**. 게이트 조항이 구조적으로
만족 불가능하다.

```
internal/app/engine.RequiredEndpoints()   GET ×6 + POST /api/v1/orders
                                                 + POST /api/v1/orders/{id}/cancel
attestation을 쓰는 곳                      cmd/tossctl/soak.go:403  ← 유일
그 입력                                    soak 기록 (read-only, 설계상 POST 불가)
BuildAttestation                          비-GET을 명시적으로 거부 (attest.go:187-193)
```

`internal/soak.LiveOnlyEndpoints()`가 그 두 개를 이름으로 들고 있고 주석이 이렇게 적고 있다 —
"They are reported, never attested. They come from the supervised one-off live verification
(task 2.2) … until that has run, an attestation written here covers the reads and nothing else."
**의도된 미완성이고, 후자를 접어 넣는 코드가 아직 없다.**

증거는 이미 있다. 2026-07-27~28 감독 하 실행이 두 시장에서 두 POST를 모두 성공시켰다
(사람이 배치 단위로 승인한 실주문):

| endpoint | KR 성공 | US 성공 |
| --- | --- | --- |
| `POST /api/v1/orders` | 5 | 7 |
| `POST /api/v1/orders/{id}/cancel` | 4 | 6 |

그리고 `verifylive.Call.Endpoint`는 주석이 말하듯 **"internal/soak과 internal/app/engine이
쓰는 철자 그대로"** 기록된다 — 세 곳이 번역표 없이 비교되도록 어휘가 이미 맞춰져 있다.
빠진 것은 판독뿐이다.

2026-07-29에 이것이 실무 문제로 드러났다: attestation 자동 발급 감시자가 07-27부터 멎어
있었고, 복구해 발급해 보니 endpoint 6/8로 나왔다. 게이트를 켤 수 없는 것이 **설정 문제가
아니라 미구현**이라는 사실이 그때까지 어디에도 드러나 있지 않았다.

## What Changes

- **감독 검증 기록이 mutation endpoint의 증거원이 된다.** `soak attest`가 시장별 검증
  기록을 함께 읽고, 그 기록이 성공을 증명하는 mutation endpoint를 attestation에 싣는다.
- **soak과 검증의 역할 분리는 그대로다.** 검증 기록은 `LiveOnlyEndpoints()`에 있는 것만
  기여할 수 있다. GET은 **언제나** soak이 벌어야 한다 — 감독 하 1회 성공은 4일 무인 운전과
  같은 것을 증명하지 않는다.
- **네 가지 조건이 모두 참일 때만 싣는다**: ① 그 endpoint의 호출이 **성공**했다(오류 없음)
  ② 기록의 계좌가 soak의 계좌와 **같다** ③ 성공 시점이 attestation 유효 기간 안이다
  ④ 그 endpoint가 `LiveOnlyEndpoints()`에 있다.
- **출처를 attestation에 남긴다.** 어떤 파일·시장·시각이 어느 endpoint를 증명했는지
  기록한다 — 한 달 뒤 감사자가 "이 게이트는 무엇을 근거로 켜졌나"에 답할 수 있어야 한다.
- **한 개라도 못 채우면 그대로 부족한 채 발급한다.** 인터록 5절이 `MissingEndpoints`로
  거부하는 동작은 무변경이다.

## Non-Goals

- `RequiredEndpoints()`·인터록 조항 변경 — 없다. 이 change는 **증거를 나르는 배선**이지
  요구를 낮추는 것이 아니다.
- 검증 기록이 GET을 기여하는 것 — 금지. soak의 의미를 보존한다.
- soak이 mutation을 시도하게 만드는 것 — 금지. read-only는 그 도구의 정의다.
- 게이트 9절(`ProtectionReady`) 완화 — 무관하고 무변경. `const profileProtection =
  ProtectionUnwired`는 어떤 설정으로도 만족 불가능하므로 **이 change로는 엔진이 기동할 수
  없다**. 2c `add-protection-orders`가 그 상수를 뒤집기 전까지는 아무것도 켜지지 않는다.
- 게이트 flip 자동화 — 금지 유지(§0.7).
- 검증 기록을 쓰는 쪽(verifylive 측정 경로) 변경 — 없다. 읽기만 추가한다.

## Capabilities

### Modified Capabilities

- `engine-safety`: attestation의 endpoint 집합이 **어디서 오는가**를 규정한다 —
  읽기는 무인 soak, mutation은 사람이 승인한 감독 검증

## Impact

- Affected code: `internal/verifylive`(성공 endpoint 판독 — 신규 읽기 함수),
  `internal/soak/attest.go`(`BuildAttestation`이 감독 증거를 받는다),
  `internal/attest/attest.go`(출처 필드 — 가산·선택적),
  `cmd/tossctl/soak.go`(시장별 검증 기록 해석·전달·출력)
- High-risk 여부: **yes** — 게이트 5절을 만족 가능하게 만든다. 적대적 Eng 리뷰 +
  Pre-Edit 선언 + Function Logic Map 필수.
- 안전 검토(§0):
  - **이 change 단독으로는 게이트가 열리지 않는다.** 9절이 컴파일 타임 상수다
    (`interlock.go:175`). 이것이 이 change의 위험을 지배적으로 낮춘다.
  - 증거는 조작할 수 없다. 실린 endpoint는 사람이 배치 승인한 실행이 **성공시킨** 호출뿐이고,
    계좌·시각·목록 세 가지로 다시 좁힌다.
  - §0.4 rate limit: 새 브로커 호출 **0건**. 이미 디스크에 있는 기록을 읽을 뿐이다.
  - §0.7: 게이트 flip은 여전히 사람이 한다. 이 change는 flip의 **선행 조건**을 증거로
    채울 뿐 flip을 하지 않는다.
  - 최악의 오작동은 "실제로 증명된 endpoint를 못 싣는다"(게이트가 계속 거부 — 안전 방향)
    이거나 "증명되지 않은 것을 싣는다"인데, 후자는 네 조건 전부를 우회해야 하므로
    각 조건마다 거부 테스트를 둔다.
