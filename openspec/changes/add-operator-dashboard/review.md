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
