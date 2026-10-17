Codex · GPT-6(세부 버전 미제공) · ~/.codex 미접근

**HOLD — B2의 조기 활성화 읽기가 취소된 수집을 붙잡습니다.**

| 등급 | 지적 | 재현·결과 | 제안 |
|---|---|---|---|
| P0 | 재현 없음 | 닫힌 시장의 emission·dispatch 콜백 모두 0 | — |
| P1 | B2 읽기 지연이 수집 반환을 막음 | `TestReviewB2ProductionReadDelay`: 전판은 읽기 없이 FX_NOT_READY. 후판은 취소 후에도 150ms 대기, 읽기 해제 후 반환 | 활성화 진단 적재에 취소·지연 경계 마련 |
| P2 | B2 닫힘 사유 불변 주장 깨짐 | `TestReviewB2NilGetenv`: FX 미준비+nil getenv에서 **FX_NOT_READY → INTERNAL_FAILURE** | nil 전제 검사와 다중 실패 회귀시험 |
| P2 | B1 미검증 값·개행 노출 | `TestReviewB1DiagnosticPayload`: 후판 오류에 합성 lane_id와 개행 그대로 포함 | 필드명/index 사용 또는 제한·인코딩 |
| P2 | A 골든과 실행값 결속 공백 | 골든 1.2→1.3 및 체크섬 갱신 사본에서 전체 breakout 시험 **PASS** | 골든 배열과 실행 경계 직접 대조 |

깨지 못한 공격:

- **A:** 동일 입력의 신규 평가·prior 재사용·정정 **9,632건**, 결정·seal 전후 동일. 1.2 플래그만 1,441건 변경.
- **B1:** 필드·핀·컨텍스트 조합 **16,224건**, 수락 여부·sentinel 전후 동일. 읽기 실패 일곱 모양에서도 Undeclared 전환 없음.
- **B2:** 닫힌 시장의 ON 관측은 0→4로 변경됐지만 제안·emission·dispatch는 0. B11/B14도 강제 실행해 전후 차단 확인.
- 코드 변이 네 건은 기존 시험이 탐지. 골든 의미 변경 변이는 살아남음.

지연은 **파일 읽기 overlay로 재현**했으며, 실제 운영 디스크 교착·손절 지연을 입증한 것은 아닙니다.

[상세 보고서·재현 명령·로그 안내](review-r2/REPORT.md)

실제 저장소의 시작/종료 `git status --short`는 **바이트 일치**했습니다. 실험 파일은 지정 사본 안에만 작성했습니다.