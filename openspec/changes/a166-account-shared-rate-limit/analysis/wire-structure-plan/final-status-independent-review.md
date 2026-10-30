# a166 최종 상태 독립 문서 검토

대상 `/home/daniel/tossos-work/a166-wire-structure-plan` HEAD 9dab3dda44651f0a4d5422bec14af478a293065b 및 Manager 미커밋 tasks.md·analysis/development-status.md·wire-structure-plan/manager-verification.md. 읽기 전용 원문 대조, 시험·AST 재실행 없음.

## P2 — 4.1 완료와 runbook 초안의 용어
`tasks.md:41`의 4.1 체크는 provisioning·동일 inode/namespace·dirty 복구·rollback **절차 작성과 fixture 검증**의 완료를 뜻한다. 근거 `analysis/operator-cutover-draft.md:3`은 여전히 '개발용 절차 초안', '검증 완료 runbook이 아니다'라고 명시하며 `development-status.md`는 해당 문서를 'runbook'으로 부른다. 선택 fixture/독립검토 자체는 offline-budget-tool-lot/manager-verification.md에 존재한다. 절차의 작성 범위 완료와 운영용 runbook 승인·실환경 검증을 혼동하지 않도록 status 표를 '절차 초안 + fixture'로 쓰고, 4.1 체크가 초안 완료인지 최종 운영자 절차 완료인지 명확히 하라. 최종 절차 자체가 task의 필수라면 체크를 보류하는 것이 맞다. 4.2 사람 검증은 어느 경우에도 미완이다.

## 나머지 판단
- 2.1과 3.1은 미체크를 유지하며 a130/d2 상세 FLM/BTM 사전 작성 위반을 소급 준수로 바꾸지 않는다. `tools-outcome-lot/manager-verification.md:9`의 현재 제한 결과 수용 및 위반 기록과 일치한다.
- 2.4의 체크는 tasks 말미 및 `wire-structure-plan/manager-verification.md:7`이 83 shape/116 selector·수동 분류·지정 행동 시험의 한계와 임의 동적 경로·운영 참여자 미증명을 명시할 때만 제한된 source 범위로 해석된다. 전체 의미적 값 흐름/운영 총량 증명으로 읽지 않는다. consumer-integration.md의 실제 합성 normal/tagged race 및 a165/a166 회귀 3개와 일치한다.
- 2.2–2.3·2.5·3.2–3.5는 development-status 표의 로트별 선택 증거와 한계에 연결돼 있고, 3.4는 a087 nil 및 a169 미구현을 명시한다. 3.3의 '모든 429'는 observed required 참여자와 실패 처분이 확인된 범위로 해석해야 하며 운영 전체 보증이 아니다. 3.5 gstack은 로트별 수동 검토이지 자동 전체 pipeline이 아니다.
- 4.2–4.5는 미체크, a087 5.4/5.5·a169 reader/fence/abort 및 리허설/측정 미완, a166 active/archive 금지 유지. 전체 tests/gate·실브로커·운영·활성화 미실행을 일관되게 밝힌다. strict/PM/diff는 아직 실행 전이며 4.3 체크 불가.

P0/P1은 문서에서 추가로 발견하지 못했다. P2 해결은 검증 완료 runbook이라는 허위 주장 없이 상태 용어를 구체화하는 최소 정정으로 가능하다. 운영 승인이나 종결 승인 판정이 아니다.
