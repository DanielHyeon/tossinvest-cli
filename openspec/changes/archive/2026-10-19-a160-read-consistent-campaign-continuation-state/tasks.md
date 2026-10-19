## 1. 구현 준비

- [x] 1.1 Story/registry, 기존 분할 계약과 승인된 비권한 범위, 독립·gstack 설계 검토를 확인한다.
- [x] 1.2 base commit, research-first/기억 조회 한계, 현재 AST/그래프 조정과 Pre-Edit 증거를 보존하고 사전 sync를 완료한다.

## 2. 일관된 읽기와 검증

- [x] 2.1 ReadOnly 전용 scoped snapshot을 기존 scanner와 단일 read transaction으로 구현한다. full plan unavailable을 명시하고 부분 성공·권한 생성을 배제한다.
- [x] 2.2 KR/US 정상·peer·범위/세대/레거시/상태/결손·orphan·오류 및 blocked 관측을 검증한다.
- [x] 2.3 결정적 동시 writer snapshot, close/reopen, logical row 무변경, focused/race와 behavioral fault/변이 증거를 확보한다.
- [x] 2.4 schema·기존 writer·운영 설정·기존 증거·a156/a157 및 원본 checkout 무변경을 확인한다.

## 3. 검토 및 완료 준비

- [x] 3.1 독립·gstack 구현 리뷰 지적을 해결하고 Manager가 실제 diff·source hash·시험 결과 및 한계를 검증한다.
- [x] 3.2 최종 문서/PM 고정과 sync/check/전체 gate 계획을 확정한다. 실제 성공은 별도 종료 기록으로 판단한다.
- [x] 3.3 정식 archive/spec sync·PM 경로 동기화·strict 및 범위 한정 commit/push 절차를 준비한다.

## Post-gate lifecycle

실제 `make gate CHANGE=a160-read-consistent-campaign-continuation-state` 성공과 Manager 검증 뒤 Terra가 정식 archive(본 명세 동기화 포함), PM 경로/파생 파일 갱신, strict/PM 검증, 범위 한정 commit 및 origin main 정상 push를 수행한다. 실패 시 기록을 보존하며 force-push하지 않는다. read-only 단계 완료를 a156 후속 주문 또는 live 완료로 표현하지 않는다.
