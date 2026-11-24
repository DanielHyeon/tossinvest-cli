# a197 변이 원장 (tasks 4.1)

최종판은 **run5**(시작 sha `e8b09fe3f` — 구현 리뷰 r1 · gstack 리뷰 r2 수정 뒤 HEAD, 하네스가 단언). 원문 `run5/ledger.json` · `run5/raw/`.
하네스 `harness.py`: 저장소 루트는 파일 위치에서 유도, `git archive` 사본만 변이, 무변이 대조군(쓰는 시험 묶음 21개 전부 rc 0) 뒤 시작,
옛 문자열 1회 일치로 닿음 확인, 한 번에 한 변이, 변이마다 원래 바이트 복원 · sha256 대조, 전용 `GOCACHE`.
`-trimpath` 는 쓰지 않음 — 엔진 구조 핀(a087 · a111)이 `runtime.Caller` 경로를 열어 거짓 실패하기 때문(`analysis/verify/selected-tests.txt` 영수증).

## design §4 의 22개

| 판 | HEAD | 결과 |
|---|---|---|
| run1 | `7e830e99e` | M05 COMPILE · M20 · M22 SURVIVED |
| run2 | `fb5a4a0bc` | 21 CAUGHT · M22 SURVIVED(시험 선택 오류) |
| run3 | `ce2e53cad` | M22 CAUGHT(단독) |
| (run4) | `8b6ef490a` | M01–M08 CAUGHT 뒤 중단 — r2 수정으로 HEAD 가 바뀔 것이라 최종판을 run5 로 다시 돎. 부분 산출물은 커밋하지 않음 |
| **run5** | `e8b09fe3f` | **M01–M11 · M13–M22 = 21 CAUGHT, M12 NOT_RUN** |

- **21/22 CAUGHT, 생존 0, 실행 못 함 1.**
- M12(park 에 계정 래치 추가) — 만들 수 없음: 관측기에 진입 게이트 핸들이 없다(`ExitObserverOptions` 에 Gate 칸 없음). 계정 래치 부재는
  `TestA197AnAmbiguousCancelIsParkedWithoutAnAccountLatch` 가 엔진 조립의 진입 게이트로 직접 단언함. 「동등 변이」로 세지 않음.

## 추가 3개(design §4 밖 — gstack 리뷰 r2 수정을 잼)

| id | 변이 | 결과 | 이유 |
|---|---|---|---|
| R2a | `handleExternalBuy` 의 「종결 전 attempt」 갈래 삭제 | **SURVIVED** | 그 갈래에 닿으려면 게이트웨이가 Prepare 뒤 원장 쓰기에 실패해야 하는데, 결합 하네스(진짜 게이트웨이 · 원장)에 원장 쓰기 실패 주입 장치가 없음. 문장 · key 선택은 R2c 로, 판정은 R2b 로 잡히지만 **배선은 시험이 지키지 않음**(잔여 — review.md gstack r2 #1) |
| R2b | `externalBuyCancelHeld` 를 늘 거짓으로 | CAUGHT | `TestA197OnlyAnUnfinishedRecordedCancelHoldsTheSymbol` |
| R2c | `recordExternalBuyUncleared` 의 재시작 문장 선택 삭제 | CAUGHT | `TestA197TheUnclearedAlertSaysWhatReleasesTheStop` |
