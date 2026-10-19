## Why

초기 손절·수량·전략 종료·공통 이익보호를 구분하고, 운영 승인 없이 손절 가정의 수치 영향을 비교한다. a153 합성 검증은 정책 승인 증거가 아니다.

## What Changes

- 사용자가 승인한 범위: 수동 가정값으로 KR/KRW·US/USD의 고정 0.7%, 신호봉 저점, ATR14×1.5 초기 손절을 독립 비교한다.
- 손절가·거리·수량·계획 가격 위험·명목금액을 정확수치로 계산한다. 시세/계좌 조회, 비용·환율·목표 RR 계산과 정책 적용은 제외한다.
- 기존 최적화 화면에 별도 순수 계산 UI를 추가한다. 저장형 Store.preview 및 새 writable descriptor는 사용하지 않는다.
- 원래 계획한 검증된 정책 선택·CAS 적용은 명시적 후속 범위다. 이 change 완료는 그 기능 구현이나 운영 승인 완료를 뜻하지 않는다.
- 선행 a150, a153 완료. 기존 frozen v1, r1, a128/a147 증거·토큰 공유·원장·위험 통제·운영 설정 보존. 실주문·재시작·추가 수집 없음.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `strategy-optimization`: 저장하지 않는 가정값 손절 비교.

## Impact

기존 console에 순수 계산 leaf와 HTTP/UI 연결만 추가한다. DB migration, 새 의존성, 운영 binding 없음. Manager 직접 설계, Terra 구현/시험, 독립·gstack 리뷰, 최종 검증과 전체 gate 후 정식 archive/spec sync·PM 검증·범위 한정 commit/push.
