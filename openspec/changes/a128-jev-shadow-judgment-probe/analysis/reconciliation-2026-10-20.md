# a128 s3 측정 정합성 점검 — 2026-10-20

범위: 기존 원장·state·실행 로그의 읽기 전용 대조. 기준 snapshot captured_at은 2026-10-20 09:56:27 KST(UTC 00:56:27)다. 새 측정, API 호출, 주문, 설정 변경은 하지 않았다. 원장 원문·식별자·종목·가격은 이 문서에 넣지 않았다. 기계 판독 집계는 [reconciliation-2026-10-20.json](reconciliation-2026-10-20.json)에 보관한다.

- 스냅샷: `analysis/shadow-ledger/ledger.jsonl`, SHA-256 `35014b3bf9a3a72e0ec91a580f9d8fd3d8e9b7b02d31b845af50ab0c2f764742`. 읽기 전후 해시와 바이트 수가 같았고 JSONL 해석 오류는 0건이다.
- 연결성: judgment 13,800건, 라벨 13,440건, 실현 라벨 13,380건, 명시 라벨 결손 60건, 현재 미라벨 360건이다. orphan·중복 judgment/label 연결은 0건이다. 360건은 모두 이 점검 시점에 아직 due 전이다.
- state: 참조 digest 6,900개 모두 파일 존재·원문 canonical-content 해시 일치다.

## 3.1·3.2 범위 한정 근거

현재 `a128-s3`의 같은 selection lot에서 아래 네 세션을 대조했다.

| 시장·날짜 | 판단 | 라벨 상태 | provenance | 실행 기록 |
|---|---:|---|---|---|
| KR 2026-10-16 | 1,980 | 실현 1,920 · 명시 결손 60 · 미라벨 0 | manual-override | `/tmp/a128-kr-1016.log` — sampling window closed 33 cycles |
| US 2026-10-15 | 1,980 | 실현 1,980 · 미라벨 0 | calendar | `/tmp/a128-us-1015.log` — sampling window closed 33 cycles |
| KR 2026-10-19 | 2,040 | 실현 2,040 · 미라벨 0 | manual-override | `/tmp/a128-kr-1019.log`의 첫 6 cycles는 context canceled, `/tmp/a128-kr-1019b.log`의 28 cycles는 closed; 합계 34 cycles |
| US 2026-10-19 | 1,980 | 실현 1,980 · 미라벨 0 | calendar | `/tmp/a128-us-1019.log` — sampling window closed 33 cycles |

각 선택 행의 마지막 due와 연결 라벨 시각을 대조한 결과, 이 네 세션의 라벨은 모두 due 이후이고 미라벨은 0건이다. KR 2026-10-16의 60건은 `price_read_failed` 계열의 명시 결손으로 보존한다. 정상 실현으로 재분류하지 않는다.

`manual-override`는 tasks 3.1이 허용한 기록형 달력 우회다. provenance는 state에 남아 있으며, session gap 시각도 보존한다. KR 2026-10-16 gap은 `2026-10-15T23:40:01Z`, KR 2026-10-19 gaps는 `2026-10-18T23:13:53Z`, `2026-10-19T00:57:50Z`다. UTC 날짜만으로 gap을 같은 장과 연결하지 않는다.

US 2026-10-16의 manual-override 부분 세션과 KR 2026-10-20의 진행 중 관측(미라벨 360건)은 위 판단에서 제외했다. 이 문서는 a128의 두 시장·두 측정일과 최종 라벨 완료라는 3.1/3.2 범위만 뒷받침한다. 예측력·수익성·90% 임계·정책 승급·4절 종료를 뒷받침하지 않는다.

## 기존 증거와 추적 상태

- `analysis/audit-2026-10-17.md`의 3.3 evidence-only commit은 당시 9,000 judgment·9,000 label 스냅샷을 뜻한다. 현재 append된 snapshot의 commit을 뜻하지 않는다.
- 기존 2026-10-17 부분-US와 audit 문서는 역사 증거로 유지한다. 이 문서는 이를 덮어쓰지 않는다.
- 현재 ledger는 수정 상태이고 current state 파일은 untracked다. 이 점검은 commit·archive·gate를 수행하지 않는다.

Skill applicability: research-first — 기존 report/ledger 소스 재사용; grill-me — 기존 승인 계약 변경 없음; impeccable — N/A, 문서 정합성 정리이며 UI 영향 없음.
