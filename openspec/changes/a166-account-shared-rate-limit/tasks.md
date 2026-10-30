> **4판 proposal-freeze READY — 개발 착수 (2026-10-31 Manager).** 사용자는 strict 3건/초·보호 대기 및 남은 개발 진행을 승인했다. Manager가 계약·완료를 판정하고 Terra가 구현·시험·gstack 리뷰·종결한다. 배포·실계좌·a087 활성화·a169 측정은 별도다. 2·3판 작업표는 `analysis/tasks-through-v3-history.md`의 역사 기록이며 현재 작업 지시가 아니다.

## 0. 설계와 연구

- [x] 0.1 사용자 정책·역할·개발 진행 승인 반영. UI 없음: impeccable N/A — backend/documentation-only.
- [x] 0.2 기존 검증 도구 복구: OpenSpec 1.4.1 임시 pin, e32606d4b의 a087/a166/a169 strict valid. 원 설치 버전으로 단정하지 않음.
- [x] 0.3 TLS 수명 연구: actual official.Client token/401, full response, context, multi-process·kill, TLS 검증 및 변이. Linux 연구 SHA d0485309a. 핵심 b90524af6 독립 Go1.25.0/1.26.5 race 통과, 추가 owner counter 1b5f09589 Go1.25 통과. 연구 경계와 제품 검증을 구별함.
- [x] 0.4 queue 연구: liveness·priority·FIFO·취소·dirty marker·fault 주입·aux cleanup, targeted race와 변이. Linux 연구 SHA 41c8c16a7. 445419bc1 독립 race 통과, 최신 후속 독립 확인은 다음 항목.
- [x] 0.5 최신 연구 후속 diff의 독립 확인. queue 41c8c16a7 독립 race 6.152초, TLS 최종 build-tag diff 확인. `analysis/v4-independent/`의 원문·AST·CodeGraph 교차 확인 기록 보존. 제품 전체 증거와 구별.
- [x] 0.6 proposal/design/spec delta/tasks 정합화, research-first 버전/출처/라이선스·도구 제한, scoped strict validation·PM check. 근거: `analysis/manager-review-v4.md`, OpenSpec 1.4.1 scoped a087/a166 valid 및 PM check.
- [x] 0.7 독립 설계 리뷰 및 Manager READY. P1 dirty sync 모순 해소, 추가 P0/P1 발견 없음; 최신 연구 독립 증거·research-first·scoped strict·PM·diff check eef74a163 확인. 승인된 보호 대기 정책과 단계별 제품 검증을 구별하며 2판 값으로 계약을 추정하지 않음. Story의 옛 무지연 문구도 최신 정책에 맞춤.

## 1. 내부 모듈 — 배선하지 않는 첫 로트

- [x] 1.1 source 기준·영향 지도·CodeGraph/AST·Pre-Edit 기록. 기존 함수 변경이 없으면 FLM 해당 없음 사유; 기존 함수 편집 시 사전 FLM/BTM.
- [x] 1.2 RED: queue 우선/취소/사망·dirty·파일 검증·동일 scope, 전체 요청 수명과 local request start 간격, OFF와 required 오류 분리.
- [x] 1.3 GREEN: internal/accountbudget 신규 모듈. caller context, 주입 clock, 소유권·취소·응답·이벤트 오류 계약. 기존 실행 배선/운영 기본값 무변경.
- [x] 1.4 targeted normal/race·다중 프로세스·변이·gstack 리뷰·독립 검토. 연구 복사를 제품 검증으로 대체하지 않음.

첫 로트 판정: `analysis/core-lot1/manager-verification.md`. OFF 기존 실행은 미배선 상태로 보존했으며 공식 클라이언트 OFF golden·전체 배선·429/운영 검증은 후속 로트다. CodeGraph 미초기화, 실제 namespace·live·전체 테스트/gate는 미검증이다.

## 2. 공식 클라이언트와 별도 전송

- [ ] 2.1 official constructor/authority/doRequest 및 token 전달 경계, a112 clone, replay/probe 기존 함수 FLM/BTM·Pre-Edit.
  - 현재 함수 증거와 구현 검증은 로트별 완료했지만 a130/d2 상세 FLM/BTM의 사전 작성 위반은 소급 준수로 바꿀 수 없다. `analysis/tools-outcome-lot/manager-verification.md`에 예외 수용과 미검증 범위를 보존한다. 따라서 이 절차 항목은 완전 준수로 체크하지 않는다.
- [x] 2.2 RED/GREEN: private constructor-owned guard, exact OFF 보존, 외부 client override authority 거절, actual token/401/redirect별 중재, response TLS/body/timeout.
- [x] 2.3 RED/GREEN: 보내기 전 admission 거절과 기존 attempt 결과 구분, a165 첫 AMBIGUOUS 보존 및 journal/멱등키 회귀.
- [x] 2.4 별도 공식 API 전송 경로를 같은 예산에 결속. WTS/judge 분류와 지원 범위 명시. 전체 배선 구조 검사.
- [x] 2.5 targeted normal/tag/race·mutation·gstack·독립 리뷰·Manager 검증.

## 3. 호출자·폴링·설정

- [ ] 3.1 사전 FLM/BTM과 RED: critical 의존 사슬, normal/auxiliary, 모든 config/profile의 단일 공통 group 결속. trusted launch manifest의 예상 root/queue/wire device·inode/instance 검증, marker 복사 root와 required 누락은 bootstrap wire0. 여러 credentials/accounts도 그룹 총량 안에서 기존 계좌 선택을 유지한다(`analysis/budget-group-binding.md`).
- [x] 3.2 GREEN: 앱·엔진·flatten·CLI·console·soak·verify·프로브 배선. 실제 운영 env/toggle 변경 없음.
- [x] 3.3 RED/GREEN: 모든 429 이벤트, 엔진 폴링 같은 조회 재시도 금지, 유효 ≤30s hint 범위·복원·invalid/초과 처리. 비지연 walk와 POST 판정 보존.
- [x] 3.4 retry matrix와 코드 동시 정합화. a087·a169 후속 계약과 통합 시험 SHA 목록 작성. a087 production admission nil 유지.
  - a166 인계 범위는 동일 검증 Budget의 read-only snapshot API와 실제 official/PriceLimits auxiliary·token·401·wire 수명 증거다. a087 board/epoch/429 정지의 실제 연결은 D16/5.5, a169 event 독자/prewire fence/영속 abort/엔진 신원은 a169 후속 작업으로 남긴다. API만 제공하고 그 소비자 구현을 완료로 체크하지 않는다.
- [x] 3.5 영향 범위 normal/tag/race·변이·gstack 적대/보안 리뷰, Manager 독립 검증. 전체 미실행 범위 기록.

개발 증거 교차표: `analysis/development-status.md`. 3.1의 manifest/group·우선순위 fixture 및 현재 함수 증거는 완료했지만, 2.1의 사전 작성 예외 때문에 포괄적 절차 준수 체크는 보류한다. 2.4는 문서화한 source 범위의 분류·형태 검사와 로트별 행동 시험이며 모든 동적 호출 또는 운영 중 참여자 증명이 아니다.

## 4. 종결과 운영 경계

- [ ] 4.1 운영자 provisioning·동일 inode/namespace 확인·dirty 복구·rollback 절차 작성. 라이브 접근 없이 fixture 검증.
  - offline provision/check와 fixture는 완료. operator-cutover-draft는 개발용 초안이며 최종 운영 절차 검증으로 표시하지 않는다. 실환경 결속 확인과 함께 마무리한다.
- [ ] 4.2 배포 전 사람 확인: 실제 호스트/컨테이너 namespace·UID·filesystem·동일 계좌 scope와 모든 참여자 포함.
- [x] 4.3 변경 범위 검증·strict OpenSpec·PM 최종 확인. 전체 시험/전체 gate는 해당 작업의 명시 요청 때만 실행.
- [ ] 4.4 Manager 완료 판정 뒤 Terra commit/merge/push/archive·Story/PM 정합화. 사람 단계 미완료를 완료로 체크하거나 active change를 허위 archive하지 않음.
- [ ] 4.5 운영 이미지/설정 적용은 별도 사람 승인. a087 5.4/5.5, a169 리허설·측정은 자동 실행하지 않음.

현재 종결 판정: 코드 로트는 선택 검증 뒤 게시한다. 4.2 실환경 사람 확인 미완료로 active change를 유지하고 archive하지 않는다. 4.3은 해당 change strict·PM tracker·diff 범위 검사만 확인했다. 전체 시험·전체 gate는 미실행이다.
