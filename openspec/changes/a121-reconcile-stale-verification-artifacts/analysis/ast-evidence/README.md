# a121 개정 1판 — 분기 근거용 AST 산출물

개정 1판(design.md 「Revision 1」)이 **함수 내부의 분기**를 근거로 쓰는 자리마다, 그 함수의
`tools/logic-map` AST 를 먼저 만들었다(FLM-first). 이 디렉터리는 근거 열거만 담는다 —
Function Logic Map·Branch Test Map·risk report 번들은 편집 전 산출물이고 task 1.2 의 몫이다.
그래서 게이트 5단계가 읽는 `analysis/function-logic/` 가 아니라 여기에 둔다.

- 생성: `go run ./tools/logic-map --file <file> --func <func>`
- 생성 시점 HEAD: `cb378a6332dea5ae659590886ba8326effe4b26e` (2026-09-27)
- 각 파일의 `source_sha256` 이 그 시점 소스를 묶는다. 소스가 바뀌면 이 인용은 낡는다.

| 파일 | 함수 | 분기 | 근거로 쓰는 분기 |
|---|---|---:|---|
| `verifylive--outstandinglines.ast.json` | `record.go:542-567` `outstandingLines` | 6 | B3(키 = `Kind\x00ID`), B4(terminal 뒤 비-terminal 은 되살리지 않음), B6(비-terminal 만 투영) |
| `verifylive--artifact.terminal.ast.json` | `record.go:575` `Artifact.terminal` | 0 | 단일 술어 `Cancelled \|\| Filled` |
| `verifylive--decodeentry.ast.json` | `record.go:437-447` `decodeEntry` | 2 | B2(`FormatVersion > 1` 만 거절 — 모르는 `Kind`·필드는 거절하지 않음) |
| `verifylive--runner.m0recoverpending.ast.json` | `m0_recovery.go:128-201` | 20 | B2(OPEN·CLOSED 둘 다), B3·B14(페이지 상한), B4(반복 커서), B6(읽기 실패), B12·B13(빈 커서) |
| `verifylive--runner.livecount.ast.json` | `mutate.go:672-685` `Runner.liveCount` | 3 | B2·B3(outstanding 이 노출 상한에 셈해짐) |
| `verifylive--new.ast.json` | `runner.go:320-425` `New` | 26 | B9(`:346` trigger 모드는 outstanding 이 있으면 거절) |
| `verifylive--runner.stepconditionalcancel.ast.json` | `steps.go:778-795` | 4 | B3(취소 뒤 by-id 읽기가 **어떤 오류로든** 실패하면 `gone_after=true`) |
| `official--client.conditionalordersraw.ast.json` | `conditional_reads.go:156-211` | 7 | B1(status 필수 — 오류 문구가 OPEN/CLOSED 어휘를 적음), B6(읽기 실패), B7(행 복사 — 스냅숏 필드 없음) |
| `tossctl--resolveverifyaccount.ast.json` | `verify.go:922-953` | 7 | B5·B6(DisplayName 이 빈 값이 아닌 **첫** 계좌), B7(숫자 아닌 id → seq 0) |
| `tossctl--resolveverifyrecordfor.ast.json` | `verify.go:764-777` | 3 | B1(`--record` override 가 이긴다), B2(`--config-dir`) |
| `tossctl--buildverifybroker.ast.json` | `verify.go:874-906` | 5 | B5(seq 0 이면 지연 해석 클라이언트) |
| `tossctl--validatem0triggermode.ast.json` | `verify.go:405-431` | 9 | B9(`:427` outstanding 이 있으면 trigger 모드 거절) |
