## 1. 설계 및 증거 준비

- [x] 1.1 Story·spec·설계 범위와 독립·gstack 설계 리뷰를 확정하고 strict/PM 검사를 통과한다.
- [x] 1.2 base commit, research-first·memory 한계, CodeGraph/AST 근거와 pre-edit 범위를 보존한다.

## 2. 오프라인 보존 및 복원

- [x] 2.1 KR·US × 세 lane의 계획·위험 원형을 단방향 detached capture로 구현하고 실제 RED→GREEN 증거를 남긴다.
- [x] 2.2 bounded canonical codec과 외부 content digest·scope 결속을 구현한다. 버전·손상·중복키·결측·초과·잘린 입력을 거절한다.
- [x] 2.3 nonzero/known-zero/unknown/overbudget·false latch·fill/cancel 원형·weekly 의미를 보존하고 임시 파일 close/reopen·deep-copy·기존 digest를 검증한다.
- [x] 2.4 외부 입력이 native 실행 객체·FX 인장·risk authority를 만들지 않고 기존 실행·a160·원장·증거·운영/a157을 변경하지 않음을 검증한다. 실제 behavioral mutation/fault 반례와 focused/race 결과를 남긴다.

## 3. 검토 및 완료 준비

- [x] 3.1 독립·gstack 구현 리뷰 지적을 해결하고 Manager가 source/diff/원본 시험·해시·한계를 검증한다.
- [x] 3.2 metadata·PM을 고정하고 최종 sync/check/전체 gate 실행을 준비한다. 성공 판정은 실제 종료 기록으로 한다.
- [x] 3.3 formal archive/main spec sync·PM 경로·strict·범위 한정 commit/push 절차를 준비한다.

## Post-gate lifecycle

실제 전체 gate PASS와 Manager 검증 후에만 Terra가 정식 archive(본 spec 동기화 포함), PM 갱신·검사, strict 검증과 범위 한정 commit 및 origin main 정상 push를 수행한다. 모든 실패·수정·리뷰 이력을 보존하고 force push하지 않는다. 오프라인 검사 복원을 실계좌 현재 위험 복원이나 주문 재개 완료로 보고하지 않는다.
