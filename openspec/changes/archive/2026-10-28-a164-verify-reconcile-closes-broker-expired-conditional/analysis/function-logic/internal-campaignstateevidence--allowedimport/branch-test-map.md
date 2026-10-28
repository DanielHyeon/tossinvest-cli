# Branch Test Map: `allowedImport`

RED/GREEN 영수증: `../../red/red.txt` · `../../red/green.txt`, 변이: `../../red/mutation.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 파싱 불가 원문 | `TestInspectionPackageHasNoRuntimeConsumers`(walk 오류 경로, 기존) | n/a(기존) | yes |
| B2 | import 반복 | `TestGuardRejectsRuntimeCaptureUse` | n/a(기존) | yes |
| B3 | 경로 리터럴 해독 실패(방어) | 없음 — 파서가 먼저 거절 | n/a | n/a |
| B4 | lane import 별칭 기록 | `TestGuardRejectsRuntimeCaptureUse` | n/a(기존) | yes |
| B5 | 이름 있는 lane import | `TestGuardRejectsRuntimeCaptureUse` | n/a(기존) | yes |
| B6 | lane 점 import 거절 | `TestOfflineAssessmentIsTheOnlyNamedConsumer`(dotCapture) · `TestGuardRejectsRuntimeCaptureUse` | M5 CAUGHT | yes |
| B7 | 허용 목록 밖 evidence import 거절 / 평가 파일 허용 | `TestInspectionPackageHasNoRuntimeConsumers` · `TestOfflineAssessmentIsTheOnlyNamedConsumer`(other.go) | yes(red.txt receipt 2) · M3 CAUGHT | yes |
| B8 | `offlinecontinuation` import 거절(평문 · 이스케이프) | `TestOfflineAssessmentIsTheOnlyNamedConsumer` | yes(red.txt) · M1 CAUGHT | yes |
| B9 | 평가 파일의 lane 캡처 허용(루프 뒤) | `TestOfflineAssessmentIsTheOnlyNamedConsumer` · `TestInspectionPackageHasNoRuntimeConsumers` | yes(red.txt receipt 2) · M4 · M5 CAUGHT | yes |
| B10 | 캡처 selector 탐색 | `TestGuardRejectsRuntimeCaptureUse` | n/a(기존) | yes |
| B11 | lane 별칭 캡처 거절(형제 파일 포함) | `TestGuardRejectsRuntimeCaptureUse` · `TestOfflineAssessmentIsTheOnlyNamedConsumer`(captureOnly) | M2 1 판 생존 → 시험 추가 뒤 CAUGHT | yes |
