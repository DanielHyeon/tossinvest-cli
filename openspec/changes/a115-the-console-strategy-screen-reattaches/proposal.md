# a115 — 콘솔 전략 화면도 재부착한다

> **상태: 등록만 먼저 했다(2026-08-16).** a109 A2 P1-1이 기록하고 design "선언된
> 생략"이 후속으로 미룬 항목이다. **착수 선행 조건 없음** — 단 a114와 같은
> `cmd/tossctl/console.go` 부팅 경로를 편집하므로 동시 착수 시 합본 판단을 먼저 한다.

## Why

콘솔의 전략 projection dial은 부팅 1회이고, **dial 실패를 nil로 접는다**(a109 기록
당시 `console.go:411–421`, a109 review §2 A2 P1-1). 콘솔 page는 dormant와
unavailable을 구분하도록 만들어져 있지만, 콘솔 boot가 실패를 접으므로 화면은
NOT_CONFIGURED로 남는다 — 구성돼 있으나 닿지 않는 상태를 미구성으로 오귀속한다.

httpapi 쪽 같은 병은 a109 D4가 고쳤고 그 오귀속 금지는 http-api-service spec에 있다.
콘솔은 선언된 생략으로 남았다.

## What Changes

- 부팅 1회 dial의 nil 접힘을 제거하고 httpapi 재부착 계약을 이식한다: 부재/실패는
  접히지 않는 상태로 화면에 전달되고, 재시도는 백그라운드 single-flight다.
- 화면 문구는 미구성(dormant)과 도달 불가(unavailable)를 구분하며 엔진 부재를
  단정하지 않는다(a109 D3a-2 준용).

## Impact

- operator-console spec: 전략 화면의 상태 구분·재부착 요구 ADDED.
- 코드: `cmd/tossctl/console.go` 부팅 경로 + 테스트. a114와 표면 공유.
- 착수 시 Function Logic Map 필수: 등록 문서의 분기 주장은 a109 기록 인용이다.
