# Function Logic Map: `newlyListedAt`

- Source: `internal/candidate/veto.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `base`, L694–702, 분기 2개)
- Risk scan: `risk-pattern-report.md`

**이 함수는 더 이상 존재하지 않는다.** 커밋 `515c658`이 삭제했고, 여기 실린 `ast.json`은
change의 base commit `b268593`에서 손으로 추출한 것이다(`"revision": "base"`). 선례는
`add-candidate-discovery`의 `cmd-tossctl--consolemutationconfirmer/ast.json`이다.

## 무엇을 하던 함수였나

저장된 최초 관측 위치에 대해 소스의 `newly_listed` 플래그를 되찾으려 했다. 관측 슬라이스를
훑어 (source, instant, rank, total) **넷이 모두 일치하는** 행을 찾고 그 행의 bool을
돌려줬다. 못 찾으면 `false`.

## 왜 사라졌나

그 조회는 **작동할 수 없었다**. `Assess`는 최근 `DefaultAssessHistory`(10분)의 관측만
읽는다. 최초 관측 행은 후보 수명 대부분의 시점에서 그 창 밖이므로, 10분보다 오래된 **모든**
후보에 대해 일치하는 행이 슬라이스에 없다. 답은 하루 종일 `false`였다.

그리고 `false`가 두 가지를 뜻했다 — "소스가 아니라고 했다"와 "우리가 그 행을 더 이상 갖고
있지 않다". bool에는 세 번째 철자가 없었다. 다섯 층에 선언된 사실이 한 번도 기록되지 않은
채 출하된 두 번째 이유가 여기다(첫 번째는 어떤 생산 소스도 그 필드를 대입하지 않았다는 것).

이것은 D20이 이미 한 번 고친 붕괴의 재발이다 — prunable 테이블보다 오래 살아야 하는 사실을
그 테이블에서 읽으려 한 것.

## 무엇이 그 질문에 답하는가

`candidates.first_rank_newly_listed` **칼럼**이다. `NoteFirstRank`가 위치와 **같은 문장에서**
쓰고, `decodeFirstRank`가 3-상태로 복원하며, `MeasureFirstSighting`이 `first.NewlyListed`로
직접 읽는다. D17(`first_price_at`)과 D20(`first_rank_at`)이 이미 두 번 내린 것과 같은
결정이다(issues.md I1).

그리고 답은 이제 **소비된다**. 예전 주석은 "미스는 false로 두며, 그것으로 아무것도 결정되지
않는다"고 적었다. 지금은 `unknown`이 `NEW_ENTRANT_UNKNOWN`이라는 명명된 거부를 만든다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `observations` | `Assess`가 넘긴 최근 10분 | `ObservationsSince` | **최초 관측 행은 대개 여기 없다** — 이 함수가 삭제된 이유 |
| `first` | 저장된 위치 | `candidates.first_rank*` | 네 필드가 모두 일치해야 매치 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, o := range observations` | 없음 | — | 삭제됨 — 대체 경로는 `Store.FirstRank`의 칼럼 읽기 |
| B2 | source·instant·rank·total 4중 일치 | 없음 | 일치하면 그 행의 bool, 아니면 루프 끝에 `false` | 삭제됨 |

B2의 4중 일치는 신중한 설계였다 — 한 스캔이 한 심볼의 모든 소스 행에 같은 instant를
찍으므로 instant만으로는 여러 행이 선택된다. 문제는 정밀도가 아니라 **슬라이스에 행이
없다는 것**이었다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.ObservedAt.Equal(first.At)` | instant 일치 | 순수 | ast.json calls (base) |

## State mutations and fallbacks

- 없었다 — 순수 조회였다.
- fallback이 있었고 그것이 결함이었다: 미스가 `false`로 접혔고, `false`는 '측정된 아니오'와 구분되지 않았다.

## Safety conclusion

- Safe edit boundary: **삭제**. 호출부는 `MeasureFirstSighting` 한 곳이었고 저장 칼럼 읽기로 대체되었다. `rg 'newlyListedAt'` 0건.
- High-risk impact: no — 이 함수는 아무것도 결정하지 않았다(주석이 그렇게 적었고 사실이었다). 그것을 대체한 칼럼 경로는 결정한다: `unknown`이 `seen_late`를 명명된 미측정으로 만든다.
