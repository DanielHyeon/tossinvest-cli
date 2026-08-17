# a114 — 콘솔이 자기 lifecycle에 재부착한다

> **상태: 등록만 먼저 했다(2026-08-16).** a109 proposal-freeze P2-7이 기록하고 design
> "선언된 생략"이 후속으로 미룬 항목이다. **착수 선행 조건 없음** — 단 a115(콘솔 전략
> 화면 재부착)와 같은 파일(`cmd/tossctl/console.go`)의 부팅 경로를 편집하므로 동시
> 착수 시 한 change로 합치는 판단을 먼저 한다.

## Why

콘솔의 engine lifecycle client는 부팅 1회 dial이다(a109 기록 당시 `console.go:397`,
a109 design.md 선언된 생략 P2-7). httpapi가 a109 이전에 가졌던 병과 같다: 엔진이
콘솔보다 늦게 뜨거나 가동 중 재시작하면, 콘솔은 재시작 없이 다시 붙지 않는다.

a109 D4는 "재부착은 endpoint 주인이 아니라 **소비자의 것**"이라는 논지로 httpapi에
재부착 wrapper를 세웠고 그 계약은 http-api-service spec에 있다. 같은 논지가 콘솔에는
아직 적용되지 않았다 — a109는 httpapi만 고쳤다(선언된 생략).

## What Changes

- 콘솔 lifecycle client에 httpapi 재부착 계약을 이식한다: 백그라운드 single-flight
  재시도, 렌더·요청 경로에서 dial 금지(Dial은 connect probe를 품는다 — a109 freeze
  P0-2), 전이 시에만 로깅, 밀려난 client의 자원 해제.
- 콘솔 UI에 타이핑 확인·추가 승인 마찰을 넣지 않는다(사용자 지시, 2026-07-27).

## Impact

- operator-console spec: 콘솔의 엔진 lifecycle 읽기 재부착 요구 ADDED.
- 코드: `cmd/tossctl/console.go` 부팅 경로 + 신규 wrapper 파일 + 테스트.
- 착수 시 Function Logic Map 필수: 등록 문서의 분기 주장은 a109 기록 인용이다.
