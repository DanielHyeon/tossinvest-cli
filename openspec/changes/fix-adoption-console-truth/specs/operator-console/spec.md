## ADDED Requirements

### Requirement: active profile journal identity
The operator console SHALL resolve its read-only journal from the same active
profile rule as the engine: when an explicit `--config-dir` is present it SHALL
read `<config-dir>/journal.db`, and without that override it SHALL use the
default data-directory journal. The console SHALL continue to open the selected
file through the read-only journal API and SHALL NOT create, migrate, copy,
merge, modify, or fall back to another journal. A selected journal that lacks
the required v9 `exit_states.policy_id` column SHALL be classified as too old
before account queries run.

#### Scenario: explicit profile has a newer journal than the default data directory
- **WHEN** the console starts with `--config-dir P`, `P/journal.db` is a readable current-schema journal, and the default data-directory journal is stale or incompatible
- **THEN** the positions screen reads management state only from `P/journal.db` and does not query or modify the unrelated default journal

#### Scenario: no explicit profile
- **WHEN** the console starts without `--config-dir`
- **THEN** it resolves the same default journal path as before and opens it read-only

#### Scenario: selected profile journal is unavailable
- **WHEN** the selected profile journal is missing, too new, or too old
- **THEN** the corresponding existing failure state is rendered without falling back to another journal

#### Scenario: selected v8 journal lacks policy identity
- **WHEN** the selected journal has the v8 tables but lacks `exit_states.policy_id`
- **THEN** the console renders the too-old migration guidance before any positions query and never exposes a raw `no such column: policy_id` error

## MODIFIED Requirements

### Requirement: 편입 설정 화면
콘솔은 편입 설정 화면을 제공해야 한다(SHALL): 상단 navigation은 **"외부 종목 자동관리"** 메뉴를 표시하고 `/settings#adoption`으로 기존 편입 설정의 첫 섹션을 직접 열어야 한다(SHALL). 화면 제목과 첫 설명은 수동 매수한 외부 종목이 편입 완료 후 기존 공통 익절·보호선·손익 극대화 정책의 관리 대상이 된다는 결과와, 저장 자체가 편입·주문을 실행하지 않고 엔진 대사 루프가 다음 기동부터 수행한다는 경계를 명시해야 한다(SHALL).

화면은 `adoption.enabled`·`default_stop_pct`·`exclude_symbols`·`include_symbols`를 표시·편집한다(SHALL). 합성 손절폭은 기존 CSP 아래에서도 현재값과 편집값이 일치하도록 **명시적 백분율 숫자 입력**으로 표시한다(SHALL — 이전 마우스 슬라이더/타이핑 금지 결정은 CSP가 inline 갱신을 차단해 현재값을 거짓 표시하므로 이 요구가 대체한다). 입력 필드명은 `default_stop_percent`이고, 허용값은 유한한 2%..20%의 0.5 percentage-point 단위다(SHALL). 서버는 브라우저 검증을 신뢰하지 않고 empty·비수치·NaN·무한대·범위 밖·단위 밖·과거 fraction 형식 제출을 저장 전에 거부하며(SHALL), 허용 percentage tick만 100으로 나눠 기존 fractional `default_stop_pct`로 저장한다(SHALL). 렌더링은 불필요한 0이나 부동소수 artifact가 없는 결정적 10진 표기여야 한다(SHALL). 기존 파일에 엔진은 허용하지만 콘솔 0.5% grid에는 맞지 않는 값이 있으면 정확한 현재값과 보정 필요를 표시하고, 저장 시 침묵 반올림하지 않고 거부한다(SHALL).

목록 직접 기입은 고급 접힘 안에만 두고, 거부된 블록은 파일 원문 값과 거부 사유를 함께 표시한다(SHALL). 저장은 다른 편입 필드를 보존하고 audit을 남기며, 반영 시점("다음 엔진 기동부터 반영"), 편입 비가역 귀결, 지정의 상시성, 현재 엔진 실행 여부를 안내한다(SHALL). 저장은 엔진을 재시작하거나 기존 관리 포지션을 변경하지 않는다(SHALL NOT). 이 백분율 control에는 client-side script나 CSP 예외가 없어야 한다(SHALL NOT — 같은 template의 다른 legacy confirm handler는 이 change 범위 밖이다). automation gate 편집 표면은 존재하지 않는다(SHALL NOT). 포지션 화면의 관리 외 보유 행은 기존 종목 편입 지정/해제 행위를 유지하고, 그 행위는 include 목록만 바꾸며 편입을 직접 수행하지 않는다(SHALL NOT).

#### Scenario: 상단 메뉴에서 외부 종목 자동관리 설정 열기
- **WHEN** 운영자가 어느 콘솔 화면에서든 상단의 "외부 종목 자동관리"를 선택하면
- **THEN** `/settings#adoption`이 열리고 자동 편입 ON/OFF·합성 보호폭·제외·지정 설정과 기존 정책 적용 설명이 첫 섹션에 표시된다

#### Scenario: 현재 비기본값 표시
- **WHEN** 저장된 adoption fraction이 `0.075`이면
- **THEN** 화면은 inline event handler나 CSP 예외 없이 편집 가능한 `7.5%`를 표시한다

#### Scenario: 유효한 비기본값 저장
- **WHEN** 운영자가 `default_stop_percent=7.5`를 제출하면
- **THEN** `default_stop_pct=0.075`가 저장·audit되고 다른 adoption 필드는 보존되며 실행 중 엔진에는 재시작이 필요하다고 안내한다

#### Scenario: 전체 허용 grid
- **WHEN** 2%부터 20%까지 0.5 percentage-point 단위의 어느 값이든 제출하면
- **THEN** 정확히 100으로 나눈 fraction이 저장되고 재로드한 config가 같은 fraction을 반환한다

#### Scenario: 잘못된 백분율
- **WHEN** empty·비수치·NaN·무한대·2% 미만·20% 초과·0.5 단위 밖 값 또는 과거 `0.05` fraction 형식을 제출하면
- **THEN** 구체적 안내와 함께 저장이 거부되고 adoption 블록은 변경되지 않는다

#### Scenario: 기존 grid 밖 값
- **WHEN** 기존 config에 `default_stop_pct=0.076`이 있으면
- **THEN** 화면은 `7.6%`를 정확히 표시하고 허용 단위로 보정해야 함을 안내하며, 그대로 저장해도 침묵 반올림 없이 거부한다

#### Scenario: deny-by-default CSP
- **WHEN** remote settings 응답이 배포 CSP와 함께 렌더되면
- **THEN** adoption percentage control에는 inline script가 없고 값은 표시·편집 가능하다

#### Scenario: 종목별 편입 지정과 해제
- **WHEN** 포지션 화면에서 관리 외 보유의 지정 또는 기존 지정의 해제를 수행하면
- **THEN** include 목록만 멱등 갱신되고 이미 편입된 포지션과 엔진 프로세스에는 영향이 없다

#### Scenario: seam 미배선 빌드
- **WHEN** 편입 설정 seam이 주입되지 않은 콘솔에서 설정 화면을 열면
- **THEN** 저장 불가와 그 이유가 안내되고 다른 화면은 정상 동작한다
