## ADDED Requirements

### Requirement: 미국 include-only 편입 경로는 실제 engine boundary에서 회귀 검증된다
미국 보유분은 market-agnostic include 규칙과 기존 신선도/대사 조건을 충족하면 한국 보유분과 동일한 fold→adopt→exit t0 경로를 사용해야 한다 (SHALL). adoption quote는 candidate market과 일치하는 currency provenance(KR→KRW, US→USD)를 가져야 하며 (SHALL), currency가 비었거나 다르거나 같은 symbol이 서로 다른 candidate market에 중복되면 편입을 연기해야 한다 (MUST). 다른 symbol/market의 가격을 사용해서는 안 된다 (MUST NOT).

#### Scenario: 미국 include-only 보유분 편입
- **WHEN** Verified engine에서 adoption.enabled=false, include_symbols=[AAPL], stable/fresh US AAPL holding 두 관측과 fresh official AAPL USD quote 200이 주어진다
- **THEN** 다음 정상 RunOnce는 Folded=1, Adopted=1, Unmanaged=0이고 external-adoption provenance, t0 entry/high-water 200, 5% synthetic initial-stop/baseline 190을 하나의 편입 transaction으로 영속한다

#### Scenario: account-wide quantity mismatch가 미국 편입을 보류한다
- **WHEN** 같은 candidate에 account-wide permanent quantity-mismatch block이 active다
- **THEN** RunOnce는 adoption transaction과 exit state를 만들지 않고 read projector는 미국 미지원이 아닌 `RECONCILE_BLOCKED`로 설명한다

#### Scenario: 미국 candidate에 KRW 또는 무통화 quote가 온다
- **WHEN** US AAPL candidate의 quote currency가 KRW이거나 비어 있다
- **THEN** 편입과 exit t0 생성은 연기되고 잘못된 가격 provenance는 journal에 영속되지 않는다

#### Scenario: 같은 symbol이 KR/US candidate에 동시에 있다
- **WHEN** 동일 symbol이 KR과 US의 미편입 candidate로 동시에 존재하고 quote transport가 symbol-only다
- **THEN** market identity가 모호하므로 두 candidate 모두 편입되지 않는다
