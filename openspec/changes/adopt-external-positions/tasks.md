# Tasks: adopt-external-positions

> 선행: ① 리뷰 라운드 2 통과(라운드 1 P1 8건 반영 — design.md가 결정 정본) ② **add-operator-dashboard의 journal 조각(RO open·계좌 질의) landed 후 착수**(design D9 — `internal/journal` 동시 작업 금지). 실효는 게이트 ON + `adoption.enabled` 사람 승인 이후. `internal/console`은 건드리지 않는다(대시보드는 데이터 주도라 exit_states 행이 생기면 자동 표시).

## 1. 원장 (v7, additive만) [T]

- [ ] 1.1 `position_adoptions` 테이블 + `positions.adoption_id` 컬럼(design D1 DDL이 정본) — decisions 테이블·CHECK enum 무접촉, 구버전 바이너리는 ErrSchemaTooNew(§0.6 확인)
- [ ] 1.2 `adoption_id` **set-once 전용 tx API** + 정적 스캔 가드(guarded-column 스캔 방식 — 소유 파일 외 UPDATE 언급 거부; `entry_decision_id`에는 어떤 쓰기도 추가하지 않음). 주의: 편입 코드를 `internal/journal` 새 파일에 두면 기존 guarded-column 전문 스캔(apply hook 소유 4컬럼)에 걸린다 — 파일 배치를 가드와 정합시킬 것
- [ ] 1.3 exit 자격 **단일 술어 함수**(`entry_decision_id OR adoption_id`)로 통합 — position/provenance.go·exitloop 열거·exit_state open의 산재 판정을 술어 하나로 교체(drift 테스트)
- [ ] 1.4 exit_state open 경로가 편입 출처를 수용(exitloop.go:505의 RiskIntent 단언 확장 — 편입 포지션은 position_adoptions에서 EntryPrice=observed_price·InitialStop=synthetic_stop·HighWater seed=observed_price), lineage `ADOPTION → POSITION → EXIT_EVENT` 질의 arm 추가
- [ ] 1.5 조정으로 수량 0 → exit_state completed + trade_outcome ADJUSTMENT_CLOSED 동결(고아 방지 — 편입·엔진 포지션 공통)

## 2. 편입 파이프라인 [T]

- [ ] 2.1 **엔진 reconcile 구동 루프 신설**(프로덕션 호출자 0인 IngestExternalPositions/Converge에 구동자): 주기 60초, holdings 스냅샷 → Stabiliser(2초 간격 연속 2회 동일) → fold → 편입 후보 판정. §0.4 계상 문서화. 실행 술어는 `AutomationStatus.Verified`
- [ ] 2.2 편입 판정: enabled=true + 비RECONCILE + 신선 조건(Stabiliser 통과·staleness ≤ 10s) + 비제외 → position_adoptions 영속(원가는 브로커 원문 decimal 문자열 보존 — float 경유 금지, 부재 시 cost_basis_src=ABSENT) → adoption_id set-once → exit_state open. 크래시 복구(영속 후 open 전 → 재시작 완결·중복 금지). 재대사 시 기편입 인식(reconcile external.go:225-234 가드와의 경합 처리 포함)
- [ ] 2.3 **manage-forward 불변식 테스트**: 편입 직후 첫 관측 틱에서 매도 발의 0건(원가 대비 ±50% 보유 포함 — 리뷰 P1-1 회귀 방지), R=0 시작, 이후 상승분부터 래칫 작동
- [ ] 2.4 config: `adoption.enabled`(기본 false — zero-value 안전 확인 테스트), `default_stop_pct`(0<pct<1 검증, 위반 시 설정 거부, 기본값 provenance 기록), `exclude_symbols`. enabled flip의 audit 기록(§0.5)
- [ ] 2.5 알림·이벤트(design D8): 편입 성공 이벤트(편입가·합성 손절), 제외·실패만 알림, 정상 지연 무알림. 외부 수량 증가 감지 알림(동결 유지)
- [ ] 2.6 trade-analytics 구분 집계(adoption_id 조인, 표본 수 병기)

## 3. 완료 게이트 [M]

- [ ] 3.1 §0 검토 기록: 즉시 매도 없음(D2)·기본 OFF(§0.2)·flip 승인(§0.7)·긴급 중지 서술 정직성(D4), flatten이 편입 포지션을 덮는지 테스트
- [ ] 3.2 테스트 전수(-race)·`openspec validate adopt-external-positions --strict`
- [ ] 3.3 `make gate CHANGE=adopt-external-positions`
