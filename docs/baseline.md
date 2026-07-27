# TossOS 베이스라인

TossOS는 upstream `tossinvest-cli`를 포크한 저장소입니다. 이 문서는 포크 시점의
**고정 베이스라인**(커밋·툴체인·빌드/테스트 결과·커버리지·알려진 갭)을 기록합니다.
이후 모든 변경은 이 기준선과의 차이로 평가합니다.

- 기록 시점: Phase 0 (`add-tossos-foundation`) — tasks 2.1~2.5
- 원칙: 이 문서의 수치는 **재현된 실측값**입니다. 추정치를 섞지 않습니다.

---

## 1. 고정 커밋

| 항목 | 값 |
| --- | --- |
| 고정 커밋(HEAD) | `57348a7ffb234c98d6d0c9ee1d6ae3c9a5af2867` |
| 기준 위치 | upstream `v0.30.0` 이후 커밋 |
| 히스토리 | 전체 히스토리 클론(non-shallow) — 483 커밋, 75 태그 |
| 작업 브랜치 | `feat/p0-foundation` (`main`은 upstream 추적용으로 보존) |

### 리모트 구성

| 리모트 | URL | 비고 |
| --- | --- | --- |
| `upstream` | `https://github.com/JungHoonGhae/tossinvest-cli` | fetch 전용으로 사용 |
| `origin` | *(미설정)* | 의도적으로 비워 둠 — 실수 push 방지 |

> `origin`이 없으므로 `git push`는 리모트를 명시하지 않으면 실패합니다. 이는 버그가
> 아니라 안전장치입니다.

---

## 2. 도구 버전

| 항목 | 값 |
| --- | --- |
| Go module path | `github.com/JungHoonGhae/tossinvest-cli` |
| `go.mod` go 지시자 | `go 1.25.0` |
| 로컬 툴체인 | `go1.26.5` |

> 모듈이 요구하는 최소 버전(1.25.0)보다 로컬 툴체인(1.26.5)이 높습니다. 베이스라인
> 측정은 1.26.5에서 수행되었습니다.

### 버전 문자열의 실제 출처

`VERSION` 파일(`0.8.0`)은 **dead 파일**입니다. 릴리스 버전은 git tag에서 뽑아
`Makefile`의 `LDFLAGS`(`-X .../internal/version.Version=...`)로 빌드 시 주입됩니다.
따라서 `VERSION` 파일의 내용은 실행 바이너리의 `--version` 출력과 무관합니다.
버전 관련 작업을 할 때 `VERSION` 파일을 고쳐서는 의미가 없습니다.

---

## 3. 빌드·검사 결과

| 명령 | 결과 |
| --- | --- |
| `go build ./...` | **PASS** |
| `go vet ./...` | **PASS** |

---

## 4. 테스트 결과

`go test ./... -count=1` (캐시 무효화, 1회 실행)

| 항목 | 값 |
| --- | --- |
| 통과 | **650** (top-level 576 + subtest 74) |
| 실패 | 0 |
| 스킵 | 0 |
| 패키지 수 | 25 |
| `ok` 패키지 | 21 |
| 테스트 없는 패키지 | 4 — `internal/domain`, `internal/export`, `internal/replay`, `internal/version` |

**베이스라인은 green입니다.** 이후 어떤 변경도 이 650 통과 / 0 실패를 깨뜨리면
회귀로 간주합니다.

---

## 5. 패키지별 커버리지

`go test ./... -count=1 -cover` 기준 statement coverage(%). 테스트 파일이 없는 4개
패키지는 표에서 제외했습니다.

| 패키지 | 커버리지(%) |
| --- | --- |
| `internal/onboarding` | 100.0 |
| `internal/i18n` | 93.3 |
| `internal/mcp` | 84.7 |
| `internal/official` | 84.7 |
| `internal/updatecheck` | 83.3 |
| `internal/orderintent` | 75.2 |
| `internal/config` | 75.0 |
| `internal/orderlineage` | 74.5 |
| `internal/trading` | 74.5 |
| `internal/auth` | 70.5 |
| `internal/client` | 70.3 |
| `internal/push` | 68.3 |
| `internal/hybrid` | 49.6 |
| `internal/session` | 48.1 |
| `internal/output` | 47.2 |
| `internal/selfupdate` | 38.5 |
| `cmd/tossctl` | 30.6 |
| `internal/tui` | 29.5 |
| `internal/ops` | 20.5 |
| `internal/monitor` | 17.2 |
| `internal/doctor` | 7.7 |

### 읽는 법

- **높은 쪽**(`onboarding`, `i18n`, `mcp`, `official`, `updatecheck`)은 순수 로직 비중이
  높아 단위 테스트가 잘 붙은 영역입니다.
- **거래 핵심 경로**(`trading` 74.5, `orderintent` 75.2, `orderlineage` 74.5,
  `client` 70.3)는 70%대입니다. 실계좌 주문을 다루는 코드이므로 이 수치는 유지
  하한선으로 취급하고, 이 경로를 건드릴 때는 커버리지를 떨어뜨리지 않습니다.
- **낮은 쪽**(`doctor` 7.7, `monitor` 17.2, `ops` 20.5, `tui` 29.5, `cmd/tossctl` 30.6)은
  I/O·터미널·외부 프로세스 의존이 커서 테스트가 얇습니다. TossOS에서 이 영역을
  다시 쓸 때는 순수 로직을 분리해 커버리지를 올릴 여지가 큽니다.

전체 합산(total) 커버리지는 **51.9%**입니다 (`make cover` → `go tool cover -func`).

---

## 6. upstream 알려진 갭

`docs/architecture.md`의 "Current Gaps" 절과 `TODOS.md`에서 추출한, upstream이
스스로 미완으로 표시한 항목입니다. TossOS가 이 경로를 건드릴 때는 아래가 여전히
미검증 상태라는 전제에서 출발해야 합니다.

### docs/architecture.md — Current Gaps

- **갭 1 — `amend`의 추가 live 검증**: 정정 주문의 실계좌 검증이 부족합니다.
- **갭 2 — `place`와 `amend`의 상태 판별 추가 검증**: 주문 제출 후 상태 판정 로직의
  검증이 부족합니다.
- **갭 3 — 비소수점 시장가 주문(US/KR)**: 소수점 수량을 쓰지 않는 시장가 주문 경로가
  미구현/미검증입니다.
- **갭 4 — interactive auth challenge가 필요한 mutation 분기 일반화**: 추가 인증이
  요구되는 mutation 처리가 케이스별로 흩어져 있어 일반화가 필요합니다.

### TODOS.md

- **갭 5 — Sell 주문 live 검증 확대**: 분할 매도(일부 수량), 전량 매도, 보유량 초과
  요청 시 API 동작이 미검증입니다. FX consent 방향(USD→KRW), auth-required 분기,
  holdings rejection도 함께 검증 대상입니다. (근거: 2026-03-21 Codex 리뷰 —
  "partial sell is the normal case")
- **갭 6 — KR stock cancel/amend live 검증**: 국내주식 cancel/amend가
  `InferMarketFromStockCode`로 시장을 **추론**합니다(`pendingOrderDetails`에 market
  필드가 없기 때문). 실제 KR 주문에서 올바른 시장으로 reconcile되는지 미검증입니다.
  (근거: 2026-03-21 KR trading eng review)

> 갭 5·6은 실제 보유 주식 / 실제 KR 미체결 주문이 있어야 검증 가능합니다. 즉 코드
> 리뷰만으로는 닫을 수 없는 갭입니다.

### 문서 신뢰도 주의

`docs/architecture.md`의 **Package Map은 stale**합니다. 실제 24개 패키지(`cmd/tossctl`
포함 25개) 중 8개만 기재되어 있습니다. 아키텍처를 파악할 때 이 절을 완전한 목록으로
신뢰하면 안 됩니다. 실제 목록은 `go list ./...`로 확인하십시오.

---

## 7. 환경 주의사항

### NTFS 마운트와 `core.filemode`

저장소가 **NTFS(fuseblk) 마운트** 위에 있습니다. NTFS는 POSIX 실행 비트를 보존하지
않으므로, git이 모든 파일의 모드 변경을 오탐합니다.

이를 막기 위해 로컬에 다음이 설정되어 있습니다 — **필수 설정이며 해제하면 안 됩니다.**

```ini
[core]
    filemode = false
```

결과적으로 다음 사항을 유의하십시오.

- **실행 비트가 보존되지 않습니다.** `chmod +x`는 로컬에서 의미가 없고, 커밋에도
  반영되지 않습니다.
- 새로 추가하는 스크립트가 실행 가능해야 한다면, 실행 비트에 의존하는 대신
  `bash script.sh` 형태로 인터프리터를 명시하거나, 커밋 후
  `git update-index --chmod=+x <path>`로 인덱스 모드를 직접 지정해야 합니다.
- `git status`에 파일 모드만 바뀐 변경이 대량으로 뜨면 `core.filemode` 설정이
  풀렸는지 먼저 확인하십시오.

### 재현 명령

베이스라인을 다시 확인할 때:

```bash
git rev-parse HEAD          # 57348a7ffb234c98d6d0c9ee1d6ae3c9a5af2867 인지 확인
go version                  # 툴체인 확인
make build                  # 빌드
make vet                    # go vet
make test                   # 전체 테스트 (650 통과 / 0 실패)
make cover                  # 합산 커버리지
make validate               # openspec 스펙 검증
```
