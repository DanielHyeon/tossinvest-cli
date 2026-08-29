# Review: attest-covers-supervised-mutations

## Pre-Edit Gate (High-risk)

```text
Pre-Edit Gate:
- change id / task id: attest-covers-supervised-mutations / 1.1~1.10
- 대상 심볼(패키지.함수):
    internal/soak.BuildAttestation        (기존 함수 내부 편집)
    internal/soak.acceptSupervised        (신규 leaf — 정책이 여기 있다)
    internal/soak.mutationNote            (신규 leaf)
    internal/soak.normaliseEndpoint       (신규 leaf)
    internal/attest.Attestation           (가산 필드 SupervisedBy)
    internal/attest.Proof                 (신규 타입)
    internal/attest.SameAccountMasked     (신규 leaf)
    internal/verifylive.SucceededEndpoints(신규 파일 endpoints.go)
    cmd/tossctl.supervisedProofs          (신규)
    cmd/tossctl.runSoakAttest             (기존 함수 내부 편집)
    cmd/tossctl.newSoakAttestCmd          (플래그 1개 추가)
- 기존 동작 파악 근거:
    · attestation을 쓰는 유일한 곳: cmd/tossctl/soak.go:403 (변경 전 기준). grep으로 확인
    · 인터록 조항 순서와 5·9절: internal/app/engine/interlock.go:449~516
    · 9절이 상수임: interlock.go:175 `const profileProtection = ProtectionUnwired`
    · 기존 테스트: internal/soak/attest_test.go 12건, cmd/tossctl/soak_test.go의
      TestSoakAttestDoesNotSatisfyTheEngineInterlockOnItsOwn,
      TestSoakAndLiveEndpointsCoverTheEngineInterlock,
      internal/app/engine/interlock_test.go의 TestTheGateRefusesWithoutBrokerSideProtection
    · 실측: capability-verify.jsonl / capability-verify-us.jsonl의 endpoint별 성공·실패 집계
      (POST /api/v1/orders KR 5·US 7, POST /api/v1/orders/{id}/cancel KR 4·US 6)
    · Call.Endpoint 철자 계약: internal/verifylive/record.go의 Call 주석
- upstream 상속 테스트 영향: no — internal/soak·internal/attest·internal/verifylive는
  TossOS 신규 패키지다. 전체 `go test ./...` 3723 → 3742 (신규 19건), 회귀 0.
- 실패 테스트 선행 작성: yes — RED 2회 관측·기록(컴파일 실패 1회, 배선 되돌림 1회).
  branch-test-map.md에 실패 출력 원문 포함.
- 안전 불변식 §0 위반 여부 검토: 통과 (아래 §0 대조표)
```

### §0 대조

| 조항 | 판단 |
|---|---|
| §0.1 승인 없는 LIVE side effect 금지 | 통과 — 브로커 호출 **0건**. 이미 디스크에 있는 기록을 읽을 뿐이다. |
| §0.2 토글 OFF = upstream | 해당 없음 |
| §0.3 손절·비상 청산 즉시성 | 해당 없음 — 엔진 런타임 무변경 |
| §0.4 rate limit 계상 | 통과 — 새 호출 종류 0, 호출 수 0 |
| §0.5 운영 설정 audit | 통과 — 게이트 수락 audit(`acceptanceDetail`) 무변경. attestation 자체가 무엇으로 통과했는지를 새로 담는다(개선 방향) |
| §0.6 스키마 변경 | 통과 — `SupervisedBy`·`Proof`는 **가산·`omitempty`**. `FormatVersion` 불변, 옛 파일 그대로 읽힘(테스트 고정) |
| §0.7 운영 토글 flip은 사람이 | 통과 — 이 change는 flip을 하지 않는다. flip의 **선행 조건을 증거로 채울** 뿐이다 |
| §0.8 scope 밖 주문·위험 코드 변경 금지 | 통과 — 인터록·`RequiredEndpoints()`·엔진 런타임 무변경 |
| §0.9 보수 방향만 | 통과 — 최악의 오작동 방향이 "증명된 것을 못 싣는다"(게이트가 계속 거부)이다 |

## 리뷰 (2026-07-29, 적대적 Eng 관점 포함)

### A1. "게이트를 열기 쉽게 만드는 change 아닌가" — **핵심 논점, 수용하되 반박**

이 change **단독으로는 게이트가 열리지 않는다.** 9절이
`const profileProtection = ProtectionUnwired`이고, 설정·플래그·환경변수 어느 것도 그것을
바꿀 수 없다. `SetProtectionReadyForTest`는 `_test.go` 파일에만 있어 빌드된 바이너리의 API에
존재하지 않는다. `TestTheGateRefusesWithoutBrokerSideProtection`이 "모든 앞선 절을 통과한
설정으로도 9절이 거부한다"를 이미 고정하고 있고, 이 change는 그 테스트를 건드리지 않는다.

즉 이 change가 하는 일은 **하나의 미구현을 다른 미구현 뒤에서 메우는 것**이다. 게이트가
열리는 시점은 2c가 보호주문을 배선하고 그 상수를 뒤집을 때이며, 그때 사람이 §0.7로 flip한다.

### A2. "감독 검증 1회를 무인 4일과 같이 취급하는가" — **아니다, 그것이 닫힌 집합의 이유**

허용 집합은 `LiveOnlyEndpoints()` 두 개로 닫혀 있다. GET을 감독 기록에서 받으면 soak이
**실패한** 읽기를 감독 1회가 세탁하게 된다 — `TestSupervisedEvidenceCannotStandInForTheSoak`이
그것을 막는다. 이 거부는 기존의 반대편 거부(soak 기록의 비-GET 거부)와 대칭이고, 둘이 함께
있어야 각 증거원이 자기 몫만 증명한다.

### A3. "검증 기록은 조건주문 endpoint도 성공시켰다. 왜 안 싣나" — **요구되지 않았기 때문**

`POST /api/v1/conditional-orders`·`DELETE /api/v1/conditional-orders/{id}`도 성공 기록이
있지만 인터록이 요구하지 않는다. 요구되지 않은 능력을 attestation이 주장하면, 2c가 그
목록을 늘릴 때 **근거를 다시 따지지 않은 항목이 이미 자리를 잡고 있게 된다**.
`TestSupervisedEvidenceIsClosedToOtherMutations`가 고정한다.

### A4. "계좌 불일치를 왜 건너뛰지 않고 거부하나" — 의도적

기대 경로에 다른 계좌의 검증 기록이 있다는 것은 설정 오류다. 건너뛰면 그 오류가
"아직 검증 안 함"과 **구별되지 않고**, 운영자는 없는 검증을 다시 돌리려 할 것이다.
반면 오래된 증거는 평범한 상태(검증 후 시간이 지남)이므로 건너뛴다. 두 처리를 다르게 한
근거를 design.md D3과 코드 주석에 남겼다.

### A5. "유효 기간을 `Validity`(30일)로 잡은 근거는" — **`MaxRecordAge`(48h)를 의도적으로 기각**

48시간을 쓰면 게이트를 열어 두기 위해 **이틀마다 실주문을 내야** 한다. 그것은 감독 검증을
자동화하라는 압력이 되고, 이 시스템 전체가 막으려는 방향이다. attestation 자체가 30일 뒤
만료되어 재검증을 강제하므로, 증거의 창을 같은 30일로 잡는 것이 일관된다.

실무 귀결: 2026-07-27~28 증거가 오늘 유효하다. 실기록으로 확인함(아래).

### A6. "정규화한 철자를 실으면 되지 않나" — **안 된다**

`Call.Endpoint`는 세 패키지가 번역표 없이 비교하도록 철자를 맞춰 둔 값이다. 대문자로
정규화한 키를 그대로 실으면 **네 번째 철자 권위**가 생기고 drift 가드가 무의미해진다.
구현 중 실제로 이 실수를 했고 테스트가 잡았다(`SucceededEndpoints`의 첫 GREEN 시도가
3건 실패). 정규화는 "같은 endpoint인가" 판단에만 쓰고 기록에는 원래 철자를 쓴다.

### A7. "`internal/soak`이 `internal/verifylive`를 import하게 되지 않나" — 안 된다

판독은 verifylive가, 조립은 cmd가, 정책은 soak가 한다(design.md D4). read-only 도구가
측정 도구의 타입을 끌고 들어오지 않는다. `internal/soak/static_test.go`의 import 그래프
단언이 그대로 통과한다.

### A8. "실기록에서 정말 되는가" — 확인함 (계좌 호출 0건)

```
./bin/tossctl soak attest --out /tmp/…
  endpoints    GET ×6 + POST /api/v1/orders + POST /api/v1/orders/{id}/cancel
  supervised   POST /api/v1/orders ← capability-verify.jsonl (KR, 2026-07-28T01:34:35Z)
  supervised   POST /api/v1/orders/{id}/cancel ← capability-verify.jsonl (KR, 2026-07-28T01:34:35Z)

  Every endpoint the automation gate requires is now covered.
  The gate will still refuse to start: … interlock clause 9 …
```

8/8이 채워지고, 같은 출력이 게이트가 여전히 거부한다고 말한다. 운영자가 "8/8"만 보고
반대로 결론 내리지 않게 하는 것이 그 문장의 목적이다.

### A9. 남은 위험 — `LiveOnlyEndpoints()`가 넓어질 때

2c가 보호주문 endpoint를 인터록 요구 집합에 추가하면 `LiveOnlyEndpoints()`도 넓어지고,
그와 동시에 감독 증거의 허용 범위도 넓어진다. 그때 이 함수의 거부 테스트를 함께 갱신해야
한다. FLM `Safety conclusion`과 `internal-soak--buildattestation`에 남겼다.

### 결정

수용하고 진행한다. 미해결 코드 이슈 없음. A9는 2c로 이연(코드가 아니라 그때 지켜야 할 규율).

## Function Logic Map

적용함. 17개 target — 수정한 기존 함수 2건(`BuildAttestation`, `runSoakAttest`/
`newSoakAttestCmd`), 신규 leaf, 이 change의 테스트 함수들, 그리고 파일 끝 덧붙임으로
diff 문맥에 걸렸을 뿐 **수정하지 않은** `seedQualifyingRecord`(base revision으로 고정).

`python3 tools/logic-map/check_analysis.py --change attest-covers-supervised-mutations`
→ `evidence complete or diff-proven exempt`

## 운영 기록: attestation 감시자

이 change의 계기는 2026-07-29에 attestation 자동 발급 감시자가 07-27부터 멎어 있었던 것이다
(crontab·systemd timer·프로세스 어디에도 없었고 로그 마지막 기록이 07-27 14:57). 그 사이
soak은 필요한 조건을 모두 채웠고 아무도 발급하지 않았다. 감시자는 같은 날
`~/.config/systemd/user/tossos-attest.{service,timer}`로 복구했다 — 6시간 간격,
`Persistent=true`(절전으로 창을 놓치면 깨어날 때 실행). 이 change의 코드 범위 밖이지만,
같은 실패가 재발하면 이 change의 효과도 같이 사라지므로 여기에 기록한다.
