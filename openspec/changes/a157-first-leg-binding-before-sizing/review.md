# a157 review

## proposal-freeze 리뷰 (2026-10-18)

- 보이스: 분리 컨텍스트 둘(적대 Eng).
  - Opus 서브에이전트: 저장소 밖 사본에 A1 을 실제로 적용하고 probe 시험을 HEAD · A1 양쪽에서 돌렸다.
  - codex `exec -s read-only`.
- 원문: `analysis/review/freeze-opus.md` · `analysis/review/freeze-codex.md`.
- 판정:
  - Opus **APPROVE-WITH-FIXES**(P0 0 · P1 3 · P2 4 · P3 6). 받아들이는 집합 불변을 실측했고, 주문을 늘리는 갈래는 없다.
  - codex **REJECT as written**(P1 1 · P2 3). D2-c 에서 a137 강화(`ENTRY_BLOCKED` 영속)가 사라진다는 지적이다.
- 사람 결정(2026-10-18, 「1-1, 2-1」):
  - 1-1: 일일 손실 체인 거절을 결속 결함보다 먼저 보고한다(codex P1 수용 — 강화 보존).
  - 2-1: 조립 감독자에 `EntryGate` 를 nil 가드와 함께 배선한다.

| 지적 | 처분 | 반영 자리 |
|---|---|---|
| codex P1 · Opus P2-4 D2-c 강화 상실 | 수용(사람 1-1) — 보고 우선순위 「일일 손실 > 결속 > 그 밖」, 판별 함수 하나를 `admit` 과 공유 | design D1 · D2(c)(d) · spec ADDED · tasks 0.4/1.2 · M13/M14 |
| Opus P1-1 「결함 하나면 그대로」 거짓(실측 2) | 수용 — 「한쪽 판정만 거는 요청」으로 다시 씀, D2(f) 추가 | spec 시나리오 · design D2(f) · D3 · tasks 0.4 |
| Opus P1-2 `EntryGate: c.Entry` 타입 있는 nil → panic | 수용(사람 2-1) — `if c.Entry != nil` | design F10 · D4 · D5 · spec MODIFIED |
| Opus P1-3 부분 이동 · 복제를 못 잡음 | 수용 — 상한 0 + Bn 일곱(B5 둘), 구조 AST 시험, M2~M8 부분 이동 · M15 복제 | design D6 · tasks 0.4/1.1 |
| Opus P2-1 sha 대조 불가(파일 sha) | 수용 — qfinal.go 는 파일 sha, 첫 다리 발급은 함수 본문 sha | design D3 · tasks 1.3 |
| Opus P2-2 dedup 수명 · Run 시험 닿지 않음 | 수용 — 시나리오를 「같은 함수 · 같은 게이트」로, 시험은 칸 단언 + 기록자 직접 호출, dedup 상향은 후속 | design D4 · spec · tasks 후속 |
| Opus P2-3 spec 자기모순(일일 손실 결과 · a137) | 수용 — MODIFIED 에 단서, 1-1 로 a137 충돌 소멸 | spec |
| codex P2 D2 열거 불완전(강화 실패 · 모인 범위 오류 · 측정 불가 기록) | 수용 — 1-1 로 (c)(d) 는 Y 그대로가 되어 모두 보존, 표에 명시 | design D2 |
| codex P2-3 · Opus P3-2 시계 주장 과장 | 수용 — 「같은 입력 · 같은 시계 관측」으로 한정, 지터 명시 | design F1 · D2 · spec |
| codex P2 시험 공백(엔진 경로 · 강화 성공/실패) | 수용 — 엔진 행 B2 · B8 × 일일 손실 둘 | tasks 0.4 · design D3 |
| Opus P3-1 :392 → :394 | 수용 | proposal · design F9 |
| Opus P3-3 F3/F4 는 원문 읽기 | 수용 — 근거 표기 정정 | design 머리 · F3 · F4 |
| Opus P3-4 SHALL NOT 범위 | 수용 — 「일곱 결속 판정 각각, precheck 안에서」 | spec |
| Opus P3-5 「모든 생산 조립」 모호 | 수용 — 두 `Context` 생성자 이름 명시 | spec MODIFIED |
| Opus P3-6 컴파일 실패는 변이 아님 | 수용 — 빌드 메모로 | design D6 |

- 거절: 없음.
- 사람 결정 1-1 로 범위가 넓어졌다. `admit` 의 일일 손실 두 조건이 판별 함수로 옮겨 가므로 `admit` 이 편집 대상에 들어간다(FLM/BTM 필요).
  이 범위 확장이 다른 보고 결과를 바꾸지 않는지는 구현 뒤 독립 리뷰(2.3)에서 다시 본다.

## Pre-Edit Gate

(구현 Teammate 가 FLM/BTM 뒤에 기록)
