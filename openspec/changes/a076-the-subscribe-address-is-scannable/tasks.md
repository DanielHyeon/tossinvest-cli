# a076 · Tasks

## 1. 편집 전 선언

- [x] 1.1 기존 함수를 하나도 편집하지 않는 배치를 확인한다 — 새 패키지, 새 파일,
      템플릿 상수뿐.
- [x] 1.2 `check_analysis.py`가 요구하는 함수가 0개임을 확인한다.

## 2. QR 인코더 (`internal/qr`)

- [x] 2.1 GF(256) 테이블과 곱셈 (0x11D, 생성원 2).
- [x] 2.2 Reed-Solomon 생성 다항식과 EC 코드워드.
- [x] 2.3 BCH — 형식 정보(15,5)와 버전 정보(18,6).
- [x] 2.4 레벨 M 버전 1–10의 블록 배치표.
- [x] 2.5 바이트 모드 비트스트림 — 모드·길이·종단자·패딩.
- [x] 2.6 블록 분할과 인터리빙.
- [x] 2.7 기능 패턴 — finder·separator·timing·alignment·dark module.
- [x] 2.8 형식 정보 두 벌, 버전 정보 두 벌(v≥7).
- [x] 2.9 지그재그 데이터 배치.
- [x] 2.10 마스크 8종과 벌점 4규칙, 최저 선택.

## 3. 검증

- [x] 3.1 형식 정보를 표준의 공표값 8개와 비교한다.
- [x] 3.2 버전 정보를 표준의 공표값 4개와 비교한다.
- [x] 3.3 EC 코드워드를 신드롬 0으로 검증한다 (인코딩과 다른 방향의 계산).
- [x] 3.4 **데이터 모듈 수를 표준의 용량표와 대조한다** — 배치 코드와 무관한 출처.
- [x] 3.5 구조 — finder·separator·timing·alignment·dark module.
- [x] 3.6 왕복 — 행렬에서 payload를 다시 읽는다.
- [x] 3.7 버전 1–10 전부가 실제로 도달 가능함을 확인한다.

## 4. 카드 표시 (`internal/console`)

- [x] 4.1 `NotificationQR`이 구독 주소를 모듈 좌표로 만든다.
- [x] 4.2 수평 런으로 병합해 요소 수를 줄인다.
- [x] 4.3 정적 여백 4모듈.
- [x] 4.4 템플릿이 정수 좌표로 `<rect>`를 출력한다 — `template.HTML` 없음.
- [x] 4.5 전송이 꺼져 있으면 렌더하지 않는다.
- [x] 4.6 인코딩 실패는 조용히 생략한다.

## 5. RED → GREEN

- [x] 5.1 카드가 SVG를 그린다.
- [x] 5.2 그려진 심볼이 카드가 표시하는 주소의 심볼과 같다.
- [x] 5.3 여백이 사방 4모듈 이상이다.
- [x] 5.4 주소가 없으면 그리지 않는다.
- [x] 5.5 꺼져 있으면 그리지 않는다.
- [x] 5.6 SVG 안의 속성이 전부 숫자다.

## 6. 변이 검증

- [x] 6.1 N1 — 정렬 패턴 판정을 예약 여부로 되돌린다 → RED
- [x] 6.2 N2 — 형식 정보 BCH 생성다항식 변경 → RED
- [x] 6.3 N3 — RS 생성 다항식의 근 순서 변경 → RED
- [x] 6.4 N4 — 여백 제거 → RED (첫 시도에서 아무 테스트도 죽지 않아 테스트를 고쳤다)
- [x] 6.5 N5 — 다른 주소를 인코딩 → RED
- [x] 6.6 N6 — 꺼진 상태에서도 렌더 → RED
- [x] 6.7 N7 — 런 병합이 밝은 칸을 넘어간다 → RED
- [x] 6.8 변이 후 모든 파일이 바이트 동일하게 복원됨을 확인한다.

## 7. 게이트

- [x] 7.1 `go build ./...` · `go vet ./...` · `go test ./... -count=1`
- [x] 7.2 `openspec validate --all --strict`
- [x] 7.3 `check_analysis.py --change a076-the-subscribe-address-is-scannable`
- [x] 7.4 `make sdd-sync` → `make sdd-check`
- [x] 7.5 PM story/registry/feature/generated
- [x] 7.6 `make gate CHANGE=a076-the-subscribe-address-is-scannable`

## 8. 배포 후 실측

- [ ] 8.1 배포 후 실제 폰 카메라로 카드의 QR을 찍어 ntfy 앱이 열리는지 실측한다.
