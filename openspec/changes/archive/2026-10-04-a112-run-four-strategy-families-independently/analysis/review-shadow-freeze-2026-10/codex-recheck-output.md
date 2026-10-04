모델: GPT-6 기반 Codex · 세부 버전 미공개 · 「~/.codex 미접근」

**FAIL — 1차 P0/P1: CLOSED 5건, PARTIAL 1건. 새 P0 0건·P1 1건.**

CLOSED는 **설계 수준 종결**입니다. 생산 구현 검증을 뜻하지 않습니다. 아래 V2는 설계 브리프, A2는 amendment 파일 자체의 줄 번호입니다.

| 1차 ID | 판정 | V2 · amendment 좌표 | 근거·재현 |
|---|---|---|---|
| P0 함수 값 재결속으로 dispatch 접근 | CLOSED | V2:38–43,65–69,91–94 · A2:17,30,36 | 타입 해소·호출 폐포·소스 동결 추가. 스케치에서 Submit 재결속 식별, 소스 변이 검출 |
| P1-1 census 한계·같은 패키지 주조 | CLOSED | V2:35–43 · A2:17,36 계약 유지 | 별도 패키지 비공개 필드 주조는 컴파일 거절. 중첩 별칭·generic instance 검출. `any` 잔여는 소스 동결로 보완 |
| P1-2 반사실 입력 유실 | CLOSED | V2:10–15,55–61 · A2 직접 변경 없음 | 미선언·부분 ON 모두 관문 전 **3건**, 기존 선택 입력 **2건**. OFF 제안 1건 보존 |
| P1-3 활성화 ON 우선 판단 불가 | CLOSED | V2:73–78 · A2:17 | 같은 파도 OFF/OFF 술어 추가. OFF 관측 1건, ON 관측 0건 확인 |
| P1-4 모든 읽기가 Validate 경유한다는 주장 | CLOSED | V2:18–21 · A2 직접 변경 없음 | 외부 경계로 주장 축소. 기존 `Context.Read`의 검증 누락도 재현 |
| P1-5 정규 직렬화가 작성 출처를 증명한다는 논증 | **PARTIAL** | V2:24,28–31 ↔ A2:7,15,17 | 브리프는 수정됐지만 결정 63-v2가 여전히 결정 61의 잘못된 근거 (c)를 재채택 |

**새 P1 — 운반 타입의 dispatch 비도달 주장 충돌**

V2:55–59대로 `strategyProposalMarketAuthority`에 shadow 입력을 넣으면, 그 타입을 수신자로 받는 `dispatchHandoffs`도 입력을 보유합니다. 재귀 `go/types` 스케치가 이를 검출했습니다. 복사 접근자는 읽기 접근을 차단하지 않습니다.

shadow 운반 값과 dispatch 권한을 분리하거나, 보장을 “보유 금지”에서 “읽기·영향 금지”로 수정하고 검사 범위를 다시 정의해야 합니다. 실주문 재현은 아닙니다.

표적 시험 **9건 PASS**, 주조 공격은 기대한 컴파일 실패였습니다. 전체 시장 주기·handoff·gateway 차등 시험은 미수행입니다. 새 P2는 없습니다.

**실저장소 시작·종료 HEAD와 `status --short` 동일**—두 비교 모두 `cmp exit 0`. HEAD는 `2817064cc1f78c4d0d6abc04211de906a7c3efc5`. 원본 소스·모듈 **1,916개 바이트 변경 0건**입니다.

[상세 판정·재현·한계](review-recheck/report.md) · [시험 로그](review-recheck/tests-final.log) · [저장소 비교](review-recheck/repository-check.txt)