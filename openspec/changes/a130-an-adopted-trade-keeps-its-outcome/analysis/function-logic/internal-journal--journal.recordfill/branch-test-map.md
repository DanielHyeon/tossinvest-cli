# Branch Test Map: `Journal.RecordFill`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:314` `if strings.TrimSpace(obs.OrderID) == "" {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B2 | `:318` `if !scope.complete() && !scope.legacyUnscoped() {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B3 | `:324` `if err != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B4 | `:328` `if math.IsNaN(filled) \ | not-applicable: 무편집 갈래이고 이 패키지 시험 전체가 도달하지 않음(측정) — a130 이전부터의 시험 공백, a130 범위 밖으로 기록 | no | n/a |
| B5 | `:340` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B6 | `:346` `switch {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40 | no | yes |
| B7 | `:347` `case err == nil:` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B8 | `:348` `case errors.Is(err, ErrFillNotFound):` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40 | no | yes |
| B9 | `:350` `default:` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B10 | `:356` `if ownerErr != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B11 | `:364` `if hadPrev && scope.complete() {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B12 | `:368` `if owners == 1 && prev.CommittedAt <= ownershipAt {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B13 | `:374` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B14 | `:380` `if hadPrev {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B15 | `:385` `if refusal := classifyFillRefusal(obs, prev, hadPrev, filled, prevFilled); refus` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B16 | `:392` `if err := markFillRefused(ctx, tx, scope, obs, refusal, now, hadPrev); err != ni` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B17 | `:401` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B18 | `:405` `if err := tx.Commit(); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B19 | `:413` `if nearlyZero(delta, filled) {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B20 | `:420` `if hadPrev && delta == 0 && sameSnapshot(obs, prev) {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B21 | `:421` `if err := tx.Commit(); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B22 | `:437` `if err := upsertFillSnapshot(ctx, tx, scope, obs, now); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B23 | `:441` `if correction {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B24 | `:442` `if err := recordExecutionCorrection(ctx, tx, ExecutionCorrection{` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B25 | `:464` `if delta > 0 {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40 | no | yes |
| B26 | `:465` `if _, err := tx.ExecContext(ctx,` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B27 | `:504` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B28 | `:512` `if obs.Terminal && locallyOwned {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40 | no | yes |
| B29 | `:516` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B30 | `:526` `if locallyOwned {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40 | no | yes |
| B31 | `:536` `} else if err := latchRiskBucketFillFailureForScope(ctx, tx, applied,` | `TestNonConfirmedFillDoesNotEnterTradeOutcome` | no | yes |
| B32 | `:527` `if err := j.applyRiskBucketFillInTx(ctx, tx, applied); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B33 | `:530` `if err := j.releaseTerminalRiskBucketOrderInTx(ctx, tx, applied); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B34 | `:533` `if err := applyWeeklyReservationLifecycleInTx(ctx, tx, applied); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B35 | `:536` `} else if err := latchRiskBucketFillFailureForScope(ctx, tx, applied,` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B36 | `:550` `if locallyOwned {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40 | no | yes |
| B37 | `:551` `if err := j.runApplyHooks(withOutcomeRefusals(ctx, &refusals), tx, applied); err` | `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B38 | `:556` `if err := tx.Commit(); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
