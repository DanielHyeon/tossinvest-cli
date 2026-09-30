판정: **BLOCK**

| # | 등급(P0/P1/P2) | 파일:줄 | 무엇이 틀렸나 | 근거(코드 인용 또는 실행 결과) | 제안 |
|---|---|---|---|---|---|
| 1 | **P0** | [notifier.go:594](/tmp/claude-1000/a092-r25-tree/internal/obs/notifier.go:594) | 임차 반납의 **행 없음**을 정상 선점으로 처리해 차단·승격을 생략한다. frozen 델타의 “행 없음·모르는 결과·원장 오류는 선점이 아니다”를 위반한다. | `ReleaseAlertClaim`은 행 소실 시 **`SettleNotFound, nil`**을 반환한다([alert_claim.go:372](/tmp/claude-1000/a092-r25-tree/internal/journal/alert_claim.go:372)). 따라서 `deliver`의 `default` → `return false, true, latchVerdict{}`에 도달한다. 이어 `claimAndDeliver:348`이 `owed=false`로 바꾸므로 `notifyCritical:238`의 판정 적용도 생략된다. 로그만 남고 전달 실패 차단은 없다. | 정상 선점은 `SettleAlreadySettled`·`SettleLeaseLost`만 열거한다. `SettleNotFound`·미지 결과는 별도 실패 판정으로 처리하고 해제 세대 규칙을 적용한다. 마지막 실패 기록과 반납 사이에 행을 소실시키는 결정적 시험을 추가한다. |
| 2 | **P2** | [notifier.go:112](/tmp/claude-1000/a092-r25-tree/internal/obs/notifier.go:112), [notifier.go:264](/tmp/claude-1000/a092-r25-tree/internal/obs/notifier.go:264) | 잠금 설명이 변경된 구현과 반대다. | `mu serialises the whole claim-and-send`, `both under the delivery mutex`라고 적지만 실제로는 `:344`에서 해제한 뒤 `:347`에서 전송한다. | 잠금은 기록·claim과 승인 셈~해제를 직렬화하고, 전송 중 배제는 임차가 담당한다고 정정한다. |

발견 1은 **코드로 확인한 장애 분기**이며 실행 재현 결과는 아니다. 편집 전 FLM에도 반납의 포괄적 선점 분기가 있으므로, 이번 커밋이 새로 만든 회귀라고 단정하지 않는다. 다만 이번 검토 대상인 `deliver`가 [frozen 델타:62](/tmp/claude-1000/a092-r25-tree/openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md:62)를 만족하지 않는다는 점은 명확하다.

잠금 보유자와 critical 기록 호출자 전수 결과는 다음과 같다.

| 전수 대상 | 확인 결과 |
|---|---|
| `n.mu` 보유자 | `recordCritical`, `claimAndDeliver`, `Acknowledge`, `Flush` 네 함수. 앞의 셋은 원장·메모리 게이트·로그 작업을 수행하며 원격 전송을 덮지 않는다. |
| `Flush` 예외 | `notifier.go:814`부터 잠금을 보유하고 `:859`에서 전송한다. **생산 호출자 0**이므로 현재 생산 배선의 위반으로 세지 않았다. 호출자 0 구조 핀은 단위 ⑤의 잔여다. |
| `RecordAlert` 호출자 | `record_only.go:118` 하나. `n.mu` 아래에서 호출한다. |
| `ClaimAlertForDelivery` 호출자 | `notifier.go:294` 하나. `n.mu` 아래에서 호출한다. |
| `EnqueueAlert` 호출자 | `replay.go:553`의 `parkAlert` 하나. `:538`에서 `ReasonUnresolvedInDoubt`를 먼저 세우며 삽입 전 해제가 없다. 엔진은 Gateway와 Notifier에 동일한 `entry`를 넘긴다(`gateway.go:302,323`). |
| 「모든 발송자」 규칙 | 기존 세 래치 자리의 해제 세대 읽기, 조건부 차단, 승격 유지, 승격 실패 시 무조건 차단, 승격 미포함 처리, 승인 시각 비사용은 확인했다. **반납의 행 없음 분류는 발견 1 때문에 불충족**이다. |

저자 주장별 판정:

1. **참 — k3 합성은 현재 코드에서 성립한다.** `ConfirmedFloor`의 Holdings·SellableQuantity 조회는 모두 `f.retrier.Query`를 거친다(`exitwiring.go:207,231`). floor 복사본에는 exit Retrier가 들어가고, 인증 실패는 그 Retrier의 `Announcer`로 간다(`retry.go:413`). 우회 조회·별도 통지자 경로는 발견하지 못했다.

2. **참 — 유효 임차가 전송 시간을 덮는 정본의 유계 실행 전제 아래다.** `n.mu`는 전송 전에 풀린다. L19 원장은 임차 무시 변이가 동시 관측 등 시험 8개를 실패시켰다고 기록한다. 해당 변이는 이번 세션에서 재실행하지 못했다.

3. **참 — 정상 승인과의 겹침에 한정한다.** 승인된 행은 `AlreadySettled`로 종료하고, 판정 뒤 해제는 세대로 구별한다. 승격 실패의 무조건 차단도 있다. 다만 이 판정이 **모든 정산 장애에서 안전하다**는 뜻은 아니다. 행 소실 반납은 발견 1의 별도 결함이다.

4. **참 — 생산 전이 호출 기준이다.** `TransitionOperatingMode`는 무변화·AUTO 완화 시 통지 전에 반환한다(`operating_mode.go:410,417`). 커밋된 전이에만 통지하며 키에 `rec.ID`가 들어간다. 여기서 통지 수는 논리적 통지 사건 수이며 전송 재시도 횟수가 아니다.

5. **참 — 행 구성과 직렬화 기준이다.** 키·제목·본문·payload 필드는 유지되며, `encodeFields`도 `json.Marshal(map)`을 사용한다. `remindAfter=0`으로 이전 enqueue의 재알림 창 동작을 유지한다. 바이트 대조 시험은 존재하지만 이번 실행은 막혔다.

6. **참 — 열거된 critical 네 도달 경로 기준이다.** Alerts·관측 두절 Announcer·가격 조회 Retrier·floor Retrier가 기록 전용으로 연결된다. 생산 `engineRuntime`은 동기 Announcer를 덮어씌우지 않는다. **일반 등급은 `record_only.go:50`에서 여전히 동기 전송**하므로 전체 a092 비대기 SHALL은 아직 미완성이다. 요청대로 단위 ⑤의 알려진 잔여로 취급했다.

시험 명령은 다음과 같이 시도했으나 **시험 실행 전에 실패**했다.

```text
GOCACHE=/tmp/claude-1000/a092-gocache TMPDIR=/tmp/claude-1000 \
GOPROXY=off GOSUMDB=off go test \
./internal/journal ./internal/obs ./internal/app/engine ./cmd/tossctl \
-run 'TestA092|TestAcknowledgeCannotClearTheGateMidSend|TestAHeldRowIsNotWhispered' -count=1

go: creating work dir: mkdir /tmp/claude-1000/go-build3742721552: read-only file system
```

파일 변경·git 상태 변경·운영 명령은 실행하지 않았다. 금지된 리뷰 원문과 운영 원장·자격 증명 파일도 열지 않았다.

Recommendation: 반납 결과 분류를 수정하고 행 소실·승인 경합 시험 후 재리뷰 because 현재 구현은 행 소실을 선점으로 오인해 frozen 델타가 요구하는 실패 차단을 건너뛴다.
