# a177 design

## 원칙
- 조회 창(일수)·시장·lane·lineage 네 값의 **표시 정본은 하나**다: `/performance-history` 화면이 이미 렌더하는 서버 조회 기본값 객체. 최적화의 성과·이력 카테고리는 그 객체를 받아 같은 문구를 렌더한다(spec operator-console 851 「요약을 재계산하지 않는다」준용).
- 값(일수)을 고치는 일과 표기를 단일화하는 일을 분리한다. 이 change 는 표기만 고친다. 일수가 잘못된 것으로 판명되면 `issues.md` 에 사람 결정 항목으로 남기고 멈춘다.

## 조사 순서 (1단계, 코드로 확정)
1. `internal/console/templates_optimization.go:217` 리터럴의 템플릿이 받는 데이터 타입과 그 핸들러.
2. `/performance-history` 의 「서버 조회 기본값」 칩이 읽는 필드(일수·시장·lane·lineage)와 그 값의 출처(상수·설정·서버 응답·빈 값 기본). "0일" 이 **의도된 기본값**인지, 아니면 seam 미배선·빈 값이 0 으로 렌더된 것인지(spec 706: 미측정은 0 으로 꾸미지 않는다 — 0일이 '값 없음' 의 가면이면 그 자체가 두 번째 결함이다).
3. 결과를 proposal 「가설」 절에 H1/H2 중 하나로 확정해 적는다.

## 구현 (H1 기준, H2 면 표기 단일화까지만)
- 성과 이력 뷰 모델의 기본값 요약(문구 생성 포함)을 공용 헬퍼로 두고, 최적화 카테고리 뷰 모델이 같은 헬퍼 결과를 필드로 받는다. 템플릿 리터럴 삭제.
- 값이 없으면 두 화면 모두 "미측정 · <사유 코드>" 를 같은 형식으로 렌더한다(0일로 꾸미지 않음).

구현 결과: unwired → 두 화면 모두 `미측정 · seam_unwired`; LoadErr → 두 화면 모두 요청 창 상수 표기(0일 아님), 실패 사유는 성과 이력의 기존 오류 줄 1회(Manager 수용 2026-10-25).

## 시험
- RED: 같은 fixture 로 `/performance-history` 와 `/optimization?category=performance-history` 를 렌더해 조회 창 문구가 같음을 단언 → 현재 FAIL(30일 vs 0일).
- GREEN 뒤 변이: 리터럴 "최근 30일" 을 템플릿에 되돌리면 FAIL 해야 한다(기록).
- 빈 값 fixture: 두 화면이 같은 미측정 형식을 렌더함을 단언(빈 표본 통과 방지: fixture 가 실제로 두 화면을 렌더했는지 길이 > 0 단언).

## 영향 범위
`internal/console/templates_optimization.go`, 성과 이력 뷰 모델 파일(1단계에서 확정), `internal/console/*_test.go` 신규 1 파일. 라우트·핸들러 등록·POST 0.
