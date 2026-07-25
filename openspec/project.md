# TossOS Project Context

## 제품

tossinvest-cli의 product fork로 만드는 토스증권 자동매매 시스템. Go 모듈식 모놀리스(기존 CLI·MCP 보존 + 엔진·데몬·웹 추가). 전체 계획: docs/ROADMAP.md.

## 규약 (openspec 스킬 경유 작업 포함 필수)

- 워크플로 계약: **docs/WORKFLOW.md** — 리뷰 게이트(proposal-freeze 시 gstack 리뷰 + `openspec/changes/<id>/review.md` 기록), 완료 게이트(`make gate CHANGE=<id>`), 역할 분리(Manager는 스펙·리뷰, Teammate가 구현)
- change 명명: 동사-선행 kebab (add-*, harden-*). 브랜치: `feat/p<N>-<change-id>`
- 스펙 형식: Requirement는 SHALL, Scenario는 `####` + WHEN/THEN. `openspec validate <id> --strict` 필수
- 불변 규칙: 주문은 공식 Open API만(WTS 조회 전용), 토스 계좌가 포지션 권위, 실계좌 주문 자동 테스트 금지, upstream 테스트 회귀 금지, upstream push 금지
- change 완료는 tasks 전 항목 체크 + review.md 존재 + test/vet/validate green 이후에만 선언하고 `openspec archive`로 확정한다

## 기술 스택

Go 1.25+(모듈 경로는 upstream 유지), 표준 라이브러리 중심(cobra + charmbracelet만 외부 의존), SQLite(원장 — 저장소 밖 ext4 경로), React/TS(운영 콘솔, Phase 5). 저장소는 NTFS 마운트(core.filemode=false, 실행 비트 없음 — 스크립트는 `bash path/script.sh`로 실행).
