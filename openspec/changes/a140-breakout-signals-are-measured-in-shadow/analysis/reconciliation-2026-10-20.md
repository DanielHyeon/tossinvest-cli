# a140 측정 정합성 점검 — 2026-10-20

범위: 기존 그림자 원장의 읽기 전용 대조. 기준 snapshot captured_at은 2026-10-20 09:56:27 KST(UTC 00:56:27)다. 새 측정, API 호출, 주문, 설정 변경은 하지 않았다. 원장 원문·식별자·종목·가격은 이 문서에 넣지 않았다. 기계 판독 집계는 [reconciliation-2026-10-20.json](reconciliation-2026-10-20.json)에 보관한다.

- 스냅샷: `analysis/shadow-ledger/ledger.jsonl`, SHA-256 `ce4c672551237a1c5f112ff69b330edcffc61a43084616df894a9fbcbea897a2`. 읽기 전후 해시와 바이트 수가 같았고 JSONL 해석 오류는 0건이다.
- 종류: universe 3, signal 228, label 14, gap 52, budget 84, session 2.
- KRX 세션 키는 3개다. 2026-10-16 및 2026-10-19는 각각 session summary 1개, proposed 7개·label 7개, orphan·중복 연결 0개다. 2026-10-20은 session summary가 없는 진행 중 키이며 proposed 2개·label 0개다.
- 완료한 두 세션의 총 라벨은 14건이고 결과는 close 13·target 1이다. `tools/a140-breakout-signal-shadow/report.go`의 n<30 규칙상 결과 판정은 보류다.

D9/tasks 3.1은 KR 5세션과 D11 호출 예산의 사람 승인 기록을 요구한다. 현재 evidence에는 closed KR 2/5만 있고, 현재 a140 change/output 검토 경로에서 preserved D11 approval receipt를 찾지 못했다. 따라서 tasks 3.1–3.3과 4.1–4.3은 계속 미완료다. 현재 ledger는 untracked이며 commit 주장을 하지 않는다.

이 문서는 a128의 시장별 두 세션 측정 요건이나 D16의 사람 임계 판단과 a140 D9의 5 KR 세션 요건을 서로 대체하지 않는다. 수익성·S4 착수·정책 승급을 결론내리지 않는다.

Skill applicability: research-first — 기존 report/ledger 소스 재사용; grill-me — 기존 승인 계약 변경 없음; impeccable — N/A, 문서 정합성 정리이며 UI 영향 없음.
