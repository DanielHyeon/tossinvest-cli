# Function Logic Map: `writeDescriptor`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (466-528)
- AST evidence: `ast.json` — revision `current`. AST 분기 13
- Risk scan: `risk-pattern-report.md`

Fix 라운드가 이 함수를 다시 썼다(design D1-2). 예전 본문은 세 줄이었다 — 최종 이름에
`O_EXCL`로 만들고, 거기에 write하고, sync했다. 그 셋은 원자적이지 않으므로 사이에서
죽으면 **최종 이름을 가진 0바이트·잘린 JSON**이 남았고, 회수의 내용 검증이 그것을
영구 거부했다(A1 F2 실측 3/3). 분기는 3 → 12로 늘었다: 의례의 각 단계가 분기다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | `…/.strategy-runtime-read/endpoint.json` | 호출부(`Start`) | 디렉터리는 이미 0700으로 만들어져 있다 |
| `descriptor` | 스키마 v1 · socket 이름 · 토큰 · PID | `Start` | B1 marshal 실패 |
| 임시 파일 | `.staging-endpoint-*`, 0600 | `os.CreateTemp` + `Chmod` | B2·B3 — 실패해도 임시 이름만 남고 그것은 회수 대상이다 |
| 발행 | 같은 디렉터리 안 rename | `os.Rename` | B9 — 원자적. 최종 이름은 완성된 파일에만 붙는다 |
| 내구성 | 파일 sync + 디렉터리 sync | `File.Sync`, `Dir.Sync` | B6·B12 — 전원이 끊겨도 rename이 살아남는다 |

### `O_EXCL`이 사라진 자리 (선언된 약화와 그 근거)

예전 코드는 `O_EXCL`로 "이미 있으면 만들지 않는다"를 보장했다. rename은 덮어쓴다.
이 약화가 안전한 이유는 호출 시점 때문이다: `Start`는 방금 `Mkdir`한 **빈 디렉터리**
안에서만 이 함수를 부른다(디렉터리가 이미 있었으면 회수가 통과한 뒤 다시 만든다).
덮어쓸 최종 이름이 애초에 없다. 같은 저장소의 다른 세 endpoint가 쓰는 의례와 같다
(`internal/app/engine`의 `publishPrivateDescriptor`).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | JSON marshal 실패 | 없음 | 그 오류 | 없음 |
| B2 | 임시 파일 생성 실패 | 없음 | 그 오류 | 없음 |
| B3 | 임시 파일 `Chmod` 실패 | 임시 파일 제거(defer) | 그 오류 | 없음 |
| B4 | 임시 파일 `Stat` 실패 | 임시 파일 제거 | 그 오류 | 없음 |
| B5 | 본문 write 실패 | 임시 파일 제거 | 그 오류 | 없음 |
| B6 | `Sync` 실패 | 임시 파일 제거 | 그 오류 | 없음 |
| B7 | `Close` 실패 | 임시 파일 제거 | 그 오류 | 없음 |
| B8 | 닫힌 임시 파일이 만든 그 파일이 아니다 | 임시 파일 제거 | `staged descriptor changed before publication` | 없음 |
| B9 | `Rename` 실패 | 임시 파일 제거 | 그 오류 | 없음 |
| B10 | 발행된 파일이 만든 그 파일이 아니다 | — | `published descriptor is not the staged file` | 없음 |
| B11 | 디렉터리 열기 실패 | — | 그 오류 | 없음 |
| B12 | 디렉터리 `Sync` 실패 | — | 그 오류 | 없음 |

## 분기가 아니라 **성질**을 재는 핀

열두 분기 중 열하나는 파일시스템 오류 주입이 있어야 닿는다. 이 함수가 만드는
**성질**은 그것과 무관하게 잰다.

- `TestDescriptorPublicationIsAtomic` — 재발행 300회와 동시에 읽으면서, 읽는 쪽이
  반쯤 쓰인 descriptor를 **한 번도** 보지 않는다. rename이 발밑에서 파일을 다른 완성
  파일로 바꾼 경우(`errDescriptorChanged`)는 손상이 아니라 원자성의 증거이므로 구분해
  센다. 뮤테이션 M9(제자리 `O_EXCL`)와 M9b(제자리 `O_TRUNC`)가 이 하나로 죽는다.
- `TestStartPublishesBothArtifactsByRename` — 기동이 끝난 디렉터리에 임시 이름이
  남아 있지 않다.
- `TestStartRecoversFromUnpublishedStagingLeftover` — 임시 이름으로 죽은 잔재는 다음
  부팅이 자기 잔재로 치운다.

## Calls and live bindings

| Callee | Why called | Error contract | Evidence |
|---|---|---|---|
| `os.CreateTemp(dir, ".staging-endpoint-")` | 임시 이름 확보 | 실패는 그대로 | AST calls, B2 |
| `staged.Chmod(0o600)` | umask와 무관하게 0600 | 실패는 그대로 | AST calls, B3 |
| `os.SameFile` ×2 | 만든 파일 == 발행한 파일 | 다르면 거부 | AST calls, B8·B10 |
| `os.Rename` | **발행** | 같은 디렉터리 안 원자적 | AST calls, B9 |
| `Dir.Sync` | rename의 내구성 | 실패는 그대로 | AST calls, B12 |

## State mutations and fallbacks

- 만드는 것: `.staging-endpoint-*` 하나. rename으로 사라지거나 `defer`가 지운다.
  **어느 실패 경로에서도 임시 잔재를 남기지 않는다** — 남더라도 회수가 자기 잔재로
  치운다(이중 방어).
- 최종 이름을 부분 상태로 만드는 경로가 **없다**. 그것이 이 함수의 계약이다.

## Safety conclusion

- Safe edit boundary: 임시 이름 접두(`stagingPrefix`)는 회수와 공유하는 계약이다.
  바꾸면 회수의 분류(B7)도 함께 바꿔야 하고, 안 하면 발행 잔재가 낯선 엔트리로
  분류돼 **영구 거부**가 돌아온다.
- High-risk impact: yes — 기동 7단계의 마지막 발행. 실패는 강등이지만 부분 발행은
  다음 부팅의 영구 거부였다.
