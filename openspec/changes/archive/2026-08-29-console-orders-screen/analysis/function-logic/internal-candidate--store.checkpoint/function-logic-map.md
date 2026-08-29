# Function Logic Map: `Store.Checkpoint`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. D16의 세 번째 관찰 — WAL은 오래 열린 독자가 있으면 체크포인트되지 않고 콘솔의 후보 화면이 정확히 그런 독자다 — 때문에 존재한다. 행을 지운 저장소가 그 삭제로 돌려받았어야 할 바이트를 원장의 파일시스템 위에서 그대로 쥐고 있을 수 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (인자 없음) | — | — | — |
| `PRAGMA wal_checkpoint(TRUNCATE)` 결과 | busy 플래그, WAL 페이지 수, 체크포인트된 페이지 수 | SQLite | 세 값 모두 NULL 가능하므로 `sql.NullInt64`로 받는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `row.Scan` 에러 | 없음 | `false, 0, wrap` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.QueryRowContext` | `PRAGMA wal_checkpoint(TRUNCATE)` | PASSIVE가 아니라 TRUNCATE — 크기가 요점이다(D16) | ast.json calls |

## State mutations and fallbacks

- WAL 파일을 잘라 디스크 바이트를 돌려준다. 테이블 행은 건드리지 않는다.
- **busy는 에러가 아니다.** 다른 연결이 읽는 동안 TRUNCATE는 완료할 수 없고, 화면 하나가 열려 있다고 스캔이 실패하면 정리가 정리 없음보다 덜 신뢰할 수 있게 된다. `busy` 반환이 '건너뛰었다'를 말해 주므로 호출자가 일어나지 않은 회수를 보고하지 않는다.
- 부분 실패는 없다 — PRAGMA 한 번이다.

## Safety conclusion

- Safe edit boundary: busy를 에러로 승격하거나 TRUNCATE를 PASSIVE로 낮추는 것은 D16의 목적을 지운다
- High-risk impact: no — 발굴 저장소 자신의 WAL만 자른다. 원장 파일에 닿지 않는다. 목적 자체는 원장의 파일시스템을 지키는 방향이다.
