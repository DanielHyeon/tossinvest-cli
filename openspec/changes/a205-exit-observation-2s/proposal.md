# a205 — 청산 관측 주기를 5초에서 2초로 줄인다 (exit-observation-2s)

## Why

사용자 지시(2026-11-27): 보유 종목 시세 평가 주기를 5초에서 줄이고 싶다 — 「1분만에도 요즘은 등락폭이 높아」. 사용자가 고른 안은
**두 시장(KR · US) 모두 청산 관측 주기 5s → 2s** 다.

청산 관측 루프(`ExitObserver.Run`)는 KR · US 를 한 루프 · 한 가격 배치로 보므로 상수 하나가 두 시장을 함께 바꾼다(사용자가 그것을 고름).
워터마크는 관측 **표본**의 최고가이고 표본 사이에 스친 트리거는 놓친다 — 그 허용폭을 정하는 것이 관측 주기다(exit-policy · `exitpolicy/doc.go`).
주기를 줄이면 손절 · 익절 트리거를 늦게 보는 최대 폭이 약 5.1초에서 약 2.1초로 줄어든다(보수 방향 — 손절을 더 빨리 봄).

Manager 실측(읽기 전용, 2026-11-27, 계좌 정보 없음 — design §1 에 인용):

- 지금 실제 간격은 대기 5s + 처리 0.07~0.16s(US 1종목 106 간격 중앙 5.083s, 6s 초과 0). KR 처리시간 p99 0.49~0.68s, 최대 2.6s(8월).
- 60초 이상 관측 두절은 10~11월 9회, 전부 계정 전체 브로커 읽기 장애. 체결 감지 SLO 양보 실측 0회.
- 가격 읽기는 주기당 배치 1회 → 분당 12회가 30회가 된다. 계좌 정책 상한(a166) 3 req/s, 운영 예산 OFF.

## What Changes

1. `internal/app/engine/exitloop.go` `DefaultExitObservationInterval` 5s → **2s**(상수 변경, 설정 아님 — design D1).
   틀린 주석(「It is four intervals」 — 60/5 는 12 이고 이제 60/2 = 30)을 고친다.
2. 하한가 갱신기의 읽기 시한은 **주기에서 계속 유도한다**(코드 변경 0 — design D2): 2s 에서 읽기 1s · 사전조회 1s, 예산 합 = 주기(5s 때와 같은 구조).
   주기와 분리하면 예산이 주기를 넘어 a087 의 「시세 읽기와 시간상 겹치지 않는다」(SHALL NOT)를 구조적으로 깸. 생산 갱신기는 admission `nil` 이라
   지금 꺼져 있음(`engine.go:627`) — 짧아진 시한의 영향은 a166 admission 배선 뒤의 조건부 위험(design R1 · R1b).
3. 콘솔의 두 사본 상수(`lineRefreshInterval` · `engineRuntimeInterval`)를 2s 로 맞춘다 — 콘솔은 엔진 패키지를 import 할 수 없다
   (`TestTheConsoleHoldsNoBrokerOfItsOwn` 금지 목록, design D4). operator-console 정본이 runtime 읽기 간격 ≤ 엔진 관측 주기를 요구한다.
4. `a197_external_opposite_buy.go` 의 `externalBuyListTimeout`(3s) 주석 「관측 주기 5 초보다 짧고」를 고친다 — 값은 그대로(design D3).
5. exit-policy spec 델타(`관측 경로와 fail-safe` MODIFIED): 「주기 기본 5초」 → 2초, 「관측 주기는 … 기본 5초」 → 2초, 시나리오 「기본 관측 주기 (a205)」 추가.
6. 시험: 엔진 기본값 2s · 하한가 예산 유도값(읽기 1s · 사전조회 1s) 고정 시험 새 파일, 콘솔 두 시험의 기대값 2s, 콘솔 개요 재로드 시험 횟수 10 → 16.

## Non-Goals

- 주기를 설정값으로 만들지 않는다(사용자가 요구하지 않음 — YAGNI). KR · US 를 따로 두지 않는다.
- 두절 임계 60s(시간 기준) · 시세 증거 15s(`QueryPriceEvidenceDuration`) · a197 목록 시한 3s · 하한가 갱신 공식 · 체결 감지 주기 3s 는 바꾸지 않는다.
- 저널 쓰기 줄이기(heartbeat · 거절 반복 행)는 이 change 가 아니다 — 늘어나는 양은 design §5 에 적고 남은 위험으로 둔다.
- 배포 · `make image` · 엔진 교체 · 운영 토글 · main 병합 · archive 는 하지 않는다(사람 승인 별도).

## Impact

- 위험도: **High-risk**(손절 · 익절 관측 경로). 바뀌는 기존 (생산) 함수 본문 0(상수 · 주석만) — 기존 시험 함수 본문 3개는 FLM 번들(design §7).
- 브로커 호출: 청산 가격 배치 0.2 → 0.5 req/s, 막힌 손절 재제출 포지션당 ≤0.2 → ≤0.5 POST/s(평균 추산 — design §5). 하한가 갱신기는 생산 dormant(활성화 뒤 그날 기록이 없는 동안 주기당 ≤1건).
- 저널 쓰기: 포지션마다 주기마다 `exit_states` heartbeat UPDATE → 2.5배. 거절 반복 구간의 `exit_events` 행도 2.5배(Manager 실측 근거).
- 토글 OFF = upstream: upstream tossinvest-cli 에는 엔진 · 콘솔이 없다(design D6) — 규칙의 대상 동작이 없음. 콘솔 재로드는 엔진 실행 여부와 무관하게 2초가 됨.
- 롤백: 상수를 5s 로 되돌린 이미지(또는 직전 이미지)로 엔진 교체. 저널 · 스키마 변경 0.
