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

## Manager 결정 (2026-11-25) — design §10 확정

| # | 결정 | 반영 |
| --- | --- | --- |
| 1 | D2 = (가) 모든 깊이·모든 객체의 접히는 두 키 | design D2 그대로 |
| 2 | `ReconcileInstrument` 포함 | design D1·D2 그대로 |
| 3 | D5 체인 없음 수락, R4 잔여, 새 구조 시험 없음 | design D5 그대로 |
| 4 | 홀로 선 대소문자 변형 키 = (나) 포함 — 평평한 규칙 하나, 계약 키 목록 한 곳 + 목록 핀 시험 하나만, 재census 0 이 아니면 멈춤 | design D6 · §8 R11·R12·M10·M11 · spec 셋째 Requirement(strict valid) · 재census(design D6 끝 — 응답 키 0, 문서 구조 키 2건은 응답 아님으로 판단) |

- 기록 정정: 커밋 `23ce14079` 제목의 태그 [0.3 1.1 1.2 1.3] 는 실제 체크(1.3 만, 0.1·0.2 는 앞 커밋)와 달랐다 — 커밋은 고쳐 쓰지 않음.

## 2.1 Pre-Edit Gate (2026-11-25, 편집 전 — base `e57d06cd6`, 브랜치 HEAD `34dd7a005`)

```text
Pre-Edit Gate:
- change id / task id: a168-verify-read-decode-hardening / 2.1~2.3
- 대상 심볼: official.Client.ReconcileConditionalOrdersPage · Client.ReconcileOpenOrdersPage · Client.ReconcileInstrument ·
  Client.ReconcileConditionalOrder (기존 4) + 새 파일 internal/official/reconcile_response.go 의 새 함수
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-queries.txt · codegraph-baseline.md (1.6.0)
- CodeGraphContext 후보와 evidence reconciliation: CGC incomplete(시간 초과, 안 봤음) · analysis/code-context/evidence-reconciliation.md
- 기존 동작 파악 근거: 현재 HEAD · AST analysis/ast/(34) · 실행 측정 analysis/measure/decode_ambiguity-output.txt ·
  기존 시험 reconcile_reads_test.go · reconcile_reads_green_test.go · reconcile_detail_reads_test.go · verifylive reconcile_*_test.go
- Function Logic Map / Branch Test Map: analysis/function-logic/internal-official--client.{reconcileconditionalorderspage,
  reconcileopenorderspage,reconcileinstrument,reconcileconditionalorder}/ (편집 전 AST)
- upstream 상속 테스트 영향: no — 대사 읽기는 TossOS 전용(a121·a164). send 불변이라 다른 공식 호출 영향 0
- 실패 테스트 선행 작성: yes — 2.2 RED 를 편집 전에 실행해 실패(줄 씀 포함)를 analysis/red-green/ 에 남김
- 설정·DB·journal 변경과 rollback: 없음 — 코드만, 되돌리기는 커밋 revert
- 안전 불변식 §0 위반 여부 검토: 통과 — 주문 side effect 0(대사 GET 만), 토글 0, 손절·청산 경로 무관, 거절만 늘어남
```

## 2.2 · 2.3 RED → GREEN (2026-11-25)

- RED(편집 전, HEAD `34dd7a005`·생산 코드 미편집): `analysis/red-green/red-official-before.txt` — 하위 시험 54 실패(모호한 본문 30 ·
  200 아닌 2xx 20 · 401→201 4). `red-verifylive-before.txt` — 10 실패 중 **9 건 「줄 씀」**(R8a · R8b ×2 · R8c · R9 · R11 ×3 · R12 lone
  `Result`, 각 `1 line(s) appended`), 1 건 코드 불일치(R12 행 단독 변형 → 편집 전 `closed-fired-trace`). 처음부터 GREEN 인 대조·특성
  (C1 · C4 · C5 · R-T · 입력 검사 · 204) 은 RED 로 세지 않음.
- GREEN(`ee029b37e`): `green-official-after.txt` · `green-verifylive-after.txt` — 실패 0.

## 3.1 변이 원장 (사본 worktree `~/tossos-work/a168-mut`, 시작 sha 단언 · 무변이 대조군 GREEN · 한 번에 한 판)

- 하네스 `analysis/mutation/run_mutations.py`, 원장 `analysis/mutation/mutation-ledger.txt`(최종 `c25e2fad0`): **CAUGHT 22 · SURVIVED 0 ·
  INVALID 0**. 선택 시험 `go test -count=1 -run TestReconcile ./internal/official/ ./internal/verifylive/`.
- 1판(`ee029b37e`)에서 M1 이 `net/http` import 미사용 컴파일 실패로 거짓 CAUGHT 였음 — 변이를 `/100` 비교로 바꾸고 하네스가 빌드 실패를
  INVALID 로 따로 세게 고친 뒤 다시 돌림(기억 「뮤테이션은 대상에 닿아야 한다」).
- 단일 증인 변이: M7(조건주문 목록 순서)·M8·M13 은 시험 하나씩만 잡음 — 각각 그 갈래만 낼 수 있는 문구로 단언.

## 3.3 선택 시험 (최종 코드 `c25e2fad0` — 결과는 아래 「3.3 결과」)

- `go test -count=1 ./internal/official/ ./internal/verifylive/ ./cmd/tossctl/` 무태그 · `-tags tossos_testseams`, `make lint`.
  원문 `analysis/verify/selected-tests-notag.txt` · `selected-tests-tossos_testseams.txt` · `make-lint.txt`.
- 미실행: `make test` · `make test-seams` 전체 · `make test-race` · `make gate`(사용자 명시 요청 시에만). 다른 패키지는 바뀐 심볼을 안 씀
  (CodeGraph 1.6.0 — 생산 호출자는 cmd 위임뿐).
- R5 비용(기록용): `analysis/measure/r5_cost-output.txt` — 100행 53 KB ≈ 5 ms.

## 3.2 리뷰 — gstack `review` + 적대 (2026-11-25, 대상 `e57d06cd6..ee029b37e`, 수리 `c25e2fad0`)

- 보이스(1회차): gstack review 전문가 3(Testing · Maintainability+Simplification · Security+Performance — 각자 신선한 문맥, 읽기 전용) ·
  Claude 적대(「Think like an attacker and a chaos engineer」) · Codex 적대(`codex exec -s read-only`, 다른 계열). 2회차: Codex(수리 검증).
  gstack 원 절차에서 줄인 것: 전문가를 5→3 으로 묶음(security+performance, maintainability+simplification), Exploratory QA 미실행
  (대사 CLI 는 실 브로커 없이 못 돌림 — httptest 시험이 그 역할), 리뷰 로그 기록 안 함.
- 필수 질문 「이 change 뒤에도 줄이 써지는 200 응답」: 두 적대 보이스 모두 **해독 층에서는 없음**(Claude 적대: 계약 키 23 × U+0000~U+2FFFF
  단일 대입·삽입 전수에서 `encoding/json` 이 묶지만 `EqualFold` 가 놓치는 키 0, Testing: U+0080~U+1FFFF × a~z 차이 0). 남은 길은 판정 의미의
  기존 공백 하나(F-A1).

| # | 보이스 | 요지 | 처분 |
| --- | --- | --- | --- |
| F-A1 | Claude 적대 · Codex | 만료 모드에서 대상 아닌 CLOSED 행의 `first.status`(예 `WATCHING`·`TRIGGERED`)·`type` 을 안 봄 → top-level `EXPIRED` 만으로 통과 | **Manager 결정 대기**(design §10b-6) — a164 부터의 판정 의미 공백, 해독 모호함 아님 · 범위 밖 |
| F-S1 | Security | 경로 `name` 에 브로커 키가 날것으로 붙어 제어 문자·ANSI 가 운영자 거절 문구로 나감 | **수리 `c25e2fad0`** — `strconv.Quote(key)`, 시험 제어 문자 키 사례 + 변이 M13 |
| F-P1 | Security+Perf · Claude 적대 · Codex | O(본문×깊이) 재토큰화 · 객체당 O(키²) 짝 비교 — 병적 본문에 수 초·수 GB | **잔여 R5 + Manager 결정 대기 §10b-5**(상한·한 번 훑기는 새 가드/구조 변경). 정상 응답 5 ms 실측 |
| F-T1 | Testing | 「해독 뒤 판정」 순서 핀이 조건주문 목록뿐(일반 주문 위치 변이가 살아남음 — 리뷰어 overlay 실측) | **수리** — C5 를 네 리더로, 변이 M7b~d |
| F-T2 | Testing | 200 아닌 2xx 시험이 「오류가 났다」 만 봄 | **수리** — `ErrReconcileSchema` + 상태 문구 단언 |
| F-T3 | Testing | 정확한 중복 판정(`reconcileObject` 호출) 변이가 원장에 없음(리뷰어가 직접 돌려 잡힘 확인) | **수리** — 변이 M12 추가 |
| F-M1 | Maintainability | 계약 키 핀이 손으로 고른 함수 5개만 훑어 새 키 읽기 함수가 생기면 목록이 조용히 낡음 | **수리** — 두 파일의 모든 함수를 훑고 200 아닌 본문 파서(`reconcileDetailSilent`)만 이름으로 뺌(허용된 핀 하나 안에서) |
| F-M2~4 | Maintainability | 시험 헬퍼 중복(`ambWrap`=`wrapResult`, `rcAmbEmptyCond`=`rcEmptyCond`), 안 읽는 `seen` 필드 | **수리** — 기존 헬퍼로 바꾸고 삭제 |
| F-A3 | Claude 적대 | 바깥 관측자 가림(R4) | 잔여 R4(Manager 결정 3 — 새 구조 시험 없음) |
| [ADVISORY] | Maintainability | 시험의 「깨끗한 본문」 지도 두 벌 | 건너뜀(절감 3줄) |

- 2회차(Codex, `ee029b37e..c25e2fad0`): **VERDICT: CLEAN** — 수리 정확, 새 결함·공허 시험 없음, 원장에 컴파일 실패·엉뚱한 이유로 잡힌 변이 없음.
- Gate edits(gstack 정의 — 시험·러너 완화): 해당 없음 — 기존 시험 함수 수정 0, 새 시험 파일 3개만 추가(gate-diff 도구는 안 돌림).

## 3.3 결과 (최종 코드 `c25e2fad0`)

| 명령 | official | verifylive | cmd/tossctl | rc |
| --- | --- | --- | --- | --- |
| `go test -count=1`(무태그) | ok 6.599s | ok 48.489s | ok 92.527s | 0 |
| `go test -count=1 -tags tossos_testseams` | ok 6.292s | ok 51.100s | ok 94.199s | 0 |
| `make lint`(gofmt · `go vet ./...` · 태그 vet) | — | — | — | 0 |

- `check_analysis.py --change a168-verify-read-decode-hardening` rc=0(required 4, 「evidence complete」). `openspec validate --strict` rc=0.
