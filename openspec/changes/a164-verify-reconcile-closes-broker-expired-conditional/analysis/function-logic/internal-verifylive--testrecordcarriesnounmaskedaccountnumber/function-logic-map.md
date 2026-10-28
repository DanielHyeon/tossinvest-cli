# Function Logic Map: `TestRecordCarriesNoUnmaskedAccountNumber`

- Source: `internal/verifylive/record_test.go` (156-177, 편집 전)
- AST evidence: `ast.json`(revision current, 편집 전 소스 sha `fe4b744224c3…`) — AST branches 5(B1~B5, 전부 `if`) · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: a164 종결(2026-10-28, 종결 에이전트) — **편집 전**. 사용자 승인 「a164 A」(Manager 전달): 종결을 막는 이 시험의 우연 일치를 같은 브랜치에서 근본 수정.
  이 함수는 a131(`f269c6157`)이 마스크 검사(B4)를 더한 시험이고 a164 범위 밖임.

**역할.** 가짜 브로커로 검증 한 판을 돌린 뒤 기록 파일(JSONL)에 계좌 정보가 평문으로 없는지 본다. 원래 계좌 `123-45-678901`(B3)과
마스크 꼬리 `8901`(B4)을 **원 바이트 전체**에서 부분 문자열로 찾고, 봉인을 열면 마스크가 나오는지(B5) 확인한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 가짜 계좌 | `123-45-678901`(마스크 `attest.Mask` = `*********8901`) | `fake_broker_test.go` 의 harness `opts.AccountRef` | — |
| 기록 파일 | 줄마다 JSON 객체 하나, `account_ref` 는 봉인되어 `account_ref_sealed`(`v1.` + base64url) | `record.go` · `redact.SealRecordLine` | 읽기 실패 시 B2 Fatal |
| 기록 안의 실행마다 바뀌는 값 | `calls[].request_digest`·`response_digest`·`observations[].value` 의 `sha256:` + 16진 32자, 봉인 값(무작위 nonce), `process.pid` | 생산 코드 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `h.run` 오류 | 로그만 | 계속(오류여도 기록은 남아야 함) | 이 시험 |
| B2 | 기록 파일 읽기 실패 | — | `t.Fatalf` | 이 시험 |
| B3 | 원 바이트에 `123-45-678901` | — | `t.Error` 평문 계좌 | 이 시험 |
| B4 | 원 바이트에 `8901` | — | `t.Error` 마스크 평문 — **결함: 16진 digest 와 우연 일치(5/3000 실측)** | 이 시험 |
| B5 | 열린 기록의 `AccountRef` 가 `8901` 로 안 끝남 | — | `t.Errorf` | 이 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newFakeBroker().withHolding`, `newHarness`, `h.run` | 가짜 브로커로 한 판 실행 | 오류는 B1 로그 | AST calls |
| `os.ReadFile` | 원 바이트 | B2 | AST calls |
| `strings.Contains`, `strings.HasSuffix` | 판정 | — | AST calls |
| `LoadEntries` | 봉인을 열어 읽기 | 오류는 B5 에 합침 | AST calls |

## State mutations and fallbacks

- 시험 임시 디렉터리 안의 기록만 씀. 공유 상태 없음.
- 결함 기전(측정): 실패 판 출력의 일치 자리가 전부 `sha256:` digest 안(예 `…a604f74635bf9a0d128901a3b9…`). 단독 `-count=3000` → 5 실패.

## Safety conclusion

- Safe edit boundary: 시험 함수 하나. 생산 코드 무변경.
- 편집 계획: B4 를 둘로 나눈다 — (a) `attest.Mask` 가 만드는 마스크 전체(`*` 포함)를 원 바이트에서 찾음(`*` 는 16진·base64url·시각·숫자에 없음), (b) 꼬리 `8901` 은 JSON 을 풀어 키와 문자열 값에서 찾되,
  값 전체가 해시 digest(`sha256:` + 16진)이거나 봉인 봉투(`account_ref_sealed`)인 것만 건너뜀. 둘 다 원래 의도(마스크가 평문으로 기록에 들어가면 FAIL)를 유지함 — 약화 아님.
- High-risk impact: no — 시험 파일.
