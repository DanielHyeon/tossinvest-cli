# a067 · Review

- 날짜: 2026-08-03
- 시점: proposal-freeze (첫 구현 task 착수 전)
- 대상: `proposal.md`, `design.md`, `specs/operator-console/spec.md`
- 보이스: Manager 셀프리뷰 + **적대적 Eng**. WORKFLOW의 등급제로는 UI change라
  경량 리뷰로 충분하지만, 이 change가 그리는 것이 보호선이라 적대적 관점을 붙였다.
- 남은 절차: 구현 후 **별도 컨텍스트의 독립 검증**(WORKFLOW §9)은 아직 수행되지
  않았다. 이 문서는 그것을 대체하지 않는다.

## 적대적 Eng — 제기된 반론과 처리

### A1. "엔진이 살아 있다"는 "이 포지션이 관측되고 있다"가 아니다 — **수용, 부분 반영**

프로세스 마커는 프로세스에 대한 사실이다. 엔진이 돌면서 특정 심볼만 관측에서 빠질
수 있는 경로가 실제로 있다.

| 경로 | a061이 닫는가 |
|---|---|
| 활성 quarantine (`ErrExitSnapshotQuarantined`) | **닫는다** — D3 |
| lifecycle generation 불일치 / RELEASED | 이미 닫혀 있다 (B3/B4) |
| snapshot 무결성 실패 | 이미 닫혀 있다 |
| **특정 심볼의 시세 두절** | **닫지 못한다** |

마지막 줄은 잔여 위험이다. 엔진에는 outage ladder가 있어 60초를 넘으면 critical
알림과 ENTRY_BLOCKED 강화가 발동하지만, 그 상태는 계좌 단위 `operating_modes`에
남고(이미 다른 이유로 2026-07-31부터 ENTRY_BLOCKED다) 심볼 단위로 읽을 수 있는
신호가 아니다. **그 구멍을 정직하게 메우는 유일한 방법은 엔진이 변화 없는 관측에도
하트비트를 남기는 것**이고, 그것이 후속 change다. a061의 판정 구조는 그때 조건
하나가 **추가**되는 형태라 버려지지 않는다.

수용 근거: a067 이전에는 이 구멍이 **더 나쁜 방식으로** 가려져 있었다 — 시세가
정상이든 두절이든 30초 뒤 똑같이 `—`였으므로, 화면은 두절을 알려 준 적이 없고
정상도 알려 준 적이 없다.

### A2. "보수성을 낮추는 변경 아닌가" — **반박**

§0.9는 손절·익절·사이징 **로직**의 단방향 안전을 요구한다. a061은 로직을 만지지
않는다. 표시 계약만 놓고 보면 순증가다.

| | 이전 | 이후 |
|---|---|---|
| 나이 기반 억제 | 항상 (틀린 근거) | 미배선 콘솔만 |
| 엔진 정지 시 억제 | **없음** | 있음 |
| quarantine 시 억제 | **없음** | 있음 |
| generation/무결성 억제 | 있음 | 그대로 |

그리고 "항상 `—`인 화면"은 그 자체로 안전 실패다. 운영자는 언제나 비어 있는 열을
읽지 않게 되고, 그러면 진짜로 비어야 할 때 그 사실이 전달되지 않는다.

### A3. 마커의 5분 창 — **수용, 명시**

`enginelock.StaleAfter`는 5분이고 마커는 1분마다 갱신된다. crash한 엔진은 최대 5분
동안 "실행 중"으로 읽힌다. 그 5분 동안 a061은 보호선을 표시한다.

같은 trade를 `internal/runlock`과 상태 스트립이 이미 받아들이고 있고
(`engineproc.go` 패키지 주석에 명시돼 있다), 화면 상단이 그 판정을 그대로 보여 준다.
이전 30초 기준과 비교하면 노출 창이 늘지만, 이전 기준은 **정상 엔진에서도 항상
만료**됐으므로 실제로 비교 가능한 값이 아니다.

거부한 대안: 마커 창을 이 화면에서만 좁히기. 화면마다 다른 "실행 중" 정의를 갖게
되고, 상태 스트립은 실행 중이라는데 라인은 정지라고 말하는 화면이 된다.

### A4. quarantine을 안 읽고 D1만 하면? — **차단 사유, 그래서 함께 간다**

466100은 09:03:40Z부터 활성 quarantine이다. D1만 적용하면 그 행은 `—`에서
**실행 가능한 보호선**으로 바뀐다. 엔진은 그 포지션을 판정하지 않고 있는데 화면은
보호되고 있다고 말하게 된다. 이것 하나로 change 전체가 거부 사유가 되므로,
quarantine read는 스코프 확장이 아니라 전제다. spec delta에도 SHALL로 넣었다.

### A5. quarantine read 실패는 어느 방향으로 넘어지나 — **fail-closed**

`LivePositionExits` 안에서 읽고 오류를 그대로 전파한다. 호출부는 이미 그것을
`journalFailed`로 강등하고 화면은 브로커 절반만 그린다 — 라인은 전부 `—`다.
"읽지 못했으므로 격리되지 않았다고 본다"는 fail-open은 채택하지 않았다.

스키마 부재 걱정은 없다. `exit_snapshot_quarantines`는 schemaV10에 있고
`checkSchema`가 버전을 강제하므로, journal이 열렸다면 테이블이 있다.

### A6. peek로 가져온 이름이 다른 계좌의 것일 수 있다 — **반박**

holdings 캐시는 이 콘솔의 브로커 계좌 하나다. journal은 여러 계좌를 담을 수 있다.
그러나 키가 `(market, symbol)`이고 **이름은 종목의 속성이지 계좌의 속성이 아니다**.
같은 시장의 같은 심볼은 같은 종목이므로 계좌가 달라도 이름은 같다. 시장을 키에서
빼면 그때는 틀리므로, `joinPositions`가 쓰는 `symbolKey`를 그대로 쓴다(D5).

### A7. `positionpolicy.State`에 `Name`을 넣지 않은 것이 맞나 — **유지**

journal이 그 필드를 저장하지 않으므로, 넣으면 경로에 따라 채워지고 비는 도메인
필드가 된다. 엔진이 그 값을 신뢰할 근거가 생기는 것도 원치 않는다. 뷰 모델에 둔다.

### A8. 새 capability가 늘지 않았나 — **확인**

`HoldingsReader`는 메서드 1개 그대로다. 새 journal read는 `*journal.ReadOnly`에
들어가고 그것은 주입 seam이 아니다.
`TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads`가
그대로 통과해야 하며 6.4에서 확인한다.

### A9. 브로커 호출·rate budget — **영향 없음**

`/position-management`는 `peek`만 쓴다(호출 0회). 라인 판정은 파일 stat과 로컬
SQLite read다. §0.4 예산 변동 없음.

## Manager 셀프리뷰 — 스펙 품질

- ADDED만 사용했다. a055 `issues.md` I1의 미아카이브 MODIFY 부채를 늘리지 않는다.
- 기존 "stale이면 `—`" 문장과 충돌하지 않는다. a061이 정의하는 것은 **무엇이 stale을
  만드는가**이고, 기존 문장은 stale일 때 무엇을 하는가다.
- 기존 "미배선을 정지로 표시해서는 안 된다"(spec:476)를 D2 세 번째 갈래가 지킨다.
- Scenario는 판정 갈래 하나당 최소 하나씩 있고, 회귀 갈래(미배선·무결성·이전 세대
  격리)도 Scenario로 고정했다.

## 결론

**진행 승인.** 단, 다음 두 가지를 완료 조건에 남긴다.

1. A1의 잔여 위험(심볼 단위 시세 두절)은 a061이 닫지 못한다. `issues.md` I1의
   후속 change가 닫는다는 것을 완료 보고에 명시한다.
2. 466100의 quarantine 해제 여부는 §0.7 사람 판정이다. a061은 표시만 한다
   (`issues.md` I2).

---

## 구현 후 검토 (2026-08-03, 별도 패스)

**한계 명시**: 이것은 구현을 만든 컨텍스트가 diff를 다시 읽은 것이지 WORKFLOW §9가
요구하는 **별도 세션의 독립 검증이 아니다**. 이 change는 그 검증을 아직 받지 않았다.

### 변이 검증 (tasks 6.1~6.3)

각 기제를 되돌렸을 때 대응 테스트가 RED가 되는지 확인하고 원본을 바이트 단위로 복구했다.

| 변이 | 대상 테스트 | 결과 |
|---|---|---|
| `exitFreshness`를 `WithFreshness(asOf, holdingsTTL)`로 되돌림 | `TestTheProtectionLineSurvivesAJudgementThatChangedNothing` | FAIL ✓ |
| `accountActiveQuarantines`가 빈 map을 조기 반환 | `TestAQuarantinedPositionIsNotDrawnAsProtected` | FAIL ✓ |
| `q.position_generation = p.instance_seq` 제거 | `TestAReleasedGenerationsQuarantineDoesNotCloseTheCurrentOne` | FAIL ✓ |

세 번째 변이 중 `git checkout --`로 복구를 시도해 **작업 중이던 `account_views.go`의
a067 변경이 지워졌다.** 즉시 복구하고 변이 검증을 백업 파일 방식으로 다시 수행했으며,
복구 후 파일이 백업과 바이트 동일함을 `diff`로 확인했다. 커밋 전 상태이므로 손실은 없다.

### diff 재검토에서 나온 것

1. **quarantine은 `HasExit == true`인 행에만 닿는다.** `!row.HasExit`이면 B3에서
   `continue`하므로 quarantine 사유가 표시되지 않는다. 현재 코드에서 이 조합은
   발생하지 않는다 — quarantine은 `exitloop`가 exit state를 읽다가 만들고, 그 state
   행은 지워지지 않으므로 `accountExitStates`가 항상 되돌려준다. 다만 이것은 **불변식이
   아니라 관찰**이므로 `issues.md` I5에 기록했다.
2. **`/dashboard`도 같은 seam을 쓴다.** `overviewLedger`가 `LivePositionExits`를
   호출하므로 quarantine이 두 화면에 동일하게 도달한다. 처음 테스트는 `/positions`만
   확인했으므로 `/dashboard`까지 확장했다 (spec의 "두 화면의 같은 답" scenario).
3. **`held := q` 복사**는 map 값 복사 뒤 주소를 취하므로 두 행이 한 변수를 공유할 수
   없다. 확인했다.
4. **capability 증가 없음**: `HoldingsReader`는 메서드 1개 그대로이고
   `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads`가
   green이다 (652 console 테스트 전부 green).

### 실행 결과

- `go test ./...` — 78 패키지 6036건 green (upstream 상속 테스트 포함).
- `go vet ./...` — 무결함. `go fmt` — 변경 없음.
- `make sdd-check` — CodeGraph hard-evidence가 worktree와 일치. CodeGraphContext와
  GBrain은 advisory warning (아래 참조).
- `python3 tools/logic-map/check_analysis.py` — 8개 함수 evidence 완성.
- `python3 tools/pm/generate_master_tracker.py --check` — current.

### 남은 것

- **컨테이너 실측(6.7)은 사람 승인 전이다.** 배포와 live 확인은 §0.7이며 에이전트가
  단독으로 하지 않는다.
- CodeGraphContext advisory 인덱스는 `make sdd-sync`에서 300초 timeout으로 갱신되지
  않았다. hard gate가 아니며 CodeGraph는 갱신됐다.

## 병합 후 재기준화 (2026-08-04)

이 change는 `main`과 동기화되지 않은 브랜치 위에서 작성됐다. `main`은 그동안
journal schema를 15에서 19로 올렸고, 그 브랜치로 빌드한 이미지는 콘솔만 뜨고
엔진이 journal 열기를 거부했다. 배포하려면 병합이 선행 조건이었다.

병합은 Function Logic Map 검사의 의미를 바꾼다. `check_analysis.py`는
`base-commit.txt`부터 작업 트리까지를 diff하므로, 병합 이후에는 `main`이 고친
함수까지 이 change가 고친 것으로 집계된다.

| 항목 | 값 |
|---|---|
| 원래 base | `321cf78e8b63ef36c5ff94df97547c1006d8aa06` |
| 재기준화된 base | `6b47be2ff8ebd21a59dec682183e015cec8584da` (병합 커밋) |

**재기준화 전에 원래 base로 검사를 다시 돌렸다.** `321cf78e8b63` 커밋을 detached
worktree에 체크아웃하고 `check_analysis.py --change`를 실행해
`evidence complete or diff-proven exempt`를 확인했다. 즉 이 change가 실제로 고친
함수의 증거는 완전했고, 병합이 그것을 지운 것이 아니다.

`main`이 고친 함수들의 logic map은 저장소 안에 그대로 있다 —
`openspec/changes/archive/2026-08-03-a061-show-history-instrument-names/`와
`.../2026-08-03-a062-reconcile-owned-orders/`의 analysis가 그것이다.

`revision: current`로 고정된 AST 증거는 병합된 파일 내용으로 다시 추출했다.
분기 ID 집합이 달라진 타깃은 자동으로 덮어쓰지 않고 따로 처리했다.

**게이트 결과의 의미는 좁아졌다.** 재기준화 이후 이 change의 logic-map 단계는
"병합 시점 이후로 이 change가 더 고친 함수가 없다"를 확인하는 것이고, "이 change가
고친 함수에 증거가 있다"는 위의 원래 base 재검사가 확인한다. 한 명령으로 둘 다
재도출되지는 않는다.

`internal-console--console.handlepositionmanagement` 타깃은 제거했다. a069가 같은
함수를 나중에 고쳐 분기가 16개에서 20개로 늘었고, 20개 전부를 현재 트리에 묶어
RED/GREEN까지 관측한 증거가
`openspec/changes/a069-operator-can-lift-a-quarantine/analysis/function-logic/internal-console--console.handlepositionmanagement/`에
있다. a069의 표를 여기로 복사하면 a069가 관측한 RED를 a067의 관측인 것처럼 만드는
것이므로 하지 않았다. a067 시점의 16분기 증거는
`321cf78e8b63` 커밋에 그대로 남아 있다.
