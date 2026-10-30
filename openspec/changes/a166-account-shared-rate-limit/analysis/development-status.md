# a166 개발 상태와 남은 사람 단계

2026-10-31 Manager. 최종 구조 검사 후보 `9dab3dda4`는 게시된 consumer API `d9827f316`를 합친 source `6c229aac`를 검증했다. 아래는 소스 개발 범위 판정이며 실환경 적용 또는 전체 테스트 PASS가 아니다.

| 작업 | 개발 증거 | 판정 경계 |
|---|---|---|
| 공식 client·token·authority·TLS·OFF | official-lot2, runtime-v2-token-lot, observed-transport-lot의 manager-verification | 실제 official client의 local TLS fixture; 실제 브로커 미실행 |
| replay·probe·a165 첫 AMBIGUOUS | replay-probe-lot2b, wire-structure-plan/consumer-integration | 마지막 합성에서 지정 회귀 3개 normal/tagged race PASS |
| 호출자·공통 root·priority | caller-root-lot3b, remaining-callers, remaining-readers, priority-origin-lot 및 tools/console 로트 | trusted v2 manifest 및 fixture 검증; 운영 namespace와 전체 실행 참여자는 사람 확인 |
| 구조 검사 | wire-structure-plan/result, consumer-integration, manager-verification | 83 함수 형태·116 selector 위치; 임의 동적 호출의 일반 증명 아님 |
| 모든 분류의 429·폴링 | 429-store-lot, observed-transport-lot, engine-poll-lot | 개별 fixture와 실제 client loopback 증거; 전역 cooldown으로 확대하지 않음 |
| retry matrix·소비자 인계 | engine-poll-lot, consumer-api-lot/handoff-table | a087 production admission nil. 실제 board 연결과 a169 fence·abort는 후속 |
| offline provision/check·운영 절차 초안 | offline-budget-tool-lot/manager-verification | 임시 root·fault fixture만. 실제 namespace·배포·복구 명령 미실행 |

각 디렉터리의 Manager 기록은 해당 고정 후보와 당시 독립 검증을 나타낸다. 마지막 합성 시험은 구조 검사·8개 source-copy 변이·symlink 반례·지정 회귀 3개만 실행했으며 모든 과거 시험을 재실행한 것은 아니다. 후속 runtime/token, engine poll, consumer API는 각각 최신 합성 normal/tagged race 기록을 갖는다.

절차 예외 두 건을 보존한다. a130/d2는 상세 FLM/BTM이 구현 뒤 작성되어 사전 준수로 소급하지 않는다(`tools-outcome-lot/manager-verification.md`). 초기 verify RED는 외부 HTTP 401 가능성이 있었고 유효 RED에서 제외했다(`remaining-callers/test-boundary-incident.md`). 수정 fixture는 lo-only/no-route와 exact local TLS를 사용했다.

gstack은 로트별 범위 수동 checklist 검토다. 전체 자동 specialist pipeline 인증이 아니다. CodeGraph 최신 전체 색인 및 전체 tests/gate는 미검증이다. AST/FLM과 source 손읽기 및 행동 시험을 구분한다. impeccable: N/A — backend·CLI·정적 검사와 문서 변경, 새 UI 없음. 기존 Go stdlib와 이미 채택한 내부 모듈을 재사용하며 이 종결 기록에서 새 의존성은 도입하지 않는다.

미완료: 최종 운영 절차 검증(4.1), 실제 호스트/컨테이너의 UID·filesystem·time namespace·동일 root/instance/device/inode·모든 참여자 결속(4.2), 운영 이미지/설정 전환(4.5), a087 사람 실측/활성화(5.4/5.5), a169 실제 reader/fence/지속 abort·리허설/측정. strict 3건/초는 지원하는 guard의 local first-write 시작 기준이고 broker 도착 시각 보장이 아니다. 보호 요청도 대기하며 유한한 최악 wall-clock 대기 상한은 주장하지 않는다.

따라서 코드·문서의 검증된 로트를 commit/merge/push할 수 있으나 a166 change는 active로 유지한다. 사람 단계를 완료로 체크하거나 archive하지 않는다. 기존 main a128 자료는 수정하지 않는다.
