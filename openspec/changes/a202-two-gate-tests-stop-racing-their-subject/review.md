# a202 review

## 리뷰 등급과 생략 사유 (2026-11-25)

- **gstack 리뷰: 생략.** 사용자 지시 「간단한건 리뷰 할 필요 없음」(Manager 전달). 근거: 시험 전용 change — 바뀌는 것은 기존 시험 함수 두 개의 대기
  문장뿐이고 생산 diff 0(`git diff <base>..HEAD --stat -- '*.go' ':!*_test.go' 'schemas/'` 빈 출력, `analysis/verify/production-diff.txt`).
- **변이 원장: not-applicable** — `docs/WORKFLOW.md` 「비례 원칙」: High-risk 경로 · 게이트 판정을 바꾸지 않는 시험 전용 변경. 대신 사본 대조 실험이
  그 자리를 맡는다: 편집 전 RED(생산 unlock↔close 사이 sleep 에서 옛 시험 FAIL · 부모 지연에서 옛 시험 FAIL), 편집 뒤 같은 지연에서 GREEN,
  양성 대조(사본에서 생산 닫기 호출 삭제 → 새 시험 FAIL), 무변이 대조군 GREEN(design §3, `analysis/red/` · `analysis/green/`).
- **다중 적대 리뷰: not-applicable** — 같은 비례 원칙. 주문 · 위험관리 · 원장 · 대사를 건드리지 않는다.
- **proposal-freeze 리뷰:** 경량 — `openspec validate --strict` + 작성자 셀프 검토 + Manager 검증(WORKFLOW 「리뷰 게이트」의 UI · 문서 · 도구 등급).

## Function Logic Map

- 적용(면제 아님). 편집한 기존 Go 함수 2개(시험 함수)마다 번들 — `analysis/function-logic/internal-shadowsession--testchildlockandcrash` ·
  `cmd-tossctl--testthereplacedreaderisclosed`. 편집 **전에** 만들어 커밋했고 편집 뒤 AST 를 재추출해 갱신했다.
- 도우미(`a109WaitFor` · `fixture`)는 고치지 않았다.

## Pre-Edit 선언 (2.x 착수 전)

- 대상 심볼: `TestChildLockAndCrash`(`internal/shadowsession/session_linux_test.go:210`), `TestTheReplacedReaderIsClosed`(`cmd/tossctl/a109_the_reattachment_does_not_strand_itself_test.go:191`).
- 호출부: 없음(시험). 생산 호출 대상 `Begin` · `strategyRuntimeAttachment.attempt` 는 바꾸지 않는다.
- a166 census: 대상 파일 · 함수명이 `internal/accountbudgetruntime` 에 없음(`analysis/code-context/a166-census-check.txt`).
- 불변식: 생산 diff 0, 단언 · 기한 · 헬퍼 불변, 새 가드 · 헬퍼 · 의존성 0.
- 실패 시험(RED): `analysis/red/` 네 파일.
- rollback: 커밋 revert. 설정 영향 없음.
