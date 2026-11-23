# a199-a-stale-capability-proof-blocks-entry-not-exit Review

## 계약 작성 (2026-11-24)
- 작성: Manager(Fable). 사전 증거 조사: 읽기 전용 조사원(HEAD `9738a0c32`) — change 안 사본·HEAD 재확인은 tasks 1.3.
- 사용자 승인: 2026-11-24 「1번 올린다로 진행해줘」 「지금부터 작성하고 구현해줘」.
- 위험 등급 High-risk(기동 인터록 · 주문 권한 · 손절 경로 가용성).

## 1단계 — 등록 · 증거 · 제안 동결 리뷰 (팀메이트 Opus, 2026-11-24)

worktree `~/tossos-work/a199-impl`, 브랜치 `feat/a199-stale-proof-exit-only`. production·시험 Go 편집 0.

### 1.1 등록
- 계약서 넷을 옮김(`spec-delta-intent.md` 는 `analysis/`). MODIFIED 2건은 `analysis/modified-delta/generate.py` 로 생성 — 치환마다 1회 적중 단언, 역치환으로 정본 Requirement 바이트 동일 단언, 정본 Scenario 이름 보존 단언(A 12→17, B 18→20). 실행 원문 `analysis/modified-delta/run.txt`.
- 사실 정정(뜻 불변): intent A 의 추가 Scenario 자리 「`#### Scenario: dry-run` 앞 또는 뒤」는 다른 Requirement(Flatten Saga)의 Scenario 다 → 「자동화 게이트 기동 인터록」의 마지막 Scenario(「정수가 아닌 최대 보유 포지션 수」) 뒤에 둠. proposal 「영향」의 「ADDED 2~3」 → 「ADDED 없음」(intent C — A2 문단이 Requirement 안에서 strict 통과).
- B3 자리 수: 아래 「B3 자리표」.
#### B3 자리표 — Requirement 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」 안의 「재시작이 거절」 계열

정확한 문자열 「재시작이 거절될 상태」는 **2자리**, 같은 뜻을 다른 꼴로 말하는 자리까지 **6자리**(grep `거절|거부`, 정본 2213-2297).

| # | 정본 줄 | 원문 | 처리 | 뜻 |
|---|---|---|---|---|
| 1 | 2216 | `**재시작이 거절되기 시작할 시각(기한)**` | `**재시작의 결과가 정상이 아니게 되는 시각(기한)**` | 기한의 값·계산 무변, 이름만 새 귀결에 맞춤(B3 과 같은 꼴). **intent 밖 — Manager 확인 Q** |
| 2 | 2220 | `재시작이 거절될 상태의 알림은 …` | B2 의 새 문장 전체 | intent B2 |
| 3 | 2221 | `재시작이 거절될 상태(만료 · 범위 상실 · 계좌 불일치 · 무효)는` | `재시작의 결과가 정상이 아닌 상태(…)는` | intent B3, 괄호·주기·키 무변 |
| 4 | 2221 | `거절 상태가 같은 한(` | `그 상태가 같은 한(` | 3 의 집합을 가리키는 줄임말 — 뜻 불변 |
| 5 | 2221 | `다시 거절 상태가 되면` | `다시 그 상태가 되면` | 같음 |
| 6 | 2297 | Scenario 「범위 상실 알림은 결과와 할 일을 말한다」 THEN 「기동이 거부된다 … 자동 손절이 돌지 않게 된다」 | 「새 진입이 막힌 청산 전용으로 뜬다 … 손절·청산은 돌고 새 진입만 막힌다」 | B2 가 정한 범위 상실의 새 귀결을 옮김. intent 는 「만료됨」「갱신이 주문 권한을 잃음」만 점검 대상으로 적었고 이 Scenario 를 빠뜨림 — 그대로 두면 Requirement 가 자기모순. **Manager 확인 Q** |

「만료됨」「갱신이 주문 권한을 잃음」 THEN 에는 귀결 문장이 없어 무편집(intent 대로).

- Story `STORY-TOS-a199`(FEAT-TOS-003 「Engine interlocks and Guardian controls」 — a196 과 같은 Feature), `_registry.yaml`, tracker 재생성 + `--check`.
