# a179 — 초기 데이터 종류와 현재 입력을 동일 스냅숏에서 비교

## Why

초기 입력의 종류를 복원했을 때 Jev 판단이 달라지는지 비교하려면 종목·시각·질문·가격을 동일하게 고정해야 함. 과거 고수익 구간은 이미 확장된 s3 입력이므로 이 변경은 그 수익률의 재현 또는 최초 실행 전체의 복원을 주장하지 않음. 데이터 종류만 복원한 `initial-fields-v1`을 제공함.

## What Changes

- 기존 shadow 프로브에 `--input-profile current|initial|compare`를 추가함. 사용자 후속 지시(2026-10-31)에 따라 기본값 compare로 현재·초기 입력을 매 회차 자동 측정함. 명시적 current는 기존 행동·원장·ID·질문을 보존함.
- 한 번 수집한 공개 상태에서 초기 필드만 투영함. 수집기·판단 클라이언트·가격 라벨링을 재사용하며 새 외부 의존성을 추가하지 않음.
- 실제 전송 상태와 전체 수집 상태의 digest, 입력군, 비교 성공·실패를 구분함. 같은 관측의 가격 라벨을 공유함.
- 입력군별 보고와 동일 스냅숏 쌍의 J2 판단 변화·선택 성과를 오프라인에서 확인함. 기존 사전등록 평가와 원장을 오염시키지 않음.

## Scope and Non-goals

대상은 `tools/a128-jev-shadow-probe/`, 비교 분석 진입점·fixture·문서, 본 change와 PM 연결뿐임. internal/cmd, 운영 타이머·설정, 주문·손절·사이징·실계좌 원장, 수집 API 확장, 실제 모델/브로커 호출, 기존 사전등록 임계값·시간·비용·과거 결과 변경은 제외함. 기존 a128 측정 창을 종결하지 않음.

## Risk and Verification

Normal: 신규 측정 도구 기능. 실계좌 intent journal은 변경하지 않으나 shadow JSONL 추적성과 호환성은 필수임. additive optional 필드와 모드별 자동 출력 디렉터리를 사용함. 테스트는 대상 Go 패키지, 필요한 패키지 race/vet, 비교 분석 fixture, 해당 change strict validate 및 PM 연결 검사로 제한함. 전체 저장소 시험·전체 gate·실제 API 호출은 실행하지 않음.

## Ownership and Authorization

사용자 2026-10-31 요청: Manager가 OpenSpec 작성·검토·완료 검증, Terra 단일 구현자가 코딩·대상 테스트·gstack 리뷰 수행. Manager 검증 후 Terra가 아카이브·커밋·푸시함. 운영 배포·측정 시작은 본 요청의 완료 조건에 포함하지 않음. Story: STORY-TOS-a179 / Feature: FEAT-TOS-006.
