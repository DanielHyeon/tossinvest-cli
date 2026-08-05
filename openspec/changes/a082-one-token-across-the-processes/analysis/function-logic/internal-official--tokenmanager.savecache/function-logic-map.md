# Function Logic Map: `tokenManager.saveCache`

- Source: `internal/official/token.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**High-risk.** 이 함수가 쓰는 파일을 세 보유자가 공유하고, a082는 **다른 보유자가
방금 썼을 때** 정확히 그것을 읽는다. 부분 쓰기가 보이는 창이 이 change로 넓어진다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ct` | 유효한 `cachedToken` | `exchange()` | — |
| `m.cacheFile` | 공유 파일 경로 | `apppaths.OpenAPI` | 디렉터리 생성 실패는 오류 |
| 불변식 1 | **파일은 항상 온전한 토큰 하나를 담는다** | rename의 원자성 | 어기면 읽는 쪽이 토큰 없음으로 판단해 하나 사고 쓴 쪽의 토큰을 죽인다 |
| 불변식 2 | 모드는 0600 | `Chmod` | 어기면 자격증명 파생값이 노출된다 |
| 불변식 3 | 호출자는 오류를 무시한다 (best-effort) | `exchange()`의 `_ =` | 쓰기가 계속 실패하면 이 프로세스는 요청마다 401→교환을 반복한다 → issues I4 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `MkdirAll` 실패 | 없음 | 그 오류 | 기존 |
| B2 | `json.Marshal` 실패 | 없음 | 그 오류 | 도달 불가 (구조체가 고정) |
| B3 | 임시 파일 생성 실패 | 없음 | 그 오류 | 기존 커버리지 |
| B4 | `Chmod` 실패 | 임시 파일 남음 → `defer`가 지운다 | 그 오류 | 기존 커버리지 |
| B5 | 쓰기 실패 | 임시 파일 남음 → `defer`가 지운다 | 그 오류 | 기존 커버리지 |
| B6 | `Close` 실패 | 임시 파일 남음 → `defer`가 지운다 | 그 오류 | 기존 커버리지 |
| (성공) | — | **`Rename`** — 원자적 교체 | nil | `TestAReaderNeverSeesAHalfWrittenCacheFile` |

**임시 파일 + rename이 편집이다.** base는 `os.WriteFile`이었고 그것은 truncate 후
write다. 같은 디렉터리에 임시 파일을 만드는 것이 중요하다 — rename이 원자적이려면
같은 파일시스템이어야 한다.

`defer os.Remove(name)`은 rename 성공 뒤에는 no-op이다 (그 이름이 더 이상 없다).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.MkdirAll` | 디렉터리 보장 (0700) | 실패는 오류 | AST calls |
| `os.CreateTemp` | 같은 디렉터리의 임시 파일 | 실패는 오류 | AST calls |
| `temporary.Chmod` | 0600 — `CreateTemp`은 0600을 보장하지 않는다 | 실패는 오류 + 정리 | AST calls |
| `os.Rename` | 원자적 교체 | 같은 파일시스템이라야 원자적 | AST calls |

## State mutations and fallbacks

- 프로세스 밖 side effect가 이 함수의 전부다.
- 실패하면 **이전 파일이 그대로 남는다** — base는 실패 시 잘린 파일을 남길 수
  있었다. 안전 방향의 변화다.
- fallback 없음. 호출자가 오류를 무시한다.

## Safety conclusion

- Safe edit boundary: **쓰기 방식뿐.** 경로 계산, 디렉터리 모드(0700), 파일
  모드(0600), 반환 계약은 그대로다.
- 임시 파일은 **같은 디렉터리**에 만든다. `/tmp`에 만들면 rename이 파일시스템을
  넘어 실패하거나 복사가 되어 원자성을 잃는다.
- 임시 파일 이름은 `.`으로 시작해 목록에서 눈에 덜 띄고, 실패 경로마다 지운다.
- High-risk impact: **yes.** 이 파일이 자격증명 파생값을 담고, 그 내용의 온전함에
  세 프로세스의 토큰 수렴이 걸려 있다.
