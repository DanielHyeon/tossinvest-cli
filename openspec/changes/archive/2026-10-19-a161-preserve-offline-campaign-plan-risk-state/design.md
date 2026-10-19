## Context

기준은 `67810038f21dffd519a3d44edb9e255238b5b74a`이다. a160의 일관된 원장 조회는 전체 계획 원문을 갖지 않는다. 사용자 추가 결정은 **초기 계획·예산과 현재 위험액 복원을 함께 설계**하는 것이다. 이번 구현은 세 lane에서 제공된 캡처 시점 상태의 보존·검사 복원이다. 과거 원장 역산이나 실계좌 최신 위험액 확정은 아니다.

사전 source/Go AST/CodeGraph 자료는 `analysis/code-context/`에 보존한다. continuation/reversal/weekly의 native 계획 digest와 FX 표현이 다르므로 기존 digest 알고리즘을 바꾸거나 공통 digest로 대체하지 않는다. weekly RiskState의 fill Known은 식별자 여부이며 실제 위험액 정확성 인증이 아니다. weekly 취소 receipts는 별도 ReservationState에 있으므로 없는 cancels map을 만들어 채우지 않는다.

## Goals / Non-Goals

**Goals:** KR·US × 세 lane의 전체 계획 원문·초기 예산·per-share risk·배분 상한·통화·FX 자료, native RiskState의 filled/held/latches/fill 및 지원되는 cancel 기록을 손실 없이 보존한다. 캡처 후 원본 map/FX 변경과 독립된 바이트를 만들고 파일 close/reopen 후 detached 검사 자료로 복원한다. 알려진 0과 결측·unknown·초과를 구별한다.

**Non-Goals:** engine/CLI/API/UI 연결, 원장 migration/수정, 최초 admission 원자 저장, 현재 원장과 일치하는 authoritative risk 재구성, native 거래용 plan/FX/RiskState 재생성, weekly ReservationState/예약 횟수 재생성, 새 risk cap/qFinal/lease, a156 완료, a157 수정. 파일 경로 관리·daemon·DB·lock·fsync framework·power-loss durability는 만들거나 주장하지 않는다.

## Decisions

1. **단방향 export + 작은 공통 codec.** 각 lane 패키지의 새 파일에서 기존 native 계획 검증과 필요한 risk binding/형식 확인 후 자기 소유 필드 전체를 detached 자료로 옮긴다. 공통 leaf package는 stdlib만 사용하며 lane 패키지나 journal을 import하지 않는다. 기존 생성자·평가·위험전이 함수는 변경하지 않는다. 새 코드의 운영 호출자는 0이어야 한다. 코드 구조에 맞는 구체 API/자료형 이름은 Terra가 정하되 공개 복원 반환 타입은 native 거래 타입을 포함할 수 없다.
2. **원형 필드 보존.** 계획 scope는 account/market/symbol/campaign/lane/version/generation, 계획 수량·모든 ceiling·고정 배분·예산·통화 단위·policy/config digest·native plan digest를 포함한다. FX의 원문 시각·rate·haircut·출처·원래 integrity 자료는 opaque data로 보존하며 mint나 최신성 갱신을 하지 않는다. 주간과 reversal의 private 필드는 소유 package 내에서 단방향 복사한다. 기존 lineage budget digest와 native plan digest를 혼동하지 않는다.
3. **위험 상태를 예산과 분리.** filled/held는 캡처된 수치 그대로, latches는 false 값까지, fill/cancel/conflict/missing-ID keys와 모든 native record 필드를 그대로 보존한다. 금액은 정밀한 십진 정수 문자열이며 float 변환·합산 후 소거·예산으로 clamp·NewRiskState 초기화 금지. unknown actual risk에서 값이 남아 있으면 그 값은 recorded value일 뿐 exact actual risk가 아니다. unknown latch와 lane별 Known 필드 의미를 그대로 노출한다. overbudget 상태도 검사 가능해야 하며 decode가 차단을 해제하지 않는다. weekly는 cancel record가 RiskState에 없다는 사실과 ReservationState 미포함을 기계 판독 가능한 필드로 표시한다.
4. **비권한 계약을 데이터에 포함.** schema v1, family, capture timestamp, `inspection-only`, `caller-provided-snapshot`, `ledger-currentness-unverified`를 필수로 둔다. native state의 구조 검증·artifact checksum 검증·실제 원장 출처 검증은 별개다. 첫 둘만 구현한다. 공개 변경 가능한 continuation/reversal 입력과 weekly seal 모두 실계좌 완전성을 입증하지 않는다. 호출자는 export 중 입력을 변경하지 않도록 동기화해야 하며 concurrent map mutation 지원을 주장하지 않는다.
5. **외부 기대값에 결속된 엄격한 복원.** codec은 최대 4 MiB와 schema v1만 허용한다. 저장물 밖에서 전달한 SHA-256 content digest 및 account/market/symbol/campaign/lane/version/generation 기대값과 정확히 일치해야 한다. 이 기대값도 신뢰된 원장 출처를 대체하지 않는다. deterministic canonical JSON과 표준 typed decoder를 재사용한다. unknown fields, duplicate keys, trailing input, 생략된 필수 필드, 잘못된 버전·자료형·금액·scope·plan/risk binding·잘린/과대 입력은 usable partial result 없이 실패한다. canonical 재인코딩 동일성 비교로 중복/생략/비정규 표현을 거절할 수 있으며 별도 일반 JSON parser framework는 불필요하다. 원래 digest는 보존하고 artifact content digest는 별도 계산한다.
6. **저장은 caller-owned.** codec은 bytes 또는 제한된 reader/writer만 다룬다. 테스트는 Go stdlib의 신규 임시 파일을 저장·close·reopen한다. 파일 경로 덮어쓰기·원장 백업·전원 장애 복구를 담당하는 새 저장 서비스를 만들지 않는다. io.Writer API를 제공하면 short write와 에러를 전파한다. 기존 운영 파일이나 증거에 실제 쓰지 않는다. 공용 mutable DTO가 있다면 다시 encoding할 때 검증하고, 어떤 경우에도 native 권한 객체로 변환하지 않는다.

## Research / Skill Decisions

ai-memory-retrieval과 search-first 적용. 현재 tool catalog에 ai-memory 조회가 없으므로 성공한 회수로 기록하지 않는다. native memory의 구현과 권한 구분을 현재 소스로 대조했다. 세 lane 계획/risk와 a160, strict JSON 예제, strategyobserve 저장 코드를 조사했다. 기존 store는 diagnostic schema에 결속돼 있어 재사용하면 범위를 키우므로 채택하지 않는다. stdlib encoding/json·crypto/sha256·io와 기존 lane validation을 확장한다. 신규 외부 dependency·라이선스 변경 없음. 공개 API/브로커 변경이 없어 외부 API 조사는 필요하지 않다.

grill-me/grilling: 사용자 결정으로 현재 위험액까지 포함, 운영 연결은 제외했다. 공통 Known 인증 의미·seal remint·legacy zero 복원은 배제했다. impeccable: N/A — backend/offline library only, UI 없음. Manager 직접 설계, Terra 단일 작성자, 독립·gstack 리뷰와 Manager source/receipt 최종검증을 분리한다.

## Risks / Trade-offs

저장 snapshot을 최신 원장으로 오인 → status와 timestamp 및 미검증 표기를 반환. 원문 없는 기존 원장을 복구했다고 오인 → 원장 API 변경 없음, a160 unavailable 유지. hash를 인증으로 오인 → checksum 목적과 외부 기대값 출처 한계 명시. native 상태 조작·낡은 FX → 구조 검증과 실제 권위를 분리, 재생성하지 않음. unknown/false/빈 record 소실 → 원형 roundtrip 반례 시험. 새 보안 경계 → 버전/크기/형식/내용·scope 결속 음성 시험.

## Migration Plan / Open Questions

운영 migration·restart·설정·활성화 없음. 신규 library만 추가하고 rollback은 이 코드·문서 커밋의 되돌림이며 역사 자료는 유지한다. 기존 함수 내부 편집이 필요하면 먼저 Manager 검토 및 해당 AST/FLM/BTM을 마련한다. 테스트용 export/허용목록 수정도 변경 이유를 먼저 확인한다. 현재 공개 범위의 미정 정책 결정은 없으며 수익성·실거래 복원 능력은 미검증이다.
