# Function Logic Map: `ValidatePrivateControlDirectory`

- Source: `internal/positionpolicyrpc/private_endpoint.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이름을 보지 않는 control 디렉터리 검증이다. 이름은 호출자가 가지고 검사는 여기가
가진다(파일 머리 주석). a109의 회수 기계는 **이 함수를 그대로 호출한다** — 디렉터리
위생을 복사하지 않기 위해서다(design D1b: 복사한 검사는 어긋나기 시작한 검사다).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 공백 제거 후 비어 있지 않은 경로 | 호출자(`statPrivateSocket`, `OpenPrivateDescriptorFile`, engine `publishPrivateDescriptor`) | 빈 문자열이면 파일시스템 접근 없이 즉시 error |
| `filepath.Dir(path)` (엔진 디렉터리) | 비symlink · group/other 쓰기 없음 · 우리 uid | `ValidateEngineDirectory` → `validatePrivateDirectory(_, false)` | 그 error를 그대로 전달 |
| leaf 디렉터리 | 실제 디렉터리 · **정확 0700** · 우리 uid · symlink traversal 없음 | `validatePrivateDirectory(clean, true)` | error 반환 — a109도 완화하지 않는다(design D2: 디렉터리 perm 완화는 두지 않는다, P1-7①) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `strings.TrimSpace(path) == ""` | 없음 | `errors.New("private endpoint: control directory path is empty")` | 빈 경로 거부(현재 직접 커버 없음 — 호출자가 항상 `filepath.Dir` 결과를 넘긴다) |
| B2 | `ValidateEngineDirectory(filepath.Dir(clean)) != nil` | 없음 | 부모의 error 그대로 | group-writable 엔진 디렉터리에서 기동 거부 |
| 분기 밖 종단 | 위 둘 통과 | 없음 | `validatePrivateDirectory(clean, true)`의 결과 | 0700이 아닌 control 디렉터리 거부 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 경로 정규화 | 오류 없음 | AST calls |
| `filepath.Dir` | 부모 경로 산출 | 오류 없음 | AST calls |
| `ValidateEngineDirectory` | 부모 디렉터리 위생 | error 전파, 재시도 없음 | AST calls |
| `validatePrivateDirectory(_, true)` | leaf 정확-0700·소유·비symlink | error 전파 | AST calls |

## State mutations and fallbacks

- 상태 변경 없음. 순수 검증이고 fallback이 없다 — 실패는 곧 호출자의 실패다.
- 재시도 없음. 같은 디스크 상태에서 같은 답을 낸다(결정적) — a109 D3b의 "강등의 잔여
  원인은 결정적"이라는 논거가 여기서 나온다.

## Safety conclusion

- Safe edit boundary: **편집하지 않는다.** a109는 호출만 한다(회수 기계가 자기 디렉터리
  검사로 재사용). 완화하면 회수뿐 아니라 descriptor 열기·발행 전 검증까지 함께 넓어진다.
- High-risk impact: yes — 엔진 기동 경로의 첫 관문이고, 거부는 endpoint 부재로 이어진다.
