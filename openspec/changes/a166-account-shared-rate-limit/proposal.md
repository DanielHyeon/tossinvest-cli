# a166 — 계좌 공유 요청 상한과 폴링 429 처리

> 4판 proposal-freeze 후보, Manager READY 대기. 사용자 승인: 계좌 합산 3건/초 우선·보호 요청 대기 허용, 남은 개발 진행. Manager는 OpenSpec/완료 판정, Terra는 구현/시험/gstack/종결. 운영 활성화·배포·실계좌는 별도다.

## Why

엔진·CLI·그림자 도구의 개별 요청 제한으로는 같은 계좌의 합산 요청률을 제어할 수 없다. 기존 확정 표 `openspec/changes/verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md` §2는 계좌 공유 3 req/s와 엔진 폴링의 같은 조회 내 429 즉시 재시도 금지를 정했다. a087 하한가 갱신과 a169 측정도 이 공동 예산을 필요로 한다.

이 값은 사용자 정책이며 브로커가 공표한 한도 또는 서버 도착률의 실측 보장이 아니다. 현재 브로커 한도 집계 범위가 계좌·키·IP 중 무엇인지 확인하지 못했다. 선행 설계의 보호 무제한 우회와 예약시각만으로 wire 간격을 증명하는 방식은 폐기한다.

## What Changes

- enforcing 모드에서 동일 호스트·계좌의 모든 참여 공식 API 요청을 같은 파일 잠금과 대기열로 중재한다. 직접 TLS의 HTTP 요청 첫 plaintext Write를 시작으로 세어 rolling 1초에 합계 3건 이하를 지킨다. 브로커 도착 시각은 제어하지 못한다.
- 보호 조회·매도·취소·flatten·해당 token/접수 확인 사슬은 대기열 우선순위만 가진다. 이미 선택된 소유자를 추월하거나 공유 실패 시 로컬로 우회하지 않는다. 유한 wall-time 지연 상한은 약속하지 않는다.
- context 취소/시한, 응답 body 및 실제 writer의 종료를 포함해 소유권을 관리한다. 새 요청의 미전송과 앞선 주문 attempt 결과를 구분한다.
- 모든 429를 응답 단위의 공유 이벤트로 남기고 엔진 폴링의 같은 조회 재시도를 막는다. 유효 ≤30s hint는 지정된 엔진 폴링 범위에서 존중하며, 새 계좌 전역 cooldown은 도입하지 않는다. 비지연 walk와 주문 원장 판정은 유지한다.
- OFF는 기존 transport·authority·가격·요청 동작을 보존한다. enforcing 필수인데 예산이 누락/무효인 경우는 OFF로 처리하지 않는다.
- 기존 default 클라이언트, token 교환뿐 아니라 replay·a112 측정 clone·a087 probe의 직접 HTTP 경로도 포함한다. WTS/judge는 별도 분류하며 동일 예산이라고 가정하지 않는다.

## 범위와 순서

내부 모듈 → 공식 클라이언트/권한/별도 전송 → 호출자/폴링/설정 순으로 구현한다. 각 로트는 격리된 targeted RED/GREEN/race·변이와 gstack 리뷰 뒤 Manager가 검증한다. 연구 코드는 선택 근거이며 제품 코드나 생산 전체 검증이 아니다. 기능·배포 검증을 앞당겨 완료로 표시하지 않는다.

제약: Linux의 지원 filesystem/time namespace, 안정적인 동일 inode, 협조적인 참여 프로세스. 앱 수동 주문·다른 호스트·외부 프로그램·관리자의 runtime 파일 교체는 이 중재가 통제하지 못한다. 실제 활성화 전 지원 조건과 계좌 scope, 모든 참여자의 연결을 확인한다. 외부 전송 경합이 남아 있으면 계좌 전체 보장을 주장하지 않는다.

비목표: 새로운 매매 전략, 자동 배포, 실계좌 부하 시험, 승인 없는 신규 의존성, 운영 토글 변경, a087 활성화 및 a169 측정. a087 admission nil을 유지한다.

## Capabilities

### Modified Capabilities

- `order-execution`: 요청 재시도 정책과 공유 admission. 가격·수량·멱등키·a165 첫 AMBIGUOUS 보존은 유지한다.

## Impact

예상 대상: 신규 `internal/accountbudget`, `internal/official` constructor/authority/전송, execgw 재시도/재생 경계, 엔진/CLI/도구 생성자와 설정, 내부 clock의 공유 시간 표본, 정본 retry matrix. 기존 함수는 편집 전 현재 AST·FLM/BTM·Pre-Edit를 수행한다. 원장 schema 변경은 계획하지 않는다.

High-risk. 현재 문서는 검토 후보이며 READY 전 제품 편집을 허용하지 않는다. UI 영향 없음: impeccable N/A — backend/documentation-only. 연구는 기존 flock/Go stdlib와 설치된 x/sys를 우선 검토했고 런타임 의존성 추가는 없다. 과거 proposal은 `analysis/proposal-through-v3-history.md`에 보존한다.
