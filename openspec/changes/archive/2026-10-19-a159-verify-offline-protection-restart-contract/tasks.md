## 1. 계약과 증거

- [x] 1.1 번호·Story·사용자 분리 승인과 a156/a157 제외 범위를 확인하고 설계 독립·gstack 검토를 해결한다.
- [x] 1.2 기준 commit, research-first 재사용 조사, 현재 CodeGraph/CGC 조정, 사전 AST와 test-only Pre-Edit 근거를 보존한다.

## 2. 오프라인 복합 회귀

- [x] 2.1 실제 코드의 체결 적용 경로로 동일 원장의 부분 체결·보호 차단·재개방·중복 재생과 상태 보존을 검증한다.
- [x] 2.2 유효한 새 후속 leg의 정확한 차단과 양성/범위 대조군을 검증한다.
- [x] 2.3 미종결 BUY·명시적 terminal·pending SELL·비손절 상태별 보호 판단·알림·제출 한계와 중복 방지를 검증한다.
- [x] 2.4 독립적인 안전 불변식 변이 두 가지의 실제 behavioral 실패와 원본 GREEN, focused/race 결과를 보존한다.
- [x] 2.5 생산 코드·기존 증거·운영 설정 무변경, 임시 DB·fake 전송만 사용한 범위를 확인한다.

## 3. 검토와 게이트 준비

- [x] 3.1 독립 리뷰와 gstack 리뷰 지적을 해결하고 Manager가 소스·결과·한계를 직접 확인한다.
- [x] 3.2 최종 sync/check·전체 gate 실행을 위한 고정 범위와 검증 계획을 정리한다. 실제 gate 성공은 별도 종료 영수증으로 판단한다.
- [x] 3.3 정식 spec sync/archive, PM 경로 갱신, strict 검사와 범위 한정 commit/push 계획을 준비한다.

## Post-gate lifecycle

위 작업 충족 후 `make gate CHANGE=a159-verify-offline-protection-restart-contract`의 실제 성공을 확인한다. Manager 검증 후 Terra가 정식 archive와 본 명세 동기화, PM 경로·파생 파일 갱신, strict/PM 검증을 수행하고 이 범위만 커밋·푸시한다. 원격 SHA와 격리 작업 트리 상태를 확인한다. a156의 실증·승인 작업 및 a157은 변경하지 않는다.
