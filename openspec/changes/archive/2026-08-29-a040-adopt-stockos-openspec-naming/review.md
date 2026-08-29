# Review: a040-adopt-stockos-openspec-naming

- Date: 2026-07-31
- Scope: PM hierarchy, OpenSpec naming policy, validation tooling, generated trackers
- Review class: tooling/documentation lightweight review
- Voices: Manager self-review, adversarial PM consistency review

## Findings

1. 기존 Story 001~039와 비번호 change를 일괄 변경하면 archive 링크와 운영 문서가 깨질 수 있다.
   - 결정: `a040`을 명시적 cutoff로 두고 legacy ID는 유지한다.
2. Story에서 change만 확인하면 Story가 없는 번호형 change가 생길 수 있다.
   - 결정: Story → change와 change → Story를 모두 검사한다.
3. 번호만 비교하면 `a040_FOO`, 대문자, 중복 `a040-*`가 통과할 수 있다.
   - 결정: lowercase kebab-case 정규식과 저장소 전체 번호 중복 검사를 함께 적용한다.
4. StockOS prefix를 그대로 복사하면 TossOS PM 계층과 충돌한다.
   - 결정: change 규칙은 공유하되 Story prefix는 `STORY-TOS-`를 유지한다.
5. 형식 검사만으로는 `STORY-TOS-040` 또는 `STORY-TOS-a001`처럼 cutoff를 우회할 수 있다.
   - 결정: legacy 숫자 Story는 040 이상을 거부하고 번호형 Story/change는 a040 미만을 거부한다.

## Verification evidence

- PM unit tests: 15/15 pass.
- Strict OpenSpec validation: `a040` through `a051` pass.
- Generated tracker check: current after regeneration.
- Product runtime impact: none.

## Function Logic Map

Function Logic Map: not-applicable

수정 대상은 Python PM 검증기이며 `check_analysis.py`가 추적하는 기존 Go 함수가 아니다. 거래·주문·위험·원장 런타임 경로도 변경하지 않는다. 번호 일치, 잘못된 slug, 중복 번호, legacy 회귀는 `tools/pm/test_generate_master_tracker.py`의 분기 테스트로 검증한다.

## Process note

PM 검증기 RED/GREEN 작업은 이 리뷰 파일 작성보다 먼저 수행되었다. 제품 구현은 시작하지 않았고, a041~a051은 각각 독립 proposal-freeze review 전까지 구현 금지 상태다. 다음 change부터는 review 기록을 구현 task보다 먼저 고정한다.

## Verdict

명명 정책과 PM 검증 변경은 수용한다. 제품 기능 change의 구현 승인이나 LIVE 운영 토글 승인을 포함하지 않는다.
