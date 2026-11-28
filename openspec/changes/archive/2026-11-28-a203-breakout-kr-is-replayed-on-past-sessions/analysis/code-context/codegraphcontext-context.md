# CodeGraphContext 보조 문맥 — a203

`not-applicable`: 단순 도구 change(a140 패키지 안 새 파일 + `dispatch` case 1개). 새 코드가 부르는 기존 함수(`evaluateSetup` ·
`labelProposal` · `planSizing` · `contiguousPrefix` · `signalRow` · `readLedger` · `checkEvidencePath` · `officialbars.PollClosedBars` ·
`scheduler.AdaptOfficialCalendar` · `strategyevidence.(*Store).SealBarSeries`)는 편집하지 않고 호출만 하며, 그 분기 근거는
`analysis/ast-evidence/claims.md`(logic-map 산출물)와 현재 HEAD 직접 읽기로 확인함. 보조 문맥 도구가 더할 결정 근거가 없어 돌리지 않음.
