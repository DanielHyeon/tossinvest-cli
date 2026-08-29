# Design: attest-covers-supervised-mutations

## D1. 어휘는 이미 맞춰져 있다 — 번역표를 만들지 않는다

`verifylive.Call.Endpoint`의 주석: *"the call as `METHOD /path`, spelled the way
internal/soak and internal/app/engine spell their endpoint sets so the three compare
without a translation table."*

세 패키지가 같은 문자열을 쓴다. 이 change는 그 사실에 기대며, **매핑 테이블을 만들지
않는다**. 만드는 순간 네 번째 철자 권위가 생기고 drift 가드가 무의미해진다.

실측 확인: 검증 기록의 `POST /api/v1/orders`·`POST /api/v1/orders/{id}/cancel` 문자열이
`engine.RequiredEndpoints()`·`soak.LiveOnlyEndpoints()`와 바이트 일치한다.

## D2. 기여 가능 집합은 `LiveOnlyEndpoints()`로 **닫는다**

**결정**: 검증 기록은 `soak.LiveOnlyEndpoints()`에 있는 endpoint만 기여할 수 있다.
그 밖의 것이 검증 기록에 있어도(있다 — `POST /api/v1/conditional-orders`,
`DELETE /api/v1/conditional-orders/{id}`, `GET /api/v1/sellable-quantity` 등) **싣지 않는다**.

| 대안 | 기각 사유 |
| --- | --- |
| 성공한 endpoint 전부 기여 | GET이 감독 하 1회 성공으로 실린다. soak의 4일 무인 운전이 증명하는 것과 전혀 다른 것을 같은 이름으로 부르게 된다. |
| soak이 못 덮은 것 전부 기여 | soak이 **실패한** GET을 검증 1회가 덮어버린다. soak 결함이 조용히 세탁된다. |
| 비-GET 전부 기여 | 조건주문 endpoint가 실린다. 인터록이 요구하지도 않는 것을 attestation이 주장하게 되고, 2c가 그 목록을 늘릴 때 근거 없는 항목이 이미 자리를 잡고 있다. |

`BuildAttestation`이 이미 대칭인 거부를 갖고 있다 — soak 기록의 비-GET을 거부한다
(attest.go:187-193). 이 change는 그 반대편을 추가한다: 검증 기록의 `LiveOnlyEndpoints()`
바깥을 거부한다. 두 거부가 합쳐져 **각 증거원이 자기 몫만 증명**한다.

## D3. 네 조건은 모두 AND이고, 각각 거부 테스트를 갖는다

| # | 조건 | 없으면 생기는 일 |
| --- | --- | --- |
| 1 | 그 호출이 **성공**했다 (`Call.Error == ""`) | 422로 거절된 접수가 "POST가 동작한다"의 증거가 된다 |
| 2 | 검증 기록의 계좌 == soak의 계좌 | 다른 계좌에서 증명된 능력으로 이 계좌의 게이트가 열린다 |
| 3 | 성공 시각이 유효 기간 안 | 30일짜리 attestation이 1년 전 증거에 기댄다 |
| 4 | endpoint ∈ `LiveOnlyEndpoints()` | D2 |

계좌 비교: soak 기록은 `18301045921`, 검증 기록은 `*******5921`(마스킹)이다.
`attest.Mask(soak.AccountRef) == verify.AccountRef`로 결속한다 — 실측 일치 확인함.
검증 기록이 **다른** 계좌를 말하면 건너뛰지 않고 **거부**한다: 기대 경로에 다른 계좌의
기록이 있다는 것은 설정 오류이며, 조용히 무시하면 그 오류가 "증거 없음"으로 위장된다.

시각 기준은 `Call.At`이 아니라 그 호출을 담은 **entry의 `FinishedAt`**를 쓰지 않는다 —
`Call.At`이 호출 자체의 시각이고 우리가 재는 것이 "그 호출이 언제 성공했나"다.

## D4. 패키지 의존 방향

`internal/soak`이 `internal/verifylive`를 import하지 **않는다**. 판독은
`internal/verifylive`가 자기 기록에 대해 답하고(`SucceededEndpoints`), 조립은
`cmd/tossctl`이 하며, 정책(D2·D3)은 `internal/soak`이 강제한다.

```
internal/verifylive   SucceededEndpoints(entries, now, maxAge) → map[endpoint]Proof
        ↓ (cmd/tossctl/soak.go가 시장별로 읽어 합친다)
internal/soak         BuildAttestation(..., supervised map[string]attest.Proof)
                      ← LiveOnlyEndpoints() 밖이면 거부, 계좌 불일치면 거부
        ↓
internal/attest       Attestation.SupervisedBy []Proof   (가산·선택적 필드)
```

`internal/soak`이 verifylive를 알면 read-only 도구가 측정 도구의 타입을 끌고 들어오고,
D7 계열의 격리(발굴이 엔진을 모른다)와 같은 이유로 피한다. 대신 `attest.Proof`라는
**두 패키지가 공유하는 작은 값 타입**을 `internal/attest`에 둔다 — 이미 둘 다 그것을 import한다.

## D5. 출처를 남긴다

`Attestation.SupervisedBy`(가산·`omitempty`)에 endpoint별로 무엇이 증명했는지 남긴다:
endpoint, 시각, 기록 파일, 시장. 근거:

- 인터록 5절 거부 메시지는 "무엇이 빠졌나"만 말한다. 통과했을 때 "무엇으로 통과했나"는
  attestation 자신 말고 답할 곳이 없다.
- audit 항목(`acceptanceDetail`)은 attestation 만료만 싣는다. 한 달 뒤 "이 게이트는 무슨
  근거로 켜졌나"에 답하려면 이 줄이 필요하다.

format_version은 올리지 않는다 — 가산·선택적 필드이고 옛 파일은 그대로 읽힌다
(§0.6 additive-nullable 선호).

## D6. 범위 밖

- 2c가 보호주문 endpoint를 `RequiredEndpoints()`에 추가할 때 그 증거도 같은 경로로 들어온다.
  이 change는 그 자리를 만들어 둘 뿐 목록을 늘리지 않는다.
- `--verify-record` 재정의 플래그는 둔다(테스트·비표준 배치용). 기본은 `verify`가 쓰는
  것과 **같은 해석 경로**다 — 플래그를 기억해야 동작하는 설계는 2026-07-27~29에 감시자가
  멎은 것과 같은 종류의 실패를 부른다.
