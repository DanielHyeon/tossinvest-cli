# a121 GREEN — 남은 RED 1: 도달 census 허용 목록에 기록 **읽기** 통로가 없다 (Manager 판정 필요)

- 작성: 2026-10-05, Terra(GREEN), worktree `/tmp/a121-green`(a99a9059 + GREEN 편집)
- 상태(갱신): **Manager 판정 2026-10-05 — 승인·수리 완료.** `reconcileAllowedOutside` 에 `readRecordRaw` 한 줄을 더했다.
  이것은 RED 시험 수정이지만 설계가 요구한 읽기 통로(원문 바이트 지문, design F3·P2-7)의 **RED 로트 누락 수리**다.
  허용 목록의 통로는 둘뿐이다 — **추가 통로는 Recorder 3이름(OpenRecorder·Append·Close), 읽기 통로는 `readRecordRaw`
  하나.** `os.ReadFile` 잎을 record.go 에 둔 선택(기존 `TestNoAutomationBypassExists` 의 「기록 파일은 record.go 소유」
  의도 준수)도 같은 판정으로 승인됐다.
- 판정 전 기록(아래)은 경위로 남긴다.

## 두 가드가 서로를 막는다

| 가드 | 요구 | 근거 |
|---|---|---|
| 기존 `TestNoAutomationBypassExists`(static_test.go:203-213) | verifylive 에서 `os.` 는 record.go·receipt.go 만 — "only record.go and receipt.go may own local durable evidence" | 기존 경계(변경 안 함) |
| RED `TestReconcileVerifyliveFilesReachOnlyAllowlistedCode` | reconcile*.go 가 부르는 reconcile 밖 함수는 허용 목록(LoadEntries·PendingCleanup·M0Unsettled·Digest·OpenRecorder·Recorder.Append·Recorder.Close·attest.Mask)뿐 | A-RED P1-1, RED 처분 ⑤ |
| design G2·F3·P2-7 | 엄격 해독은 **모든 비공백 줄**(LoadEntries 의 꼬리 관용 금지), 지문은 **파일 원문 바이트** sha256 | 설계 |

설계는 원문 바이트 읽기를 요구한다. 허용 목록에서 그것을 주는 함수는 없다 — `LoadEntries` 는 해독 뒤 Entry 만 주고 찢긴
꼬리를 버리며, `OpenRecorder` 는 쓰기 전용이다. 읽기를 reconcile*.go 에서 `os.ReadFile` 로 하면 static 가드가 깨지고
(실측: `reconcile_record.go reaches into os`), record.go 에 두면 census 가 깨진다(실측: `reaches …verifylive.readRecordRaw,
defined outside reconcile*.go and not allowlisted`). RED 로트가 허용 목록을 만들 때 추가 통로(OpenRecorder 3이름)만 넣고
읽기 통로를 빠뜨린 것이 원인이다(내 누락).

우회(정직하지 않아 쓰지 않음): `io/ioutil.ReadFile`·별칭 import 처럼 `os.` 문자열을 피하는 방법은 static 가드의 의도를
글자로만 지킨다.

## 현재 구현(GREEN)

`readRecordRaw(path) ([]byte, error) { return os.ReadFile(path) }` 를 **record.go** 에 새 잎으로 두었다(static 가드 의도 —
기록 파일 열기는 record.go 소유 — 준수). 대사 경로는 이것 하나로 원문을 읽는다(선택 전·추가 직전 두 번).

## 제안(한 줄, Manager 결정)

`reconcile_seal_test.go` 의 `reconcileAllowedOutside` 에 `verifylivePath + ".readRecordRaw": true` 한 줄 — 설명:
"OpenRecorder 의 읽기 짝, 원문 바이트만 돌려주며 브로커·실행기에 닿지 않음". 실측: 사본 트리에서 이 한 줄을 넣으면
verifylive 전 시험 GREEN(`go test ./internal/verifylive -count=1` → `ok`, census 포함 — 2026-10-05 사본 트리 실측;
변이 원장 하네스의 대조군은 같은 결과를 이 시험 하나만 `-skip` 해서 얻는다).
