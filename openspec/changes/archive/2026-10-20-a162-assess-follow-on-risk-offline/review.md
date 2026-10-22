# 설계 검토 — 2026-10-20

범위: Manager가 작성한 a162 설계와 PM 연결. 구현·실주문·운영·배포 승인이나 a156 완료 판정이 아니다.

## 승인과 재사용 판단

사용자는 후속 진입의 오프라인 판정부터 설계하도록 승인했다. 초기의 신규 순수 평가기 제안은 source 조사에서 기존 세 lane proposal을 확인한 뒤 철회했다. 새 설계는 기존 native proposal과 한 방향 capture를 재사용하고, 명시적인 위험 가정에 대한 잔여 예산 산술만 더한다.

search-first는 기존 source·시험·Go 표준 라이브러리에서 충분한 해법을 확인해 외부 패키지 검색·설치를 생략했다. grill-me의 입력·출력·실패·권한 경계는 design에 고정했다. impeccable: N/A — UI/CLI 변경 없음. 기존 a156/실측 조건을 완화하지 않는다.

## 독립 리뷰와 수정

독립 reviewer: `a148_independent`, 작성자와 별도 컨텍스트.

| 발견 | 조치 |
|---|---|
| reversal/weekly의 비공개 예산·통화 접근 경로 불명확 | 같은 호출의 native plan/risk에 기존 CaptureOfflineEvidence를 사용. 별도 artifact 입력·새 getter·역변환 금지 |
| 전략별 digest 의미가 다름 | continuation RiskBudgetDigest, reversal/weekly PlanDigest로 명시하고 family도 결속 |
| H와 weekly 예약 슬롯의 이중 계상 위험 | H에 포함되지 않은 추가분만 판정. 기존 예약을 빼거나 재시도 수량으로 사용하지 않음. 슬롯을 금액으로 더하지 않음 |
| invalidation 이유 소실 가능 | native kind/code/invalidation code 모두 보존 |

수정 후 독립 **설계 PASS**. 제품 구현 READY 또는 실행 시험 PASS는 아니다.

## gstack 검토

별도 reviewer `a148_gstack`가 getter/capture 경계 및 산술 입력의 파싱 전 자원 한도를 지적했다. 기존 capture 비용·크기 제약을 수용하고, ASCII canonical decimal 최대 78자리 검사 후 256-bit 확인 및 min-before-uint64-convert를 명세했다. 수정 후 gstack **설계 PASS**, 남은 actionable finding 없음. Manager는 source capture 필드, 변경 명세와 검증 결과를 대조해 설계 범위로 수용했다.

## 근거와 검증

- [source/AST/CodeGraph 대조](analysis/evidence-reconciliation.md): 현재 source SHA와 AST 3개 일치. 처음 qualified 이름 검색은 실패했으나 unqualified 검색과 file-specific node로 KR·US 3개 lane의 정의·caller를 확인했다. graph 결과는 실행 권한의 증거가 아니다.
- `openspec validate a162-assess-follow-on-risk-offline --strict`: PASS.
- PM 문서 생성 및 `python3 tools/pm/generate_master_tracker.py --check`: PASS. 다른 PM 작업을 변경하지 않았다.
- 설계 문서의 로컬 링크·공백 검사: PASS.
- 제품 코드·제품 테스트·전체 suite/gate·운영 변경: 실행하지 않음.

## 판정

설계만 검토한 단계다. 새 패키지의 입력 타입/의존 경계·구현 기준 commit·RED 시험 및 구현 전 READY 확인은 task 1.2 이후로 남긴다. 기존 a156 선별 시험은 신규 모듈의 시험 결과가 아니다. 완료 task 체크, commit/push/archive, 실제 qFinal·저장·주문·마감 청산은 수행하지 않았다.

## 구현 후 검토 — 2026-10-20

위 내용은 구현 전 설계 검토 시점의 기록이다. 이후 사용자 “구현 진행”
승인으로 Terra가 구현하고, 별도 컨텍스트의 `a148_independent` 및
`a148_gstack`가 수정된 소스와 시험을 검토했다. 두 리뷰 모두 승인된
오프라인 범위 PASS이며 실행 권한이나 a156 완료를 뜻하지 않는다.

| 발견 | 최종 조치 |
|---|---|
| 산술 실패·예산 부족 시 후보 수량이 결과에 남음 | 모든 수량을 최종 성공 분기에서만 설정 |
| 잘못된 시각이 native 거절·invalidation을 가림 | native 거절을 시각 파싱보다 먼저 반환 |
| 얕은 복사로 입력 불변성을 충분히 검증하지 못함 | 여섯 양성 helper의 request digest와 native Capture→Encode snapshot 전후 대조, 변경 탐지 control 추가 |
| weekly 후속 fixture의 적용 범위 과장 가능 | ordinal 2 및 합성 날짜에 한정; 3–7·실측 증명 없음 명시 |

Manager는 현재 `assessment.go`의 typed adapter, same-request capture,
native 거절 우선순위, 수량 반환 경계와 증거를 직접 대조했다. 기존 lane
생산 코드 수정 없이 새 leaf를 추가했다. 기존 파일 변경은 기본 빌드에서
제외되는 weekly 시험 seam이다. 영향받는 기존 router 시험 1건도 최종
비캐시 검증에 포함했다.

검증 명령·해시·시험 한계는
[구현 검증 기록](analysis/implementation-evidence.md)에 보존한다.
새 모듈 일반/태그 시험 및 지정 router 시험은 비캐시 PASS다.
bounded-min 산술을 의도적으로 제거하면 시험이 실패했고 원본 해시로
복원됐다. 초기 fixture 기대/준비 오류는 제품 RED로 계산하지 않는다.

search-first: 기존 proposal/capture/codec 및 Go 표준 정수 산술 재사용.
새 외부 의존성 없음. grill-me: 승인 범위와 거절·가정·권한 경계 유지.
impeccable: N/A — backend/test/documentation-only, 화면 변경 없음.

최종 판정은 **로컬 구현 및 제한된 영향 범위 검증 PASS**다. non-nil FX
변경 검증, weekly ordinal 3–7 합성 fixture, 전체 전이 의존성 검사,
전체 suite/race/gate 및 실제 계좌 최신성은 검증하지 않았다.
실주문·운영 설정·기존 증거를 변경하지 않았으며 커밋·푸시·archive는
이 구현 턴에서 실행하지 않았다.

## 추가 통합시험 리뷰 — 2026-10-20

후속 사용자 승인으로 Terra가 시험과 weekly의 `tossos_testseams` fixture를
보완했다. 이전 검증 로그와 위 판정은 당시 범위 기록으로 보존한다.
이번 산출물은 [추가 검증 영수증](analysis/additional-integration/implementation-evidence.md)이다.

- KR·US 각각 ordinal 3–7의 선행 예약·체결을 native 전이로 구성했다.
  과거 예약의 소비 상태, 현재 예약, 누적 위험을 확인한다. Manager와
  두 리뷰어가 양성 결과의 포인터 존재만 확인하는 시험을 지적했고,
  B=1000, F=20×(순번−1), H=0, R=1000−F, U=100, 후보·가정 수량=2,
  계획 위험=200을 직접 비교하도록 보완했다.
- weekly US의 USD/KRW cross-currency 입력에서 FX non-nil과 capture
  전후 동일성, 실제 plan-bound FX 값 변조의 native/adapter 거절을
  검증했다. 변조 후 FX seal을 재발급하지 않았다. 다른 전략의 모든
  FX 조합이나 실제 환율 피드까지 검증한 것은 아니다.
- stale disclosure의 정확한 native 거절 코드와 수량 부재를 검증한다.
  합성 양성·음성 결과 모두 계좌 최신성 미확인 및 비권한 표시를 유지한다.
- 예약 map 삭제 반례는 재서명하지 않은 예약상태의 무결성 손상 시험이다.
  단순 조회 누락 분기 증명이라는 해석은 철회했다. 반환되는 native
  kind/code와 adapter code 및 수량 부재를 정확히 비교한다.

Manager는 최종 assertion·fixture 전이·원문 로그를 직접 대조했다.
독립 `a148_independent`와 gstack `a148_gstack`가 수정된 파일 해시와
비캐시 재검증 로그를 재확인해 각각 이번 추가 범위 PASS를 판정했다.
생산 `assessment.go`, 이전 구현 검증 기록, a128 원장 해시는 추가 작업
전후 동일하다. 기존 직접 소비자 시험 1건과 영향받는 시험만 실행했다.
research-first의 최종 선택은 공식 FX의 기존 tagged evidence와 native
FX 생성 함수 재사용이며, 초기 검토안의 production-authority builder를
실제로 사용했다고 주장하지 않는다. UI 없음, impeccable N/A.

실계좌 최신성 **자체**는 여전히 미검증이다. 이번 보완은 미확인 입력을
최신 계좌 근거로 승격하지 않는 경계의 검증이다. 운영 연결·실측·전체
suite/race/gate·커밋·푸시·archive는 수행하지 않았다.

## 게시 승인 및 범위 — 2026-10-20

사용자가 a162 검증분만 커밋·푸시·아카이브하도록 승인했다. 위 미게시
기록은 승인 전 검증 시점의 사실로 보존한다. 실제 게시 절차에서는
tasks의 검증 완료 항목을 코드·증거와 같은 커밋에 반영하고, main spec
동기화를 포함한 정식 `openspec archive`와 Story 경로 갱신을 수행한다.
현재 로컬 브랜치의 무관한 커밋은 게시 범위가 아니다.

전체 시험·전체 gate 미실행은 사용자 명시 제한이다. 현재 WORKFLOW와
archive 도구 확인 결과 이 change에 전체 gate를 요구하는 별도 계약이나
archive hook은 없다. 기존 영향 범위 시험, 독립·gstack 리뷰, Manager
검증 및 strict OpenSpec/PM 검사를 완료 근거로 사용한다. 이 결정은
다른 change의 검증 조건을 완화하지 않는다.
