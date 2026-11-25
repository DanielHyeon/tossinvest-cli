# a168 review

## 0. 착수 기록 (2026-11-25, 구현 팀메이트 Opus 5.5, worktree `~/tossos-work/a168-impl`, base `e57d06cd6`)

- 기억 recall: ai-memory `memory_query`(현재 프로젝트) — a164 종결·a168 정의 세션 관찰만, 설계 교훈 없음. Hindsight `tossos-dev-episodes` —
  a164 종결(마스크 누출 시험)·gate ⑤ 형제 함수 유입·a165 종결 교훈(ext4 worktree, codegraph 강제 종료 금지)·a166 census. 이 change 의
  설계 근거로 쓴 것 없음(전부 현재 HEAD 로 재측정). 파일 기억 중 적용: 「새 가드는 뒤의 가드를 못에서 뽑는다」 → D3,
  「판정이 둘이면 반증이 죽는다」 → D2 의 꼬리 처리, 「fail-closed 는 무엇을 거부하는지」 → §6, 「호출자 수는 고칠 자리 수가 아니다」 → §4,
  「a166 census」 → §7 확인.
- 사람 결정 Q1 (a) · Q2 제외 · Q3 충족 → proposal 「사람 결정」 절.
- Function Logic Map: 1단계는 AST 산출물만(`analysis/ast/`, 34개). 편집 대상 기존 함수 4개(design §7)의 FLM/BTM 번들은 2단계 Pre-Edit 에서
  `tools/logic-map/scaffold_analysis.py` 로 만든다. 1단계 코드 편집 0.

## 1.3 proposal-freeze 리뷰 — 1회차 (2026-11-25)

- 보이스: (1) Claude 독립 적대 서브에이전트(code-reviewer, 신선한 문맥, 읽기 전용 — 스크래치 탐침 하나 `go run`, 저장소 쓰기 0),
  (2) Codex 적대 패스(`codex exec -s read-only`, `</dev/null`, 다른 계열). 같은 프롬프트(필수 질문 「구현 뒤에도 줄이 써지는 입력」 포함).
  리뷰어에게 픽스처가 필요하면 `set -euo pipefail` · 저장소 밖 `mktemp -d` · `git -C` 절대경로 · toplevel 단언을 지시함.
- 필수 질문 답: 둘 다 같은 하나 — **홀로 선 대소문자 변형 키**(by-id 의 `first.TriggeredOrderId` · `Second`). D2 는 「짝」 만 보고 by-id 해독은
  정확한 키만 읽어서 증거가 사라진다. a164 부터 있던 해석 차이.

| # | 보이스 | 요지 | 처분 |
| --- | --- | --- | --- |
| F1 (P1) | 둘 다 | 홀로 선 변형 키가 by-id 에서 발동 흔적·둘째 다리를 지움 → 만료 줄 가능 | **Manager 결정 대기** — design §10-4. 사용자 지시(리뷰가 제안한 새 가드는 계획에 넣지 않음)에 따라 계획 밖. 권장 (나) 포함 |
| F2 (P2) | Claude | P5 「두 모드 공통」 틀림 — 부재 모드는 CLOSED 행을 전부 거절(빈 허용 목록). W1b 행(`CANCELED`)은 이미 두 모드에서 거절 → R8 이 RED 아님 | **수용** — design §1 P5 정정(AST `checkReconcileClosedStatus` B2 · `checkReconcileExpiry` B2·B14 인용), R8 → R8a(만료·`EXPIRED` 다른 id 행)·R8b(부재·OPEN 행 통째 숨김)·R8c, spec 시나리오 둘로 나눔 |
| F3 (P2) | Claude | R7(401→201)은 「첫 시도」 변이 M2 를 못 잡음 | **수용** — M2 는 C2 로만 잡는다고 정정 |
| F4 (P2) | Claude | 비 ASCII 접힘(`ſ`·Kelvin)·이스케이프 키가 RED 에 없어 `ToLower` 변이·원문 바이트 비교 변이가 산다 | **수용**(시험 추가일 뿐 새 가드 아님) — R1·R5 입력 추가, M3b·M3c |
| F5 (P3) | Claude | `strictMinuteCheckJSON` 재사용 안 하는 이유 없음 · O(본문×깊이) 비용 | **수용** — D2 에 비재사용 근거 3, §9 R5 |
| F6 (P2) | Codex | 404 본문의 중복 `code` 는 silent 로 가름 — spec 의 「모든 본문」 과 충돌 | **수용(spec 문장)** — 규칙을 HTTP 200 본문으로 좁힘, §9 R7(줄 쓰는 길 없음 — 만료 모드는 silent 를 근거로 안 셈). silent 분류 불변은 a164 계약 |
| F7 (P2) | Codex | 「기존 수락 시험 GREEN = 동적 census」 는 과장 — 직접 해독 시험은 새 판정을 안 지남 | **수용** — §6 을 클라이언트 경로 시험으로 좁히고 C6 추가 |
| F8 (P2) | Codex | 「모든 깊이·모든 키」 가 RED 에 없음(선택적 재귀 구현이 통과) | **수용** — R10 · M9 |
| F9 (P2) | Codex | census 술어(파이썬 `lower`/`casefold`)가 Go `EqualFold` 와 비 ASCII 에서 다름 | **수용** — census 가 비 ASCII 키 수를 셈(E = 0 → 이 모집단에서 동치), §6 문장 정정. go.mod 1.25.0 vs 측정 1.26.5 는 §9 R8 |
| — | Claude | Q2 「마지막 관측 = send 가 쓴 시도」 · Q3 해독-판정 사이 탈출 · Q5 census/시험 파손 | **발견 없음**(리뷰어 확인: 토큰 요청은 `m.hc.Do` 라 관측 안 됨, a166 전송은 오류만, silent 조기 반환은 verifylive 가 거절) |

반영 뒤 `openspec validate a168-verify-read-decode-hardening --strict --no-interactive` rc=0.

## 1.3 proposal-freeze 리뷰 — 2회차 (2026-11-25, 마지막 — 상한 2회)

- 보이스: Codex 적대 패스(`codex exec -s read-only`, `</dev/null`) — 1회차 처분과 정정 문장만 대상. Claude 보이스는 1회차로 갈음(2회차는
  정정 검증이라 한 보이스).
- 판정: 새 P0·P1 없음. 1회차 처분 F1~F9 의 방향은 유지.

| # | 요지 | 처분 |
| --- | --- | --- |
| G1 (P2) | M3c 의 이스케이프 증인이 잘못 — 같은 철자로 해독되는 이스케이프는 `reconcileObject` 줄 115 가 정확한 중복으로 먼저 거절해 변이를 못 가름 | **수용** — 증인을 `"triggeredOrderId"` + `"\u0054riggeredOrderId"`(해독 `TriggeredOrderId`, 접힘 짝)로 바꿈(design §8 R1·M3c, D2 문장) |
| G2 (P3) | Kelvin 키는 `ToLower` 변이 M3b 를 못 잡음(`K`→`k`) | **수용** — M3b 증인은 `ſ` 만, Kelvin 은 호환 확인용 |
| G3 (P3) | R5 비용에 접힘 짝 비교 O(키²) 가 빠짐 | **수용** — §9 R5 문장 보강 |

반영 뒤 `openspec validate a168-verify-read-decode-hardening --strict --no-interactive` rc=0(아래 커밋 직전 재실행).
리뷰 상한(2회)에 닿았으므로 더 돌리지 않는다. 남은 결정은 design §10(Manager 결정 대기 1~4).
