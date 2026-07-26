# Tasks: add-operator-dashboard

> 선행: 2b task 1.8 **커밋 완료** 후 착수 — 같은 `internal/console` 패키지의 충돌 방지(라우트 표·static_test 골든 목록·templates nav·cmd/tossctl/console.go를 1.8이 수정 중). 2b 측정·2c와는 독립. **이 change가 `internal/journal` 추가분(RO open·계좌 질의)을 소유**하며, adopt-external-positions 구현은 그 조각 landed 후 시작한다.

## 1. 데이터 접근 [T]

- [ ] 1.1 `journal.OpenReadOnly`(additive 신설): DB 파일·디렉터리 생성 없음, 마이그레이션 없음, DB 쓰기 없음, `mode=ro` DSN, busy timeout — WAL 공유 인덱스(`-shm`/`-wal`) 접근은 명시된 예외(스펙 문언). 파일 부재는 typed 오류 → 빈 상태 렌더; **스키마 불일치 양방향 구분**: ErrSchemaTooNew(더 새로움 → "콘솔 업데이트 필요")와 필요 테이블 부재(더 오래됨 → "엔진 기동 필요") 각각 typed 전달. 쓰기 연결 부재 가드 테스트
- [ ] 1.2 계좌 단위 질의(additive — `internal/journal`, 스키마 무변경): positions ⟕ exit_states(자격 포함 전체 — 기존 OpenExitStates는 미완결만 반환하므로 신규), 계좌 단위 exit_events 시간순(기존은 position_id 단위뿐), trade_outcomes + positions(심볼) + exit_states(진입가) 조인
- [ ] 1.3 **조회 전용 브로커 인터페이스**(holdings 계열만 선언 — verifylive.Broker 주입 금지) + mutation 메서드 0 정적 테스트. 스냅샷 캐시: lazy·백그라운드 폴러 없음·갱신당 holdings 1콜(lastPrice 사용, 시세 fan-out 금지)·TTL ≥ 15s·캐시 시각 노출. 검증 중 보류: in-process run 신호 + 타 프로세스는 runlock mtime 신선도(5분 상한)

## 2. 화면 [T]

- [ ] 2.1 **포지션 화면**: 심볼 조인 렌더 — exit 자격(이 change 시점: `entry_decision_id`만 — 편입 확장은 adopt-external-positions task 2.7) 포지션은 exit 라인(진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절·pending exit), 자격 없는 보유는 "관리 외(미편입)"(라벨 단일화·사유 안내), 어느 소스가 비어도 렌더(조인 실패 ≠ 화면 실패), 엔진 미가동·콜드 캐시·스키마 불일치(양방향) 각각의 정직한 안내, 평가손익은 캐시 시세·캐시 시각 병기
- [ ] 2.2 **거래 이력 화면**: trade_outcomes 동결 값 + 명시 조인 필드만(심볼·실현손익·실현 R·초기 수량·보유 시간(NULL은 "—")·exit 단계·청산 시각·진입가) — 청산가 등 스키마에 없는 값 표시 금지, 재계산 금지, 비면 "완결 거래 없음", 외부 매도 종결은 성과 행이 없음을 명시(exit_events가 보완)
- [ ] 2.3 라우트·내비: 신규 라우트 전부 GET·stateChanging 무추가, mutation 동사(sell/cancel/modify/gate/approve) 경로 부재 정적 스캔, 기존 404 가드 유지. static_test 정합: 파일명 화이트리스트(`rendering` 맵)·os.OpenFile/os.MkdirAll·os.Getenv 금지 목록·`.Outside`/`KRSessionAdvisory(` 소유 파일 제한을 새 파일(positions/snapshot 등)과 정합(언급만으로 실패하는 전문 스캔들)

## 3. 완료 게이트 [M]

- [ ] 3.1 테스트 전수(-race)·`openspec validate add-operator-dashboard --strict`
- [ ] 3.2 `make gate CHANGE=add-operator-dashboard`
