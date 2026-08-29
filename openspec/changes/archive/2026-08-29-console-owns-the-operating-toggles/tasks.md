# Tasks — console-owns-the-operating-toggles

## 1. 계약

- [x] 1.1 proposal·design, base-commit 고정
- [x] 1.2 operator-console 델타 — 운영 섹션 둘, 사전 판정, 감사
- [x] 1.3 engine-safety 델타 — 인터록 3절이 place·cancel도 검사한다
- [x] 1.4 `openspec validate --strict`

## 2. RED

- [x] 2.1 인터록 3절: `place=false`만으로 기동이 거부되고 이름이 열거된다
- [x] 2.2 인터록 3절: `cancel=false`만으로 기동이 거부된다
- [x] 2.3 화면이 거래 정책 네 토글을 현재 값으로 렌더한다
- [x] 2.4 거래 정책 저장이 네 개만 쓰고 amend·conditional·fractional을 보존한다
- [x] 2.5 게이트 스위치 저장이 한도를 건드리지 않는다
- [x] 2.6 한도 저장이 스위치를 건드리지 않는다 (기존 단언 유지 확인)
- [x] 2.7 사전 판정이 미충족 항목을 이름으로 열거하고, 판정 불가 항목을 밝힌다
- [x] 2.8 게이트 ON 화면이 편입 비가역성·보호 수명·진입 거부를 문장으로 말한다
- [x] 2.9 타이핑 확인 입력이 화면에 없다
- [x] 2.10 RED 관측 기록

## 3. GREEN

- [x] 3.1 `checkTradingPolicy` — place·cancel 추가
- [x] 3.2 `TradingPolicySettings` seam + surgical write
- [x] 3.3 `GateSwitch` seam (한도와 분리)
- [x] 3.4 `/settings` 운영 섹션 둘 + 핸들러 + CSRF
- [x] 3.5 사전 판정 렌더
- [x] 3.6 cmd/tossctl seam 주입
- [x] 3.7 전체 GREEN

## 4. 증거·게이트

- [x] 4.1 Function Logic Map — 변경한 기존 함수 전부
- [x] 4.2 review.md 적대적 리뷰
- [x] 4.3 `make sdd-sync` → `make gate CHANGE=console-owns-the-operating-toggles`

## 인계 (이 change의 태스크가 아니다)

바이너리 설치 후 `/settings`에서: 거래 정책 네 개 → 게이트 ON → `tossctl engine run`.
`config.json` 손편집은 더 이상 필요하지 않다.
