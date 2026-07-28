# Function Logic Map: `statfsProber.Probe`

- Source: `internal/candidate/fsprobe_linux.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change에서 본문이 바뀐 함수다. `FSInfo`에 D16의 `FreeBytes`/`FreeMeasured`가 붙으면서 `Bavail × Bsize` 계산과 `Bsize > 0` 가드가 들어왔다. 이 값이 발굴 정지 게이트의 입력이므로, 과대 보고는 원장 파일시스템을 채우는 방향이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `dir` string | 존재하는 디렉터리 경로 | `Store.Open`이 `filepath.Dir(path)`로, `Store.FreeSpace`가 같은 값으로 넘긴다 | `syscall.Statfs` 실패 시 `FSInfo{}` + wrap된 에러 |
| `syscall.Statfs_t.Type` | 커널 magic | 커널 | 32비트 마스킹으로 32-bit Linux에서 `0x9123683E`가 음수로 sign-extend되는 것을 막는다 |
| `st.Bavail` | 이 사용자에게 남은 블록 수 | 커널 | `Bfree`가 아니라 `Bavail`이다. 예약 블록은 이 프로세스가 쓸 수 없고, 세면 발굴이 원장의 write가 실패하는 지점까지 계속 쓴다 (D16) |
| `st.Bsize` | 양수여야 유효 | 커널 | 0 이하면 `FreeMeasured`를 **세우지 않는다**. 미측정은 0바이트가 아니고, 0바이트로 보고하면 아무도 재지 않은 값이 가장 놀라운 읽기가 된다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `syscall.Statfs(dir, &st)`가 에러 | 없음 | `FSInfo{}, fmt.Errorf("statfs %s: %w")` | 직접 단위 테스트 없음 — 프로덕션 전용 prober |
| B2 | `st.Bsize > 0` | `info.FreeBytes`, `info.FreeMeasured = true` 지역 설정 | 두 경우 모두 `info, nil` | 계약 대역 `spaceProber`의 두 상태로 `TestDiscoveryStopsWritingBeforeTheLedgerRunsOutOfSpace` / `TestSpaceItCouldNotMeasureIsNotSpaceItHas` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `syscall.Statfs` | 마운트의 magic과 잔여 공간을 한 번에 얻는다 | 동기 syscall, 재시도 없음. 실패는 즉시 반환 | ast.json calls |
| `FilesystemName` | magic을 허용목록의 이름으로 | 순수 map 조회, 미등록은 빈 문자열 | `fsguard.go:144` |
| `fmt.Errorf` | 경로를 담은 진단 | — | ast.json calls |

## State mutations and fallbacks

- 영속 상태를 만들지 않는다. 반환하는 `FSInfo`는 값이고, 호출자(`CheckFilesystem`, `Store.FreeSpace`)가 판정한다.
- fallback: `Bsize <= 0`이면 `FreeMeasured=false`로 나가고 `Store.FreeSpace`가 `ErrProbeUnsupported`로 올린다. watch 루프는 그것을 **정지**로 읽는다 — 못 잰 공간은 있는 공간이 아니다.
- 비-Linux 빌드에서는 `fsprobe_other.go`의 `unsupportedProber`가 항상 `ErrProbeUnsupported`다. 이 파일의 두 분기는 Linux에서만 존재한다.

## Safety conclusion

- Safe edit boundary: `Bavail`→`Bfree` 교체, `FreeMeasured`의 기본값을 true로 두기, `Bsize<=0`에서 0을 측정값으로 내보내기는 모두 D16의 안전 방향을 뒤집으므로 금지
- High-risk impact: no — 파일시스템 읽기 전용 probe이고 주문·손절·원장 write를 하지 않는다. 다만 이 함수의 반환값이 D16 정지 게이트의 **유일한 입력**이라 과대 보고는 발굴이 원장의 파일시스템을 채우는 경로가 된다. 그래서 편향은 `Bavail` 쪽(보수)으로 고정한다.
