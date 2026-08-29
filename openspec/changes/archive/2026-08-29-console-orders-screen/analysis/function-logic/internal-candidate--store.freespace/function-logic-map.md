# Function Logic Map: `Store.FreeSpace`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. `Open`이 찾은 것을 재사용하지 않고 **다시 probe한다** — 여유 공간은 프로세스가 도는 동안 바뀌는 유일한 파일시스템 속성이고, D16이 발굴을 멈추게 하는 바로 그 값이다. `Open`의 FSInfo는 '여기 써도 되는가'이고 이것은 '남은 자리가 있는가'다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.prober` | `Open`이 보관한 prober | `Options.FSProber` 또는 `SystemFSProber()` | nil이면 `ErrProbeUnsupported` |
| `filepath.Dir(s.path)` | 저장소가 있는 디렉터리 | `Open` | 원장과 같은 디렉터리·같은 파일시스템이다(D2) |
| `info.FreeMeasured` | 부재-대-0 플래그 | prober | false면 **에러**다. 못 잰 공간은 있는 공간이 아니다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `s.prober == nil` | 없음 | `0, ErrProbeUnsupported` | 직접 테스트 없음 — `Open`이 항상 채우므로 제로값 `Store`에서만 도달 |
| B2 | `Probe` 에러 | 없음 | `0, wrap` | `TestSpaceItCouldNotMeasureIsNotSpaceItHas` |
| B3 | `!info.FreeMeasured` | 없음 | `0, ErrProbeUnsupported` | 직접 테스트 없음 — 비-Linux `unsupportedProber`와 `Bsize<=0`의 프로덕션 경로 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.prober.Probe` | 잔여 바이트 재측정 | 실패는 그대로 올린다. watch 사이클은 그것을 **정지**로 읽는다 | `fsprobe_linux.go` |
| `filepath.Dir` | 저장소 디렉터리 | — | ast.json calls |

## State mutations and fallbacks

- 상태 변경 없음 — 읽기 전용 probe다.
- 세 실패가 전부 에러다. '못 쟀다'를 '넉넉하다'로 접는 것이 발굴이 원장의 파일시스템을 채우게 두는 실수이고, 0으로 접는 것은 아무도 재지 않은 값의 가장 놀라운 읽기다. 둘 다 안전하지 않으므로 호출자가 결정한다 — `Budget.Reported`와 같은 규칙.
- watch 사이클이 이 실패를 만나면 원천을 읽지도, 관측을 쓰지도 않고 사유와 함께 멈춘다.

## Safety conclusion

- Safe edit boundary: 실패를 '넉넉함'으로 흡수하거나 `FreeMeasured=false`를 0바이트로 내보내는 것은 D16의 방향을 뒤집으므로 금지
- High-risk impact: no — 읽기 전용 probe다. 다만 **원장이 쓰는 파일시스템**의 잔여 공간이 입력이고, 이 함수가 '넉넉하다'고 잘못 답하면 다음에 ENOSPC를 받는 write가 원장의 것 — 주문·체결·대사 — 이 된다. 그래서 세 실패 모두 에러로 닫혀 있다.
