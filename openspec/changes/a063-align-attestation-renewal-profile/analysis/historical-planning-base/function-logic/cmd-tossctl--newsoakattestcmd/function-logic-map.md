# Function Logic Map: `newSoakAttestCmd`

- Source: `cmd/tossctl/soak.go` lines 199-241
- AST evidence: `ast.json` (0 branches)
- 2026-09-06 Manager 사후 문서 검토. 현재 AST와 테스트 결과를 정렬한 문서이며 새 pre-edit 증거로 소급하지 않는다.

## Inputs and invariants

명령 등록은 프로세스/주문 실행이 아니다. 새 플래그는 false가 기본값이다. 기존 argv 및 command annotation을 유지한다.

## Branches and early returns

| Branch | 현재 AST의 조건/역할 | 연결 테스트 | 실제 증거와 한계 |
|---|---|---|---|
| B1 | 분기 없는 명령 생성: 기본 OFF 플래그와 기존 플래그 등록 | TestSoakCommandsAreRegisteredAndReadOnly; TestSoakAttestDefaultDoesNotWriteRenewalStatus | 정상 경로 통과; 합성 happy-path ID |

## Calls and live bindings

| AST line | Callee | Evidence |
|---|---|---|
| 203 | `strings.TrimSpace` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 225 | `runSoakAttest` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 228 | `StringVar` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 228 | `cmd.Flags` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 229 | `IntVar` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 229 | `cmd.Flags` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 230 | `DurationVar` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 230 | `cmd.Flags` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 232 | `StringVar` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 232 | `cmd.Flags` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 233 | `StringVar` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 233 | `cmd.Flags` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 234 | `StringVar` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 234 | `cmd.Flags` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 235 | `BoolVar` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 235 | `cmd.Flags` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 237 | `StringSliceVar` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |
| 237 | `cmd.Flags` | AST에 기록된 호출; 오류/부작용 계약은 아래 불변식과 실제 코드에 묶임 |

## State mutations and fallbacks

명령 등록은 프로세스/주문 실행이 아니다. 새 플래그는 false가 기본값이다. 기존 argv 및 command annotation을 유지한다.

실제 오류 반환/로컬 view/파일 저장 경계는 현재 AST와 연결 테스트를 기준으로 검토했다. 표의 직접 주입 없음/계측 없음은 실행하지 않은 분기를 뜻한다. 관련 테스트의 성공을 모든 분기의 실행 증거로 바꾸지 않는다.

## Safety conclusion

라이브 주문·설정 변경·엔진 재시작은 수행하지 않았다. a063 독립 적대적 리뷰와 gstack 코드 리뷰는 통과했지만 원래 기준 커밋의 전역 FLM 및 운영 수용은 별도 차단 상태다. `analysis/verification.md`와 `review.md`의 실행 범위 및 한계를 따른다.
