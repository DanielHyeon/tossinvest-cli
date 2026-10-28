# Function Logic Map: `TestRecordCarriesNoUnmaskedAccountNumber`

- Source: `internal/verifylive/record_test.go` (156-182, 편집 뒤)
- AST evidence: `ast.json`(revision current, 편집 뒤 소스 sha `991191491530…`) — AST branches 6(B1~B6, 전부 `if`) · returns 0.
  편집 전 판본(분기 5, sha `fe4b744224c3…`)은 커밋 `3bf5c6043` 의 이 번들에 있음 — 편집 전에 만들어 커밋함.
- Risk scan: `risk-pattern-report.md`(규칙 일치 0)
- 작성 시점: a164 종결(2026-10-28, 종결 에이전트). 사용자 승인 「a164 A」(Manager 전달): 종결을 막는 이 시험의 우연 일치를 같은 브랜치에서 근본 수정.
  이 함수는 a131(`f269c6157`)이 마스크 검사를 더한 시험이고 a164 범위 밖임.

**역할.** 가짜 브로커로 검증 한 판을 돌린 뒤 기록 파일(JSONL)에 계좌 정보가 평문으로 없는지 본다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 가짜 계좌 | `123-45-678901`(마스크 `maskedAccount` = `attest.Mask` = `*********8901`) | harness `opts.AccountRef`(`fake_broker_test.go`) | — |
| 기록 파일 | 줄마다 JSON 객체 하나, 계좌는 `account_ref_sealed`(`v1.` + base64url 암호문) | `record.go` `Recorder.Append` → `redact.SealRecordLine` | 읽기 실패 시 B2 Fatal |
| 실행마다 바뀌는 값 | `sha256:` + 16진 digest(`calls[].request_digest`·`response_digest`·`observations[].value`), 봉인 nonce, `process.pid`(숫자) | 생산 코드 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `h.run` 오류 | 로그만 | 계속(오류여도 기록은 검사) | 이 시험 |
| B2 | 기록 파일 읽기 실패 | — | `t.Fatalf` | 이 시험 |
| B3 | 원 바이트에 평문 계좌 `123-45-678901` | — | `t.Error`(불변) | 이 시험 |
| B4 | 원 바이트에 마스크 전체 `*********8901` | — | `t.Error` 마스크 평문 — `*` 는 16진·base64url·시각·숫자에 없어 우연 일치 불가 | 이 시험 · 변이 M1 |
| B5 | `recordStringsContaining(data, "8901")` 이 하나라도 찾음 | — | `t.Errorf` 꼬리 평문(찾은 키=값 목록) | 이 시험 · 변이 M1 · M2 |
| B6 | 열린 기록의 `AccountRef` 가 `8901` 로 안 끝남 | — | `t.Errorf`(불변) | 이 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newFakeBroker().withHolding`, `newHarness`, `h.run` | 가짜 브로커로 한 판 실행 | 오류는 B1 로그 | AST calls |
| `os.ReadFile` | 원 바이트 | B2 | AST calls |
| `maskedAccount`(`steps.go:1353`, `attest.Mask`) | 생산 코드가 쓰는 마스크와 같은 글자 | — | AST calls |
| `recordStringsContaining`(새 파일 `record_mask_leak_test.go`) | 줄마다 JSON 을 풀어 키·문자열 값에서 찾음. 값 전체가 `^sha256:[0-9a-f]+$` 이거나 키가 `account_ref_sealed` 인 값만 건너뜀, 숫자 값은 안 봄, 못 푸는 줄은 원 바이트로 판정 | 스캔 오류는 Fatal | AST calls |
| `LoadEntries` | 봉인을 열어 읽기 | B6 | AST calls |

## State mutations and fallbacks

- 시험 임시 디렉터리 안의 기록만 씀. 공유 상태 없음.
- 고친 결함(측정): 편집 전 B4 는 원 바이트에서 `8901` 을 찾아, 실행마다 바뀌는 16진 digest 안의 우연 일치로 5/3000 실패했음(일치 자리 전부 digest 안).
- 건너뛰는 모양이 계좌를 실을 수 없는 이유: 해시 digest 는 해시 출력(16진만), `account_ref_sealed` 는 AES-GCM 암호문. 숫자 값은 pid·지연·수량이고 계좌 참조 필드는 문자열임.

## Safety conclusion

- Safe edit boundary: 시험 함수 하나 + 새 시험 헬퍼 파일. 생산 코드 무변경.
- 약화 아님: 원래 의도(평문 계좌·마스크가 기록에 들어가면 FAIL)는 B3·B4·B5 가 유지. 변이 M1(봉인 생략 — 마스크가 `account_ref` 평문)은 B4·B5 가, M2(꼬리만 `title` 에 평문)는 B5 가 잡음.
- High-risk impact: no — 시험 파일.
