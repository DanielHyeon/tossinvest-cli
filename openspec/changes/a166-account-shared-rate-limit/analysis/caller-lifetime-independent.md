# a166 caller lifetime 독립 설계검토

대상 /home/daniel/tossos-work/a166-runtime HEAD d2ce75b6229107a8bc5f8d3c9ef1c9ef93a806f1, uncommitted `analysis/caller-lifetime-plan.md`. 모두 원문 손 읽기. AST 추출·시험 미실행. lot1/2 분리 방안 자체는 가능하나 종료와 환경모드 계약 보완 전 종결 승인 보류.

## P1 — Execute 반환 전 직접 종료
`cmd/tossctl/monitor.go:25–38`은 `newAppContext(opts)` 호출 뒤 실패 probe에서 `os.Exit(1)`한다. app이 factory를 빌리면 root.Execute에서 돌아오지 않아 실행 helper Close가 누락된다. 정상/오류 종료 코드를 유지하는 반환 경로 및 Close fixture가 필요. `cmd/tossctl/engine.go:105–113,849–863`은 두 번째 신호에서 의도적으로 즉시 hard exit(130). 이는 일반 종료와 별개로 Close가 생략됨을 명시하고, 강제 종료 후 shared budget dirty/restart fail-closed 안전 근거와 비운영 fixture를 연결해야 한다. 즉시 종료 기존 계약을 임의로 바꾸라는 뜻이 아니다.

## P1 — 두 변수 정책을 실제 parser에 고정
`internal/accountbudgetruntime/runtime_linux.go:50–62` 현재 FromEnv는 단일 변수 부재를 OFF로 판단한다. 제안된 MODE/MANIFEST 쌍 전체를 원자적으로 스냅샷하고 두 값 존재/빈 값 구별하여 파싱해야 하며 이 기존 helper를 그대로 쓰면 안 된다. 승인 matrix: 둘 다 부재 OFF; 명시 off+manifest(빈 값 포함) 거절; required+manifest 부재/빈 값 거절; 빈/미지 mode 거절; mode 부재+manifest 존재(빈 값 포함) 거절. required+유효 manifest만 REQUIRED. 값 정규화/대소문자 허용 여부는 구현·fixture와 일치시킬 것. 초기 실패를 캐시하지 않는 것과 값 스냅샷은 공존 가능하나 재시도가 외부 env 변화를 받아 OFF로 전환되면 안 된다. 어떤 경로에서 최초 parsing 하는지(official 첫 생성 이전), help/OFF 파일 무접근, 모든 builder의 공통 경유를 시험한다.

## P2 — 재시도·Close·console 수명
`console.go:787–795,819–836` consoleBroker는 실패를 캐시하지 않고 재시도; `:839–852` 별도 context-aware builder; `:1166–1180,1227` verify는 매번 새 client를 구성한다. 공통 owner는 서로 다른 client를 억지로 합치지 말고 factory만 공유. 생성 실패 화면 표시를 유지하며 WTS 무단 fallback을 금지할 것. 종료 전 cancel/join은 계획에 조건부로만 되어 있고 실제 모든 background task 완료는 미검증: lot2 수락조건으로 실제 호출자·closure join과 context 없는 Background 경로 확인 필요. 이미 닫힌 owner에 대한 기존 client 사용, Close가 실패했을 때 후속 호출/재시도 정책과 동시 생성·종료 순서를 fixture로 고정할 것. `runtime_linux.go:120–137` 현재 Close 실패에서 closed=true로 설정하지 않으므로 root의 논리적 닫힘 규칙이 별도로 필요.

`main.go:11–21` error 출력은 redact.Writer, `root.go:52–60` rootOptions는 생성 함수 로컬, `engine.go:615–624` Context.Close는 journal만 닫는다. 단일 root 소유자와 borrowed app/engine은 이 구조에 부합한다. 단 Cobra 밖 helper 우회·hard-exit 예외 위 처리 전 '모든 종료 Close' 주장은 금지. `errors.Join` 실행 오류·cleanup 오류와 출력/exit code 의미 검증 필요.

## 승인 가능 범위
공유 factory/borrow·기존 오류 및 OFF 보존 원칙과 로트 분할은 조건부 설계 승인 가능. lot1 구현 시작에는 위 직접 종료/환경 matrix 및 Close 실패 판정을 계약으로 반영해야 한다. console 전 caller 결속, priority 분류, replay origin, actual multi-process wire 및 운용 활성화는 별도 후속 증거가 필요. 코드는 수정하지 않았고 운영/전체 시험은 실행하지 않았다.
