# a206 review

## 1단계 경량 리뷰 r1 (2026-11-27, 1회 — 상한 2회 중 1회 사용)

- 리뷰어: Codex CLI 0.165.0(`codex exec --sandbox read-only`, 작성자 Claude 와 다른 계열). 범위: proposal · design · tasks · spec 델타 ·
  claims · evidence-reconciliation · a206_measure.py · Story. 지시: 읽기 전용, 저널 · DB 접근 금지, 새 가드 · 기능 제안 금지.
- 질문 4: (1) 닫힌 포지션 행의 스냅숏을 다시 읽는 생산 리더 누락, (2) C1~C7 의 file:line · 분기 정확성,
  (3) 제안 정비가 깨는 불변식 · 유일 인덱스 · 리더, (4) 측정 스크립트가 D1 과 같은 대상을 고르는지 · 원본에 안전한지.

| # | 등급 | 지적 | 처분 |
|---|---|---|---|
| 1 | — | 리더 누락 없음(SQL · CTE 검색) | 확인 |
| 2 | P3 | C2 는 `full()` 이 20 열 중 **필수 부분집합**만 요구함을, C6 은 Saved 가 없으면 `encode` 가 NULL 을 돌려줌(:803-804)을 빠뜨림 | 수리 — claims C2 · C6 문구 정정 |
| 3 | — | 불변식 깨짐 없음: completed 거절이 decision 중복 검사보다 앞(exit_state.go:463 · :494), 유일 인덱스는 NULL 제외(exit_snapshot.go:61), trade outcome · backfill · provenance · a166 replay 가 쓰는 열은 남음 | 확인 |
| 4 | P2 | 스크립트 대상 술어가 D1 과 다름(`exit_states.completed=1` 누락, 확정 PLACE 만 제외) | 수리 — 술어를 D1 그대로(CLOSED + completed=1 + broker_order_id 가진 attempt 전부 제외)로 바꾸고 재측정. S1 · S2 · S3 수치 불변 |
| 5 | P2 | 원본이 작업 디렉터리 안이면 시나리오 사본이 덮어씀 · 지움, `refuse_live` 가 운영 디렉터리 자체를 허용 | 수리 — 두 경우 모두 거절(측정 하네스 안의 안전 검사이며 생산 가드 아님) |
| 6 | P2 | S4 가 effective 를 먼저 비워 `lag(effective)` 가 NULL 이 됨 → saved 중복 제거가 측정에서 빠짐 | 수리 — 순서를 바꿔 재측정: 203,235,328 → **175,517,696** B |
| 7 | P2 | design 의 11월 건수(172 · 3)가 측정 출력(86 · 1)과 다름 | 수리 — 원인: 1차 ad-hoc 집계가 무장 행과 같은 intent 를 가리키는 PROPOSAL_REFUSED 훅 행까지 세어 두 배가 됨. 측정 스크립트에 11월 일별 무장 행 집계를 넣고 그 출력(10 · 64 · 12 · 1)으로 정정 |

- 재측정: `analysis/a206_measure.out.txt`(05:3x Z), 리더 확인 `analysis/a206_reader_check.out.txt` 도 수리된 술어로 재실행
  (사본 시점 CLOSED 129 · OPEN 1 — 그 사이 포지션 하나가 닫힘; S1 오류 124/129, S2 오류 0).
- 2회차는 열지 않음(P1 0, P2 넷 모두 측정 · 문구 수리로 닫힘).

## Function Logic Map

Function Logic Map: not-applicable — 1단계는 기존 Go 함수 본문을 편집하지 않음(생산 · 시험 diff 0). 분기 주장은 FLM 번들이 아니라
`analysis/ast-evidence/`(logic-map AST 11 개)를 근거로 씀. 2단계는 새 파일 · 새 함수만 목표(design D5).

## 변이 원장 · 다중 적대 리뷰

not-applicable — 1단계는 문서 · 측정 하네스뿐(비례 원칙). 2단계 구현은 원장 쓰기라 High-risk 이며 그때 정함.
