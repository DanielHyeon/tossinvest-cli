# a199 정본 delta 의도 (Manager 작성 — 1단계 팀메이트가 `specs/engine-safety/spec.md` 로 조립)

조립 규칙(이전 change 들과 같음):
- MODIFIED 블록은 **생성 스크립트**로 만든다: 정본 Requirement 전문을 읽어 아래 「편집」의 exact-string 치환을 적용하고, 치환마다 「정확히 1회 적중」을 단언하며, 역치환으로 「의도한 편집 외 차이 0」을 단언한다. 정본의 Scenario 이름은 전부 남긴다(수정하는 Scenario 도 이름 유지 여부를 아래에 적었다). 스크립트는 `analysis/modified-delta/` 에 커밋한다.
- ADDED Requirement 본문은 500자 이하(strict). 넘으면 쪼갠다.
- 아래 문장은 **의도**다. 정본의 실제 문자열과 글자가 다르면(공백·줄바꿈·기호) 정본을 기준으로 치환 문자열을 다시 잡고, 뜻이 달라지는 조정이 필요하면 멈추고 Manager 에게 묻는다.
- `rtk proxy openspec validate <id> --strict` 가 valid 여야 한다.

---

## A. MODIFIED — 「자동화 게이트 기동 인터록」 (engine-safety spec.md:106 근처)

### 편집 A1 — 조항 2 의 문장
- 전: `2. 유효한 capability attestation(만료·계좌 식별·성공 endpoint 집합 — verify-execution-capability change가 생성) 존재·미만료·계좌 일치.`
- 후: `2. capability attestation(만료·계좌 식별·성공 endpoint 집합 — verify-execution-capability change가 생성)이 존재하고 읽히며 온전하고 계좌가 일치. 만료와 endpoint 누락(낡음)은 기동을 거부하지 않고 아래 「낡은 증명」 절이 정한다.`
- (그 뒤의 「attestation endpoint 집합은 … drift 가드로 동기화한다」 문장은 그대로)

### 편집 A2 — 보호 절 앞에 새 문단 삽입
삽입 위치: `게이트 flip은 사람 승인 절차(§0.7)와 audit 기록을 요구한다(SHALL).` 바로 뒤, `**브로커측 보호 실행 배선(ProtectionReady)은 기동 조건이 아니라 진입 허가 조건이다**(SHALL).` 바로 앞.

삽입할 문단:

```
**capability attestation 의 낡음(만료 · 필요 endpoint 누락)은 기동 조건이 아니라 진입 허가 조건이다**(SHALL).
근거는 보호 배선과 같은 실패의 비대칭이다 — 증명이 낡은 경로로 진입하면 손절 없는 새 포지션이 생기지만,
그 경로로 청산을 시도하다 실패하면 남는 것은 런타임을 거부했을 때와 같은 상태다.
attestation 이 읽히고 온전하며 해석된 계좌를 가리키는데 낡았으면, 그리고 나머지 조항이 전부 충족되면:

- 런타임은 **청산 전용으로 기동한다**(SHALL). 루프 집합은 변하지 않는다.
- 노출을 **증가**시키는 mutation 은 거부된다(SHALL). 집행은 보호 배선과 같은 mutation chokepoint 한 곳이고
  판정은 mutation 의 형태에서 계산한다.
- 이 진입 차단은 매 기동마다 attestation 에서 다시 유도된다(SHALL) — 재시작으로 풀리지 않고,
  attestation 이 유효해진 뒤의 기동에서는 사람의 해제 절차 없이 사라진다.
- 낡음 밖의 attestation 결함(없음 · 읽을 수 없음 · 불완전 · 알 수 없는 형식 · 계좌 불일치)은 기동을 거부한다(SHALL).
  만료된 attestation 의 계좌 일치도 따로 확인해야 한다(SHALL — 만료가 계좌 불일치를 가려서는 안 된다).
- 낡음과 다른 조항의 결함이 함께 있으면 기동을 거부한다(SHALL).
- 청산 전용 기동은 수락도 거부도 아닌 결과로 audit 에 기록되고, 운영자에게 기동 시점에 전달된다(SHALL —
  읽는 문장이며 타이핑 확인·추가 승인 마찰이어서는 안 된다(SHALL NOT)).
- 실행 중인 엔진은 attestation 이 낡아져도 동작이 바뀌지 않는다(SHALL NOT) — 판정은 기동 시점뿐이다.
```

### 편집 A3 — Scenario 「attestation 만료 상태 기동」 (spec.md:137) 의 THEN
- 이름 유지. 본문을 읽어 WHEN 은 그대로 두고 THEN 을 다음 뜻으로 바꾼다: `런타임은 청산 전용으로 기동하고, 노출 증가 mutation 은 거부되며, exit 관측·대사·체결 감지는 돈다`. (정본의 현재 THEN 문자열을 exact 로 잡아 치환)
- WHEN 이 「계좌 일치」를 말하지 않으면 `이고 그 attestation 이 해석된 계좌를 가리킬 때` 를 WHEN 에 더한다.

### 추가 Scenario (같은 Requirement 끝, `#### Scenario: dry-run` 앞 또는 뒤 — 정본 순서를 깨지 않는 자리)

```
#### Scenario: 필요 endpoint 가 빠진 attestation 으로 기동
- **WHEN** 게이트 ON 이고 attestation 이 만료 전이며 계좌가 일치하지만 주문 접수 endpoint 를 덮지 않고, 나머지 조항은 충족된 채 엔진이 기동한다
- **THEN** 런타임은 청산 전용으로 기동하고 매수는 거부되며 보유의 손절 매도는 제출된다

#### Scenario: 만료됐고 다른 계좌를 가리키는 attestation
- **WHEN** attestation 이 만료됐고 그 계좌가 자격 증명이 해석한 계좌와 다르다
- **THEN** 기동은 거부된다

#### Scenario: 낡은 attestation 과 한도 출처 불일치
- **WHEN** attestation 이 낡았고 Guardian 의 한도가 인터록이 감사한 한도와 다르다
- **THEN** 기동은 한도 출처 불일치로 거부된다

#### Scenario: attestation 을 갱신한 뒤의 재기동
- **WHEN** 청산 전용으로 돌던 엔진을 유효한 attestation 으로 다시 기동한다
- **THEN** 진입 차단은 사람의 해제 절차 없이 사라지고 audit 에는 수락이 기록된다

#### Scenario: 낡은 채 재기동
- **WHEN** 청산 전용으로 돌던 엔진을 attestation 을 고치지 않고 다시 기동한다
- **THEN** 런타임은 다시 청산 전용으로 기동한다
```

---

## B. MODIFIED — 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」 (spec.md:2213 근처)

판정·등급·주기·키·기한 계산은 무변. 바뀌는 것은 **낡음 단계의 귀결 문장**뿐이다.

### 편집 B1 — 첫 문단의 질문
- 전: `「지금 재시작하면 무장하는가」`
- 후: `「지금 재시작하면 무엇으로 뜨는가(정상 · 청산 전용 · 거부)」`

### 편집 B2 — 귀결 문장
- 전(뜻): `재시작이 거절될 상태의 알림은 제목과 본문의 첫 문장에서 그 결과 — 지금 엔진을 재시작하면 기동이 거부된다 — 를 조건문이 아닌 사실로 말해야 하고 (SHALL), 본문에서 그동안 자동 손절이 돌지 않는다는 것과 운영자가 할 일을 말해야 한다 (SHALL).`
- 후: `재시작의 결과가 정상이 아닌 상태의 알림은 제목과 본문의 첫 문장에서 그 결과를 조건문이 아닌 사실로 말해야 한다 (SHALL) — 낡음(만료 · 범위 상실)이면 「지금 엔진을 재시작하면 새 진입이 막힌 청산 전용으로 뜬다」, 그 밖(다른 계좌 · 없음 · 읽을 수 없음 · 불완전)이면 「지금 엔진을 재시작하면 기동이 거부된다」. 거부되는 상태의 본문은 그동안 자동 손절이 돌지 않는다는 것을, 청산 전용인 상태의 본문은 손절·청산은 돌고 새 진입만 막힌다는 것을 말해야 하며 (SHALL), 둘 다 운영자가 할 일을 말해야 한다 (SHALL). 엔진이 이미 청산 전용으로 돌고 있으면 알림은 그것을 현재의 사실로 말해야 한다 (SHALL).`
- 어느 상태가 청산 전용으로 내려가는지의 판정은 인터록의 것과 **같은 함수**여야 한다는 문장을 `판정 규칙을 인터록과 따로 두어서는 안 된다 (SHALL NOT).` 뒤에 한 문장으로 잇는다: `어느 결함이 기동을 거부하고 어느 결함이 청산 전용 기동이 되는지도 같다 (SHALL).`

### 편집 B3 — 「재시작이 거절될 상태(만료 · 범위 상실 · 계좌 불일치 · 무효)는 고쳐질 때까지 하루 한 번」
- 표현만: `재시작이 거절될 상태` → `재시작의 결과가 정상이 아닌 상태`. 괄호 안 목록·주기·키 규칙은 그대로.
- 같은 표현이 다른 문장·Scenario 에 더 있으면(a196 이 넣은 것) **자리마다** 뜻을 확인해 고친다 — 자리 수를 세어 보고한다.

### Scenario
- 기존 18개 이름 유지. 「만료됨」「갱신이 주문 권한을 잃음」 의 THEN 에 알림 문장의 귀결이 적혀 있으면 새 귀결로 고친다(없으면 무편집).
- 추가:

```
#### Scenario: 청산 전용으로 돌고 있는 엔진
- **WHEN** 낡은 attestation 으로 청산 전용 기동한 엔진에서 판정이 돈다
- **THEN** critical 알림은 엔진이 지금 청산 전용으로 돌고 있고 새 진입이 막혀 있음을 현재의 사실로 말한다

#### Scenario: 거부가 유지되는 결함
- **WHEN** 실행 중인 엔진이 읽을 attestation 이 다른 계좌를 가리킨다
- **THEN** critical 알림은 지금 재시작하면 기동이 거부되고 그동안 자동 손절이 돌지 않는다고 말한다
```

---

## C. ADDED (필요할 때만 — A2 로 충분하면 만들지 않는다)

A2 의 문단이 한 Requirement 안에서 strict 를 통과하면 ADDED 는 없다. 한 Requirement 가 지나치게 길어 리뷰가 어렵다고 1단계가 판단하면, A2 문단을 다음 제목의 ADDED 로 빼고 A 에는 참조 한 문장만 둔다(Manager 에게 보고 후):
- `### Requirement: 낡은 capability 증명은 진입만 막는다`

---

## D. 다른 capability 의 정본
- risk-management 「Kill switch와 운영 모드」: 무편집(운영 모드를 쓰지 않는다). `EXIT_ONLY` 삭제 서술(risk-management spec.md:131)과 모순되는 문장을 이 delta 에 넣지 않는다 — 「청산 전용」은 운영 모드의 이름이 아니라 기동의 결과다.
- order-execution · exit-policy: 무편집 예상. 1단계가 「진입 래치 사유」를 열거하는 정본 문장(닫힌 목록)이 있는지 확인하고, 있으면 그 Requirement 를 MODIFIED 로 더한다.
- operator-console: 무편집(비목표). Story limits 에 「콘솔 문구 부정확」을 적는다.
