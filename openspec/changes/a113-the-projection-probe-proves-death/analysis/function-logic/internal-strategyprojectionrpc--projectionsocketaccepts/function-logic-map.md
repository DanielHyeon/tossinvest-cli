# Function Logic Map: `projectionSocketAccepts`

- Source: `internal/strategyprojectionrpc/transport_unix.go`
- AST evidence: `ast.json` — **구현 후 재생성**(:393–403, 분기 2·반환 3). 편집 전 base `54004f44` 는 :387–400, 분기 3·반환 4·호출 7 이었다.
- 구현 후 AST 대조: B3(owner-write 추정, 편집 전 :396)과 그 `os.Lstat`·`Perm` 호출이 사라졌고 B1(성공)·B2(거부·부재)·종단(생존)은 그대로다.
- Risk scan: `risk-pattern-report.md`

a113 이 이 함수에서 하는 일은 **절 하나를 지우는 것**이다(B3 = owner-write 사망 추정).
chmod-then-probe 는 이 함수가 아니라 회수 전용 새 함수(`staleProjectionSocketAccepts`, 새 파일)가
진다 — 이 함수는 `Dial`(소비자)도 부르므로, 여기에 chmod 를 넣으면 **조회 클라이언트가 엔진의
socket 권한을 바꾸는** 새 부작용이 생긴다(design D1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `socketPath` | control 디렉터리 안 socket 경로 | 호출자 2곳(`reclaimStaleControlDirectory` :293, `Dial` :417 — CodeGraph callers) | 경로 오류는 B2/B3/종단으로 분류 |
| 커널의 connect 답 | 성공 · ECONNREFUSED · ENOENT · 그 밖(EACCES·timeout·ENOTDIR) | `net.DialTimeout`(200ms) | B1/B2/종단 |
| 파일 권한 | 임의 | `os.Lstat` (B3 에서만) | Lstat 실패면 B3 불성립 → 종단(true) |

불변식: 사망(false)으로 읽는 것이 틀리면 **산 주인의 socket 이 지워진다**(회수) — 그래서 모르는 것은
생존(true)으로 읽는다. B3 는 이 불변식의 예외로 **추정**을 사망으로 읽는 유일한 절이었다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `err == nil` (연결 성공) | probe 연결 1회 열고 닫음 | `true` (:391) | `TestProjectionLivenessClausesEachDecideOnTheirOwn/수락한다` |
| B2 | `errors.Is(err, ECONNREFUSED) \|\| errors.Is(err, ErrNotExist)` | 없음 | `false` (:394) | `…/경로가_없다` · `…/아무도_수락하지_않는다` |
| B3 | `os.Lstat` 성공 && `perm&0o200 == 0` (owner 쓰기 없음) | Lstat 1회 | `false` (:397) — **a113 이 지우는 절** | 편집 전: `…/owner_쓰기_비트가_없다`(0500 죽은 socket=false). 편집 후: `…/쓰기_비트가_깎여도_수락_중이면_생존`(0400 산 socket=true) |
| 종단 | 그 밖의 오류 | 없음 | `true` (:399) | `…/죽었다는_증거가_아닌_오류`(ENOTDIR) |

B3 의 결함(a109 A1 P1-A 원형): 수락 중인 socket 의 owner 쓰기 비트가 외부 chmod 로 깎이면 connect 는
EACCES 로 실패하고, B3 은 그것을 **사망**으로 읽는다 → 회수가 산 주인의 socket 을 지운다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `net.DialTimeout("unix", …, projectionProbeTimeout)` | 주인의 생존 질문 | 200ms, 재시도 없음 | AST call :388 |
| `conn.Close` | probe 연결 반납 | 오류 버림 | AST :390 (CodeGraph 는 이것을 `internal/journal/readonly_schema_arms_test.go:51` 의 `Close` 로 오해석 — evidence-reconciliation.md) |
| `errors.Is` ×2 | 사망 오류 분류 | — | AST :393 |
| `os.Lstat` + `Mode().Perm()` | B3 의 권한 추정 | — | AST :396 — a113 이후 사라진다 |

## State mutations and fallbacks

- 디스크를 바꾸지 않는다(편집 전·후 동일). a113 이후에도 이 함수는 **순수 질문**이다 — chmod 는
  회수 전용 함수의 것이다.
- 편집 후 EACCES 는 종단(true)으로 간다. `Dial` 쪽 영향: `Dial` 은 B2(:409)에서 **정확-0600** 을
  먼저 요구하므로(`internal-strategyprojectionrpc--dial/ast.json` B2), 0600 socket 은 owner 쓰기가
  있어 B3 에 닿지 않는다. 차이는 Lstat(:408)과 connect(:417) 사이에 권한이 바뀌는 경합에서만
  생기고, 그때 `Dial` 은 "no listener" 대신 client 를 돌려주고 첫 Read 가 실패한다 — 소비자
  재부착 wrapper(a109 D4)가 그 실패를 받는다.

## Safety conclusion

- Safe edit boundary: B3 한 절 삭제 + 주석 정정. B1·B2·종단과 서명은 그대로.
- High-risk impact: no(주문·손절 경로 아님) — 단 엔진 boot 경로의 회수 기계이므로 Pre-Edit 선언과
  RED 선행을 한다. 방향은 보수적이다: 사망으로 읽는 경우가 **줄어든다**(추정 1개 제거).
