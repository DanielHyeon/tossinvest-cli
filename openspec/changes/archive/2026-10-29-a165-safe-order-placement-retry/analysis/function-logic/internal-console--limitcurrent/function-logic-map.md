# Function Logic Map: `limitCurrent`

- Source: `internal/console/settings_limits_test.go` (597-608) — base `fd6afe23b` 판, 현재 트리 598-609 에 같은 본문
- AST evidence: `ast.json` (`revision: base`, `source_sha256` 5524583f2caa…) — AST branches 2 · returns 1
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `git show 0d168e0a3:internal/console/settings_limits_test.go | sha256sum`(revision base) = `5524583f2caa…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 작성 시점: a172 종결(2026-10-28, 종결 에이전트). 사후 작성한 증거임 — FLM 이 구현 판단에 쓰이지 않음(review.md 「gate ⑤」).
- 같은 이름의 a174 아카이브 번들도 `revision: base`, sha `5524583f2caa…` 로 a172 base 판과 sha 는 같지만 **복사하지 않음**: 그 ast.json 의 좌표(598-609, 분기 601·605)는 sha `5524583f…` 판의 실제 좌표(597-608, 분기 600·604 — `git show fd6afe23b:internal/console/settings_limits_test.go` 로 다시 뽑아 확인)와 한 줄 어긋남(현재 판 좌표로 뽑은 뒤 sha·revision 만 base 로 바꾼 것으로 보임). 그래서 base 판에서 ast 를 새로 뽑고 시험 이름을 채움.

**역할.** 시험 헬퍼. 설정 화면 HTML 에서 현재 한도 표(`id="guardian-current"` ~ `</table>`)만 잘라 반환.

**본문 무변경.** base `fd6afe23b` 판(597-608)과 줄 단위로 같음. a174 `aa2864abe` 가 바로 뒤에 새 함수를 붙인 hunk `@@ -608,0 +610,43 @@` 가 함수 끝에 닿아 gate ⑤ 가 수정 함수로 셈. a172 의 편집은 0. gate ⑤ 는 이 hunk 가 base 쪽에만 닿으므로 `revision: base` 증거를 요구함 — ast.json 은 base 판 파일(`git show fd6afe23b:<file>`)에서 뽑음. 현재 판 본문도 같음(줄 이동만).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `page` | 일일 설정 렌더 HTML | 설정 템플릿 | 표 부재·미닫힘 → Fatal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 표 시작 없음 | — | Fatal | 호출 시험(실패 갈래) |
| B2 | 표 닫힘 없음 | — | Fatal | 호출 시험(실패 갈래) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Index` | 표 경계 | -1 → B1/B2 | AST calls 599·603 |
| `t.Helper` · `t.Fatal` | 호출 지점 · 계약 위반 | 시험 중단 | AST calls 598·601·605 |

## State mutations and fallbacks

- 입력 문자열의 부분 범위만 반환. 저장·주문·토글 호출 0.

## Safety conclusion

- Safe edit boundary: 시험 헬퍼, 본문 무변경.
- High-risk impact: no — 시험 파일.
