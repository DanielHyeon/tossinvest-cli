# a112 base 재고정 aeeb209e → 1e25b3a3 (사용자 승인 2026-10-04, Manager 전달)

docs/WORKFLOW.md 「사람 승인 base 재고정」 세 요건:

1. **귀속 실측**(a071 변형 — 자기 Go 커밋이 0 이 아니므로 그 커밋들이 고친 기존 함수가 fresh 번들로 덮였는지 전수):
   `attribution-receipt-aeeb209e-to-f7da2162.tsv`(harness/a112_repin_receipt.py aeeb209e → HEAD f7da2162). 자기 Go 커밋 62(디렉터리 · 제목 `a112` 합집합),
   그 커밋들이 만진 함수 301: FRESH 51 · NOT-REQUIRED 240 · STALE 2 · MISSING 8.
   - STALE 2 는 **함수가 착지에 없다**(GONE): `strategy_proposal_ambiguity_test.go` 는 파일째 삭제 · `TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse` 는
     현재 트리에 그 이름 없음(번들 머리 「base 이름, 현재 트리에 없음」). 덮을 대상 없음.
   - MISSING 8 은 **자기 커밋이 아니다**: 제목 grep `a112` 이 다른 change 의 커밋을 함께 집었다 — 82080177 = a127(7 함수), 766a8456 = a094(classifyMutation).
     둘 다 새 base 1e25b3a3 의 조상이라 새 창 밖이다.
2. **새 창 영수증**: `window-receipt-1e25b3a3.tsv`(harness/gateprep_window_receipt.py, 대상 = HEAD f7da2162 워킹트리) — required 20, **a112 FRESH 20**.
   2026-10-04 게이트 준비 때 16 이었고, 그 뒤 8.5 응답 로트가 편집한 시험 함수 셋(옛 창에서는 새 함수, 새 창에서는 기존 함수)이 더해졌다 — 경량 번들
   `harness/render_repin_bundles.py`, 편집 전 번들 `pre-edit/`(178cc196 워크트리) — 그리고 loadFamilyActivation(기존 번들 FRESH).
3. **게이트 모의**: `check-analysis-simulation-1e25b3a3.txt` — 격리 워크트리(HEAD + 이 로트 파일 + base-commit.txt = 1e25b3a3, 모의 커밋)에서 check_analysis
   rc=0 「evidence complete or diff-proven exempt」(required 20, 창 안의 착지 커밋 28).
4. **단독 커밋**: base-commit.txt 한 파일 — 이 기록 커밋 다음.

옛 base 에서의 요구(2026-10-04 게이트 준비 분류 `../gateprep-2026-10-04/classification.txt`): 173 → 150(타 change 소관 missing 148 = 「형제 착지 몫」). 새 base 는 그
형제 몫을 창 밖으로 옮긴다. SHADOW 로트 착지 뒤 영수증을 최종 갱신한다(새로 편집할 함수 편입).
