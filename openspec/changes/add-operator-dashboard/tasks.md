# Tasks: add-operator-dashboard

> 선행: 2b task 1.8(콘솔 운영 자동화·한글화) 완료 후 착수 — 같은 `internal/console` 패키지의 충돌 방지. 2b 측정·2c와는 독립(측정을 막지도, 요구하지도 않는다).

## 1. 데이터 접근 [T]

- [ ] 1.1 journal **읽기 전용 리더**: 콘솔용 조회 함수 — positions(+exit_states 조인), trade_outcomes, exit_events 시간순. SQLite RO open(쓰기 연결 부재를 가드 테스트로 고정), journal 파일 부재·빈 테이블은 오류가 아니라 빈 결과. busy timeout 설정(엔진 가동 후 동시 접근 대비)
- [ ] 1.2 **브로커 스냅샷 캐시**: 보유·매도가능·시세를 서버측 캐시(TTL ≥ 15s)로 제공, TTL 내 재요청·다중 탭은 브로커 호출 0건, 캐시 시각 노출. **runlock 활성(검증 실행) 중에는 갱신 보류**하고 캐시 값 사용 — 검증의 rate budget을 대시보드가 뺏지 않는다

## 2. 화면 [T]

- [ ] 2.1 **포지션 화면**: 심볼 조인 렌더 — 엔진 관리 포지션은 exit 라인(진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부·pending exit) 표시, `EntryDecisionID` 빈 보유는 "편입 전" 구분(사유 안내 — 편입은 adopt-external-positions), 어느 소스가 비어도 다른 쪽만으로 렌더(조인 실패 ≠ 화면 실패), 엔진 미가동 빈 상태 정직 표시. 평가손익은 캐시 시세 기준·캐시 시각 병기
- [ ] 2.2 **거래 이력 화면**: trade_outcomes(동결 값 그대로, 재계산 금지)·exit_events 시간순, 비어 있으면 "완결 거래 없음"
- [ ] 2.3 라우트·내비 추가 — 게이트·주문·설정 라우트 부재 가드는 새 라우트 표에서도 통과해야 한다(기존 정적 가드 확장)

## 3. 완료 게이트 [M]

- [ ] 3.1 테스트 전수(-race)·`openspec validate add-operator-dashboard --strict`
- [ ] 3.2 `make gate CHANGE=add-operator-dashboard`
