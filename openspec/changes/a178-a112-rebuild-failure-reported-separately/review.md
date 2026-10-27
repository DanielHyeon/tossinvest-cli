# a178 review

## proposal-freeze (2026-10-27, 경량: validate + Manager 셀프리뷰)
- 보이스: Manager(Opus 5.5) 단독. 측정 도구(운영 caller 0)의 진단 문구 변경이고 판정은 불변이라 적대 보이스는 해당 없음.
- 입력: a175 gate 1·2차 로그(`/tmp/a175-close-logs/gate.log`·`gate2.log`), HEAD `526650e50` 원문(`main.go:19,82,116`, `identity.go:155,470-495`), logic-map AST(`analysis/evidence/verifyPrescribedBinary.ast.json`), spec `breakout-retest-strategy-lane` 101행.
- 사람 결정: B-1(2026-10-27) — spec·120초 시한은 그대로, 원인 구분만.
- 기각한 대안: 시험 전용 시한 seam(시험 바이너리 digest 비교를 깨거나 운영 SHALL 을 끄는 스위치가 됨), e2e 시험 gate 분리(B-2, 게이트 구조 변경), 시한 분리(B-3, spec 변경).
- 남은 위험: 이 수정 뒤에도 바쁜 기계에서는 e2e 시험이 시한 초과로 실패할 수 있다. 달라지는 점은 메시지가 원인을 정확히 말한다는 것이고, 재시도 관례(WORKFLOW 1줄)로 다룬다.
- 결론: proposal-freeze 승인. base 는 a175 archive 뒤 main HEAD.

## 구현 리뷰 (2026-10-27, gstack 경량)
- 적용성: Low 측정 도구의 오류 분류 변경이다. `docs/WORKFLOW.md` 비례 원칙에 따라 다중 적대 보이스·전문가 dispatch는 적용하지 않고, gstack checklist의 critical/informational 항목을 구현 범위에 직접 대조함.
- 범위: 원격 `origin/main`은 `0e8ba1c7c`에 머물러 a176·a178 정의까지 포함하므로, Manager가 지정한 `4b868a18d` 이후 구현 커밋(`d62d915a8`, `c43d3420a`, `800b6df90`)만 검토함. 다른 change의 내용은 수정하거나 판정하지 않음.
- 확인: B6~B8은 rebuild 오류·명령 불일치·digest 불일치만 순서대로 분리하며 builder 호출 1회, caller context, 120초 deadline, HOLD 변환과 B1~B5/B9는 불변임. fake builder의 zero-value command는 기존 prescribed command를 유지하고 오류·명령 주입만 시험에 노출함.
- checklist 결과: SQL/데이터, 동시성, LLM trust, shell injection, enum completeness, async/time window/type boundary, 배포 항목에 신규 표면 없음. 미해결 finding 0건.
- 증거: `go test ./tools/a112-mb-us-source/ -count=1` PASS(156.072s), `go vet ./tools/a112-mb-us-source/` PASS, `gofmt -l tools/a112-mb-us-source` 출력 없음, `check_analysis.py` required 3·evidence complete. 원문은 `analysis/verification-output.txt`.
- 변이: `%w`→`%v` 1회에서 deadline wrapping 시험이 FAIL하여 killed, 즉시 `%w`로 원복하고 대상 리터럴 1개·`%v` 0개를 확인함. 원문은 `analysis/mutation-output.txt`.
- gstack 한계: 이 격리 worktree의 CodeGraph는 초기화되지 않아 질의가 불가했음. gstack local review log는 start capture 부재로 freshness `UNVERIFIED`이며 Manager 독립 리뷰를 대체하지 않음. 현재 소스·AST/FLM·패키지 시험을 직접 대조했으며, 기능상 남은 위험은 바쁜 호스트의 rebuild deadline뿐임.
- 결론: 경량 리뷰 통과. Manager의 4.3 독립 diff·시험 원문 검증은 미수행으로 유지.
