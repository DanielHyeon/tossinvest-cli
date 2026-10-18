## 1. 구현 준비

- [x] 1.1 선행 a149, a150, a153, a154 완료, Story↔change 연결, 설계 리뷰와 사용자 범위 승인을 확인한다.
- [x] 1.2 Terra 단일 writer를 배정하고 현 HEAD base, CodeGraph 호출 사슬과 필요한 FLM/BTM·Pre-Edit Gate를 고정한다.

## 2. 범위 구현과 검증

- [x] 2.1 design의 계약을 기존 모듈 재사용으로 최소 구현한다. 현재 운영 기본값과 기존 증거를 보존한다.
- [x] 2.2 주문·예약 writer 없는 구성, 모든 signal accept에서도 broker count0을 검증한다.
- [x] 2.3 disk failure/backpressure/duplicate/restart와 기존 exit 독립성, 보호 경로 지연 없음 검증을 수행한다.
- [x] 2.4 입력 불변성과 실주문·운영 변경·추가 수집 없음의 증거를 남긴다.

## 3. 완료 검증과 인계

- [x] 3.1 별도 reviewer의 독립 검토와 gstack code/security/QA 리뷰를 받고 발견사항을 해결한다.
- [x] 3.2 Terra가 대상 시험 결과와 최종 sdd-sync/check·집계 gate 실행 준비를 기록한다. 집계 gate 성공은 별도 실제 종료 영수증으로만 판정한다.
- [x] 3.3 Manager가 diff·범위·결과물을 독립 확인하고 미검증 사항을 분리한다.
- [x] 3.4 정식 spec sync/archive 영향, PM 경로 갱신과 범위 한정 커밋·푸시 계획을 준비한다. 아직 실행 완료로 표시하지 않는다.

## Post-gate lifecycle

모든 구현 task가 충족된 뒤 `make gate CHANGE=a155-observe-versioned-strategy-decisions`의 실제 성공과 독립 리뷰·Manager 최종 확인을 기록한다. 그 뒤 Terra가 정식 spec sync/archive·PM 경로 갱신·strict 검증 후 승인된 범위만 커밋·푸시한다. 체크박스 완료만으로 gate/commit/push/archive가 끝났다고 보고하지 않는다.
