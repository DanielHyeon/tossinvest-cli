# 리뷰 기록: add-operator-dashboard

## 라운드 1 (2026-07-27, 적대적 리뷰 — 판정 REVISE, P1 3·P2 6·P3 3)

Manager 재검증: journal.Open의 무조건 mkdir+migrate·mode=ro 부재, trade_outcomes DDL(진입/청산가 컬럼 없음)을 코드로 확인 — 사실.

### Manager 처분 (2판에 반영)

- **P1 journal RO 불가 + 소유권 충돌**: `journal.OpenReadOnly` additive 신설을 이 change의 task 1.1로 소유. adopt-external-positions는 이 조각 landed 후 착수(그쪽 design D9와 상호 참조)
- **P1 trade_outcomes 필드 부재**: 이력 화면 필드를 실제 컬럼 + 명시 조인(positions 심볼·exit_states 진입가)으로 축소, 청산가 표시 금지·재계산 금지 SHALL NOT
- **P1 광폭 브로커 인터페이스**: 조회 전용 인터페이스 선언 + mutation 메서드 0 정적 테스트 SHALL(스펙·task 1.3)
- **P2 §0.4 계상·fan-out**: lazy·폴러 금지·갱신당 holdings 1콜(lastPrice 활용 — 리뷰어 지적대로 별도 시세 콜 중복 제거)·TTL을 요구사항으로 성문화
- **P2 runlock 의미론**: in-process 신호 + 타 프로세스 mtime 신선도(5분 상한)로 정확화
- **P2 콜드 캐시**: 시나리오 추가. **P2 계좌 단위 질의 부재**: task 1.2로 신설 명시. **P2 1.8 선행**: "커밋 완료 후"로 게이트 명문화 + static_test 정합 task 2.3. **P2 capability 흡수**: proposal의 "execution-verification 계약 그대로" 문장 삭제, 콘솔 안전 불변식 요구사항 신설(성문화이며 동작 변경 아님을 명시)
- **P3**: 라벨 "관리 외(미편입)" 단일화, 라우트 가드 구체화(GET-only·mutation 동사 스캔), ErrSchemaTooNew 안내 시나리오

## 라운드 2 (대기 — 라운드 1 기준 "3건 반영 시 freeze 가능" 판정이었으나 편입 change와 함께 확인)

## 라운드 2 (2026-07-27 — 판정 REVISE, 신규 P1 2·P2 5·P3 2; 라운드 1 P1 3건은 전부 종결 확인)

### Manager 처분 (3판에 반영)

- **P1 인증 모델 오서술**: 채택 — 안전 불변식 요구사항을 2경로 인증(세션 토큰: 프로세스 수명·터미널 근원 / 핸드오프 토큰: 단발·0600 파일 근원·기인증 세션만 발행)으로 재작성, 성문화 범위 1.6~1.8로 정정, 핸드오프 재사용 거부 시나리오 추가
- **P1 adoption_id 순환 의존**: 채택 — 자격 표시는 entry 결정만으로 선착지(스펙·태스크 문구), 편입 확장은 adopt-external-positions task 2.7
- **P2 비루프백 시나리오**: 채택 — 리스너 수준 서술로 정정. **P2 프로세스 제어 능력 누락**: 채택 — "상태변경 행위는 검증 제어+자기·soak 재기동뿐(계좌 무접촉)" 명문화. **P2 OpenReadOnly WAL**: 채택 — `-shm`/`-wal` 예외 명시. **P2 구스키마**: 채택 — 양방향 구분 안내(신규 시나리오·task 1.1). **P2 편입 이력 공백**: 채택 — 이력 화면 한계 명시+exit_events 보완(스펙·task 2.2)
- **P3 정적 가드 목록·held_seconds NULL**: 채택(task 2.3·2.2)

## 라운드 3 (대기)

## 라운드 3 (2026-07-27 — 판정 **FREEZE-READY**)

라운드 1 P1 3건·라운드 2 P1 2건 전부 종결. 인증 요구사항을 landed 1.8 코드와 절 단위 1:1 대조(7/7 일치 — 리뷰어 대조표). 랜딩 전 편집 2건 반영:

- **P2 핸드오프 TTL**: 채택 — "짧은 유효 시간(현행 2분) 경과 시 거부" 보강(handoff.go:63-70 근거)
- **P2 시점 상대 괄호**: 채택 — 자격 정의를 exit-policy·position-ledger 소유로 위임
- P3(콘솔 문구 "1회용 링크" vs 핸드오프 용어): 구현 노트로 대시보드 태스크 수행 시 함께 정리

**freeze 확정 — 구현 착수 승인(1.8 커밋 완료 선행 충족).**
