# a169 U2 합집합 계획 독립 검토

대상 Manager 미커밋 `analysis/u2-union-plan.md`, score worktree HEAD 14fc8e2af. **bounded READY**, P0/P1 없음. 순수 계산 계획만, source/AST/시험/운영 미실행; 선행 두 로트 검증·게시와 Manager 착수 판정은 별도다.

계획 :7–21은 정본 design:136–142의 각 필터 통과 심볼 ANY-pass 합집합, 원본 volume 존재 시 V 우선, :168–169 raw 순위 tie, 선정 시각의 FetchedAt max 및 ratio 정렬과 일치한다. a169Select의 전체 Rows에서 Reason 빈 행을 사용하므로 Symbols의 N으로 후보를 자르지 않는다. 31위 이상 후보, 필터 실패한 volume raw V 우선, 2중 receipt 보존, copy/정렬 결정성, missing vs empty daily 및 시각 오류의 수용 fixture도 제시됐다. `design:175` 단일 last_price의 원장 projection 모호함은 이 로트에서 선택하지 않고 후속으로 남긴다.

P2 제안: invalid raw V와 daily map 누락이 같은 후보에서 동시에 발생할 때 `numerator_invalid`과 `baseline_unavailable` 중 어느 사유가 우선하는지 :13·19가 명시하지 않는다. 선행 score의 numerator 우선과 일관되게 우선순위를 명시하고 literal fixture로 고정하면 결과 재현성과 사유 비교가 명확해진다. 이 결정은 필터 합집합/순위를 바꾸지 않는 구현 상세다. score의 중복날짜/안전정수 등 지원 제한은 사용자 정책으로 오인하지 않는다.

두 ranking provenance/조기 제외·deadline/late·원장·profile/event fence·운영은 후속, 전체 a169 freeze/2.2/측정/archive 완료 아님.
