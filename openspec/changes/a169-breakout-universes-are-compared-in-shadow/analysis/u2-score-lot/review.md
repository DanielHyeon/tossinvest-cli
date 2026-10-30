# 순수 U2 점수 검증

계획 고정 `650003f72`; 새 source와 test만 추가한다. 입력 시간·봉 개수 → numerator → prior 날짜 중복 → history 수 → 선택 volume → S20 zero → int64 overflow 우선순위. KST 자정, 과거 정렬 복사, 오늘/미래 제외 및 경계 진단을 사용한다. 수식은 big.Int로 ratio와 increase를 별도 floor하고 각 결과의 int64 가능 여부를 검사한다. 선택 봉이 아닌 오래된 봉의 volume은 검사하지 않되 모든 prior 날짜의 중복은 거절한다. caller·선정·runtime·네트워크·ledger·CLI 수정 없음.

Go 1.25 선택 `go test ./tools/a140-breakout-signal-shadow -run '^TestA169U2Score' -count=1 -v` 0.005s 및 동일 범위 `-race -tags tossos_testseams` 1.016s PASS. 리터럴 two-floor 333333→499999와 단일 floor 500000을 구별한다. 19/20/25 봉, 순서 불변, 0 Time·오늘·미래, 전체 prior 중복, 영/비정수/NaN/Inf·2^53−1·2^53, 분모 0·결과 overflow, 장 경계와 오류 우선순위를 시험한다. 구현 전 RED는 함수 부재 컴파일 실패이며 이후 GREEN이다.

source-copy 변이 다섯 종류: today 포함, history 19, 첫 나눗셈 제거, 범위검사 제거, 두 내림을 단일 내림으로 합침 모두 거절. 독립 임시 Go 모듈에서 실제 새 source·test 복사본과 Candle의 Time/Volume 필드만 가진 작은 stub을 사용했으므로 제품 전체 실행 증거가 아니다. 오류는 assertion 또는 19봉 slice panic으로 탐지했다. 원문 gzip -n 및 원본 SHA 보존. 비test cmd/internal/tools 전체 Go AST의 `a169U2Score` 직접 호출 0; reflection·동적 간접 호출 전체 증명이 아니다.

수동 gstack checklist: 신규 순수 계산만, 외부 IO/SQL/인증/동시 상태/LLM/UI/새 의존성 없음. big.Int 변환 전 유효성·int64 저장 전 범위 검사, 실패 시 빈 결과. 시간 범위는 동일 KST일 및 장 내 제한. 별도 자동 전문 review는 미실행. search-first: 기존 domain.Candle·stdlib big.Int·sizing.go 선례 재사용, 새 dependency 없음; MCP codegraph는 이 worktree에 index가 없어 사용 불가, 원본 소스·AST를 직접 확인했다. ai-memory a169/U2 검색 결과 없음. impeccable N/A — UI 없는 순수 계산.

전체 tests/gate, 네트워크, 실제 market data·broker·운영, canonical 선정/union/tie, U2 runtime·전체 a169 freeze 및 측정은 미실행이다. 후보는 Manager 독립 검토 전이며 push하지 않는다.
