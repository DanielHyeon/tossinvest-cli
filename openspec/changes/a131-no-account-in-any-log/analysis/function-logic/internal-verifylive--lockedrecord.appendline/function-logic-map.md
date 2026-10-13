# Function Logic Map: `lockedRecord.appendLine`

- Source: `internal/verifylive/record.go` (467-501)
- Qualified: `lockedRecord.appendLine`
- AST evidence: `ast.json` (편집 뒤, `source_sha256` 9a289d21048578d8…) — AST branches 9 · returns 10
- Risk scan: `risk-pattern-report.md`
- 작성기: `analysis/harness/flm_write.py`(좌표는 ast.json 에서 옮김)

**역할.** 대사 줄 하나를 잠근 fd 로 한 번에 쓰고 동기화한 뒤 다시 읽어 대조한다(호출자 `Reconcile`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `line` | 개행으로 끝나는 줄 | encodeReconcileLine | 아니면 오류(B1) |
| `judged` | 판정한 바이트 길이 | Reconcile | 크기가 다르면 errRecordSizeMoved |

## Branches and early returns

| Branch | AST kind | Source line | a131 |
|---|---|---|---|
| B1 | if at 468:2 | `if len(line) == 0 \|\| line[len(line)-1] != '\n' {` | 무편집 |
| B2 | if at 473:2 | `if err != nil {` | 편집 영향 |
| B3 | if at 476:2 | `if line, err = redact.SealRecordLine(line, key); err != nil {` | 편집 영향 |
| B4 | if at 480:2 | `if err != nil {` | 무편집 |
| B5 | if at 483:2 | `if st.Size() != judged {` | 무편집 |
| B6 | if at 487:2 | `if err != nil \|\| n != len(line) {` | 무편집 |
| B7 | if at 490:2 | `if err := recordSync(l.f); err != nil {` | 무편집 |
| B8 | if at 494:2 | `if _, err := l.f.ReadAt(back, judged); err != nil && !errors.Is(err, io.EOF) {` | 무편집 |
| B9 | if at 497:2 | `if !bytes.Equal(back, line) {` | 무편집 |

Exact AST return positions: 469:3, 474:3, 477:3, 481:3, 484:3, 488:3, 491:3, 495:3, 498:3, 500:2

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `redact.SealRecordLine`(a131 신설) | account_ref(마스크) 봉인 + 치환 | 키 생성 실패면 오류 — 쓰기 전 | design D4 |
| `recordWrite`·`recordSync`·`ReadAt` | 쓰기·대조 | 불완전·미확인 오류 | AST calls |

## State mutations and fallbacks

- 파일 끝에 한 줄. a131: B1 검사 뒤 봉인한 줄을 쓰고 대조한다(대조 대상도 봉인 줄).

## Safety conclusion

- Safe edit boundary: B1 뒤 봉인 한 줄 + 오류 반환. 크기 판정·쓰기·대조 무편집.
- High-risk impact: yes(간접) — 검증 대사 기록. 동시성 지문은 원본 바이트 그대로.
