# Evidence reconciliation

- 기준 commit: `4d7d15571` (`base-commit.txt`)
- 변경 표면: `tools/gate.sh`(11단계), 새 `tools/sdd/check_openspec_validate.py`, 새 시험 `tools/sdd/test_gate_validate_step.py`,
  `tools/sdd/test_race_detector_actually_runs.py`(단계 수 단언 1곳), `docs/WORKFLOW.md`, 스펙 델타 `sdd-workflow`

| 항목 | CodeGraph | CodeGraphContext | 현재 HEAD·실행 | 결론 |
|---|---|---|---|---|
| ⑪ 이 부르는 명령 | 색인 밖 | not-applicable | `tools/gate.sh:338` 순회의 `validate` → `Makefile:131` | HEAD 기준 |
| gate.sh 구조를 단언하는 시험 | `GATE` 변수 1건 | — | `git grep` 3파일(baseline 표) | 단계 수 단언 1곳만 영향 |
| `make validate` 의 다른 소비자 | 색인 밖 | — | CI 없음, 문서 서술만 | `make validate` 타깃은 그대로 둔다(D2) |
| 오늘 ⑪ 의 실측 | — | — | `analysis/strict-debt/`: strict 51/59 실패(경고 262, 오류 0), strict 없음 59/59 rc 0 | Manager 실측과 일치 |

불일치로 편집을 막을 항목은 없다. Go 함수 변경 0 이라 FLM 은 not-applicable(`review.md`).

## 기억 회고

- Manager(2026-11-16): ai-memory 에 이 단계의 엄격도에 관한 이전 결정 없음. a173 은 「11/11 make validate OK」 통과 기록만.
- `scripts/memory-recall.sh "openspec validate strict"`(이 세션): 파일 기억 0건, GBrain advisory 결과 없음.
- 선례(HEAD 의 문서): `openspec/changes/a063-align-attestation-renewal-profile/issues.md:56` — 「`make validate` returned exit 2
  because the unrelated `a119-…`」. **남의 change 때문에 닫는 change 의 ⑪ 이 막힌 것은 이번이 처음이 아니다.**
