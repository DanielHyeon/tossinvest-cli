# a166 결정 및 차단 원장

## 사용자 결정 — 2026-10-30

계좌 전체 초당 3건 상한을 우선하고 보호 요청의 대기를 허용한다. a166 재설계에 한정한 결정이며 기존 즉시성 원칙보다 우선한다. 구체 지연·실패 기준, 구현·배포·활성화는 승인하지 않았다. 보호 우선은 상한 안의 순서이며 공유 장애에서도 우회하지 않는다.

## 미결 — NOT READY

- 실제 wire 시점과 HTTP 내부 재시도 통제, 조기 응답·느린 body·취소에서 전송 수명 증명.
- 보호 대기자 중재·사망 정리, 공유 파일 오류·손상·교체·mount alias, crash·reboot 복구의 fail-closed 증명.
- 보호 `/prices`·접수 확인·token exchange의 critical 사슬 분류, 장애 결과·429 정책·지연 수용 기준은 사람 결정 필요. 수치 상한을 가정하지 않는다.
- 독립 적대 리뷰 및 Manager 검증 전 구현 금지. a087 production admission nil, a169 보류.

2판 전체는 역사 기록이며 계약이 아니다. UI/UX 영향 없음: impeccable N/A — backend/documentation-only.

## 문서 검토 — 2026-10-30

독립 Terra 검토는 3판 문서 범위의 손 읽기이며 AST 검증이나 실행 시험이 아니다. OFF와 상한 필수 배포의 예산 누락 구분을 반영했다. mixed-model gstack 전체 검토는 수행하지 못했다. OpenSpec 실행 파일은 끊어진 심볼릭 링크라 scoped validate를 수행하지 못했다. NOT READY 유지.
