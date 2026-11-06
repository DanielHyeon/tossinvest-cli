# a169 daily session 독립 설계 리뷰

- 대상: `analysis/daily-session-plan.md` 2026-11-06 판, 기준 `232d6b86c024dfd71cc8ffe6fd0fbba342b3e9ec`.
- 판정: **조건부 설계 통과 — P0/P1 발견 없음**. 미배선 저장 컴포넌트의 설계에 한정한다. 구현·시험·실제 요청 차단·운영·측정·a169 전체 완료 판정이 아니다.
- 방법: 설계 원문, 기존 Q2 proposal/design 및 연구 보고서 대조. 소스 변경, 테스트, 운영 접근 없음. 저장장치 장애·비협조 프로세스는 명시된 신뢰 전제 밖이다.

## 확인 사항

- `paused`만 재개하고 active/aborted/completed는 거부한다. 정상 재개 시 기존 universe 복원은 후속 런타임 책임이다. 최초와 재개 모두 active 기록 및 동기화 성공 전 허가가 없다.
- Pause의 유효 행이 오류 반환 후에도 관찰되는 경우와 Begin 오류 후 active 또는 이전 paused가 관찰되는 경우를 구별한다. Pause 이전 drain과 이후 무전송은 caller 의무이며 모듈의 증명이 아니다.
- root/lock/state 교체, 삭제, rollback, 다른 namespace에 대한 신뢰 전제와 identity 재검사의 한계를 적었다. file sync와 directory sync, 종료 정리 오류, 동일일 독점과 crash 경계를 구분했다.

## P2 — 직렬화와 초기 파일 경쟁의 구현 명세 보강

- `daily-session-plan.md:25,31`: lock 파일은 잠금 획득을 위해 먼저 열거나 생성해야 하므로 ‘읽기/생성 전 획득’은 state 접근을 뜻한다고 명시하고, 동시 lock 파일 최초 생성에서 EEXIST 재개방·identity 재검사·생성자 sync 실패 처리 순서를 구현 전 고정할 것. 현재 문구만으로 오류 없이 구현할 유일한 순서를 판별하기 어렵다. 신뢰 전제 아래 배타성 자체의 결함이 확인된 것은 아니다.
- `daily-session-plan.md:31`: canonical JSON의 필드 순서, 숫자 범위와 checksum 인코딩을 고정하고 비정규·overflow 입력을 fixture로 확인할 것. 호환성 및 fail-closed 해석을 명확히 하기 위한 요구다.

## 미검증

실제 lock/file descriptor 정리 및 I/O fault 선형화, 장애 후 재시작, storage sync 보장, HTTP 진입·응답 drain, KST/거래일 권위와 ledger 복원 결속은 검토 대상 설계만으로 증명되지 않았다. 명시된 표적 시험과 후속 runtime 독립 검증이 필요하다.

## 수정 재검토 — bounded READY

- `daily-session-plan.md:25`는 lock open/create → EEXIST 재열기·identity 검증 → flock → state I/O 순서와 모든 admitted Begin의 lock file/directory sync를 명시했다. 앞선 P2 잠금 순서 지적 해소.
- `daily-session-plan.md:31`은 고정 필드 순서·JSON encoding·uint64 overflow 거부·SHA-256 소문자 64자리 형식을 명시했다. 앞선 P2 직렬화 지적 해소.
- `official-read-compatibility/spec.md:3-34`, `design.md:3`, `issues.md:99-101`, `tasks.md:74-75`를 대조했다. 네 상태와 정상 재개, 호출자 drain, 미배선 범위가 부합한다. P0/P1 추가 발견 없음.
- **bounded READY**: 사전 증거 고정과 신규 미배선 저장 컴포넌트 구현을 위한 설계 리뷰에 한정한다. 실제 구현·fault 시험, 후속 runtime 결속, 운영 storage와 최종 Manager 검증은 아직 미완이다. 테스트·소스 편집·운영 접근은 수행하지 않았다.
