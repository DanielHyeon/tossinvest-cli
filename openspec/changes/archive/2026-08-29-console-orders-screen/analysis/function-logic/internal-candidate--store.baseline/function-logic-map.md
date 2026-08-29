# Function Logic Map: `Store.Baseline`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: base, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**. 뒤에 삽입된 `FirstRank` 절(타입·`Recorded`·`NoteFirstRank`·`FirstRank`·`decodeFirstRank`)의 diff hunk가 이 함수와 교차해 evidence가 요구되었다. base L1051-1068의 본문은 현재 L1242-1259와 byte 동일하고, ast.json은 base revision이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` / `symbol` | 대문자화·trim | 호출자 | 정규화가 읽기·쓰기 양쪽에 있어야 패딩된 값이 조용히 0행을 고르지 않는다 |
| `first_price` 컬럼 | TEXT NULL | `NoteFirstPrice` 또는 v2 마이그레이션 백필 | 십진값은 TEXT이고 NULL은 '원천이 그 필드를 나르지 않았다'이다. 0 기본값을 가진 수치 컬럼이었다면 '원천이 0이라고 말했다'와 '아무도 재지 않았다'가 합쳐지고, D10은 그 차이가 veto까지 살아남는 것 위에 서 있다. |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `row.Scan` 결과 switch | — | — | `TestAnAbsentBaselineIsNotAZeroOne` |
| B2 | `sql.ErrNoRows` — 후보 자체가 없음 | 없음 | `Baseline{}, false, nil` | `TestAnAbsentBaselineIsNotAZeroOne` |
| B3 | 그 밖의 에러 | 없음 | `Baseline{}, false, wrap` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.QueryRowContext` | 세 컬럼 단건 조회 | 단일 문장 | ast.json calls |
| `decodeBaseline` | NULL·공백을 미기록으로, 나머지는 값·순간·원천으로 | 읽을 수 없는 `first_price_at`은 에러 | 같은 파일 |

## State mutations and fallbacks

- 상태 변경 없음 — 읽기 전용이다.
- 두 번째 반환이 '그런 후보 없음'과 '기준선 없는 후보'를 가른다. 둘 다 미기록 `Baseline`을 내지만 하나만 수명에 관한 질문이다.
- 본문 무변경이므로 이 change가 만든 동작 변화는 없다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재
- High-risk impact: no — 읽기 전용이고 본문 byte 동일. 다만 반환값이 `extended` veto의 분모라 부재를 0으로 접는 방향만 위험하다.
