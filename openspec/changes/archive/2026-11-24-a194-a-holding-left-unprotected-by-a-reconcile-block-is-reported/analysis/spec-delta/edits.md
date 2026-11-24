# a194 MODIFIED delta — 의도한 편집 목록 (Manager 결정 2026-11-23 Q5)

생성: `python3 openspec/changes/a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported/analysis/spec-delta/make_modified.py`(저장소 루트, HEAD 064f59fd9 의 정본에서 블록을 떠서 정확한 문자열 치환만 함 — 각 원문이 블록 안에 정확히 1 회 나오는지 단언).
산출: `exit-policy.modified.md`·`engine-safety.modified.md`(그대로 `specs/<cap>/spec.md` 의 `## MODIFIED Requirements` 아래에 붙임), 정본 대조 `exit-policy.diff`·`engine-safety.diff`.

| # | 정본 Requirement | 편집 | 종류 |
|---|---|---|---|
| E1 | exit-policy 「외부 취득 포지션의 자동 편입」 | 알림 문단 「전이 상태(RECONCILE 대기·인터록 미검증·Stabiliser 미수렴)만 무알림이다.」 뒤에 「RECONCILE 대기가 전이 상태인 범위는 「대사 차단으로 편입되지 못한 보유의 침묵은 전이 상태에 한한다」가 정한다.」 한 문장 추가 | 문장 추가 |
| E2 | 같은 Requirement, Scenario 「RECONCILE 중 편입 시도」 | WHEN 「심볼이 RECONCILE 상태일 때」 → 「심볼이 스스로 풀리는 RECONCILE 상태이고 유예(10분) 안일 때」 | WHEN 좁힘 |
| S1 | engine-safety 「무관리 보유 보고의 등급은 사실이 정한다」 | 첫 문단(critical 조건 집합) 끝에 「전이 상태가 아닌 대사 차단으로 편입되지 못한 편입 후보」도 critical 이라는 문장 추가 | 문장 추가 |
| S2 | 같은 Requirement | 끝에 Scenario 「운영자만 풀 수 있는 대사 차단 아래의 편입 후보」 추가 | Scenario 추가 |
| E5 | 같은 Requirement | 「critical 무관리 보고의 본문은 **편입 시도가 실패한 시각의 사건**으로 …」 문장을 「시도 실패 보고는 실패 시각, 대사 차단 보고는 관측 시각」으로 가름(Manager 결정 2026-11-23, R2-3(a)) | 문장 수정 |

검증(2026-11-23): 생성물에서 편집을 거꾸로 되돌리면 정본 블록과 바이트가 같다 — exit-policy True, engine-safety True. diff hunk 는 exit-policy 2(E1·E2) · engine-safety 3(S1·E5·S2)이고 그 밖의 글자 차이 0(E5 추가 뒤 재생성·재검증, 2026-11-23).
