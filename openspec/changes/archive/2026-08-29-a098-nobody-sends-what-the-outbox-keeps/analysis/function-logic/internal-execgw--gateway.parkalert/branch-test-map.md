# Branch Test Map: `Gateway.parkAlert`

Source: `internal/execgw/replay.go` (534-559). AST 기준 branches 2 / returns 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:535` `g.entry != nil` → `:536` 래치. **두 갈래를 한 줄에 적는다** | 참: `TestReplay422KeyConflictParksAndNeverFails` (`internal/execgw/replay_test.go:592`)가 래치를 단언한다. **거짓(게이트 nil): 없음** — fixture가 `Entry: gate`를 항상 채운다 | no | **부분** |
| B2 | `:548` `json.Marshal` 오류 → payload 버림 | **없음** — 아래 참조 | no | **no** |

## B2는 도달 불가에 가깝다 — 그래서 RED를 만들지 않는다

`:540-547`이 마샬하는 것은 `map[string]any`이고 값이 전부 `string`이다
(`rec.ID`·`rec.IntentID`·`intent.Symbol`·`intent.Market`·`intent.AccountRef`·상수 둘).
`encoding/json`이 이 조합에서 오류를 내는 경로가 없다.

**`not-applicable` 사유를 남긴다**: 이 분기는 방어적 코드이고, 그것을 타게 하려면
프로덕션 타입을 바꿔야 한다. a098의 범위 밖이다. **침묵한 생략이 아니다.**

## 내구 절반은 따로 단언된다

`TestReplayKeyConflictEnqueuesTheCriticalAlert` (`internal/execgw/replay_test.go:628`)가 outbox 행을
단언한다 — `PendingAlerts`가 1행, 타입이 `EventOrderUnresolved`, 등급이 critical,
payload가 비어 있지 않음.

**그 테스트가 단언하지 않는 것이 a098이다**: 그 행이 **언젠가 나간다**는 것.
행이 PENDING이라는 사실만 확인하고 거기서 끝난다.

## 산출물 근거

- 분기 열거: `ast.json` (branches 2, returns 0)
- 래치 단언 자리: `internal/execgw/replay_test.go:620`
- fixture가 게이트를 채운다: `internal/execgw/replay_test.go:84`
