# D11 writer 공존성 측정 (2026-10-29)

## 결론

**갈래 A — 단일 writer 보장이 코드로 존재함.** R3 `FAILED_CONFIRMED` latch를 쓰는 자동 취소와 D11이 지키는 진입 PLACE `Prepare`는 모두 `tossctl engine run`의 engine Gateway에서만 발생함. 이 명령은 journal을 열거나 engine을 조립하기 전에 journal 디렉터리의 kernel `flock`을 잡으므로, 같은 journal에 대한 두 engine writer는 공존할 수 없음.

## 경로 열거

| 경로 | 파일:줄 | 같은 journal에 접근 | R3 latch 작성 | D11 진입 PLACE | 공존 판정 |
|---|---|---|---|---|---|
| `tossctl engine run` | `cmd/tossctl/engine.go:204-210`, `internal/app/engine/gateway.go:311-319` | 예 | 예 — `EngineCancelRefusal()` | 예 — Gateway `submit`→`Prepare` | 같은 journal의 두 instance는 flock으로 거절됨 |
| `tossctl flatten-all` | `cmd/tossctl/flatten.go:226-244` | 예 | 아니오 — `FlattenCancelRefusal()`은 AP를 CONFIRMED 또는 첫 AMBIGUOUS로만 종결 | 아니오 — flatten은 reduce-only cancel/sell saga | engine과 별도 프로세스로 공존할 수 있으나 측정 대상 writer (가)/(나)가 아님 |
| 수동 `tossctl order place` | `cmd/tossctl/order.go:200-240` | 아니오 | 아니오 | 아니오 — `tradingService.Place` 직접 호출, `execgw`/`Journal.Prepare`를 통과하지 않음 | 인간 승인 live 명령의 별도 계약이며 D11 writer가 아님 |
| 운영 콘솔 | `cmd/tossctl/console.go:503`, `internal/console/engineproc.go:67` | 직접 접근 없음 | 아니오 | 아니오 | 엔진 기동은 `engine run`으로 위 lock 계약을 따름 |

## 단일 writer 근거와 시험

- `cmd/tossctl/engine.go:204-210`은 `enginelock.Acquire(dir)`를 assembly보다 먼저 실행하고 `defer lock.Release()`로 런타임 전체에 걸쳐 보유함.
- `internal/enginelock/enginelock.go:1-50`은 이 lock을 journal directory의 kernel-enforced exclusive `flock`으로 정의함. marker는 exclusion이 아님.
- `internal/enginelock/enginelock_test.go:31-53`의 `TestASecondAcquireIsRefused`가 같은 directory의 두 번째 acquire를 `ErrAlreadyRunning`으로 고정함.
- `cmd/tossctl/engine_test.go:96-124`의 `TestTheJournalDirectoryIsLockedBeforeAnythingIsAssembled`가 lock 보유 시 두 번째 `engine run`이 assembly/journal open 전에 거절됨을 고정함.

따라서 latch 확인과 `Prepare` 사이에 다른 자동 engine process가 R3 latch를 정착할 수 없다는 보장이 성립함. 이 보장이 깨지는 변경은 위 두 lock 시험과 이 측정 기록을 갱신하지 않고 허용할 수 없음. D11의 추가 마지막 전송 경계 재조회는 필요하지 않으며, 별도 고위험 `submit` 편집을 만들지 않음.
