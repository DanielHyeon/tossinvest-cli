# Contributing

tossinvest-cli에 기여해 주셔서 감사합니다.

## 개발 환경

- Go 1.25+
- Python 3.11+ (auth-helper)
- Google Chrome (auth login에 필요)
- Make

```bash
git clone https://github.com/JungHoonGhae/tossinvest-cli.git
cd tossinvest-cli
make build
make test

# auth-helper
cd auth-helper
python3 -m pip install -e .
```

## 브랜치 & 커밋

- `main`에서 feature 브랜치를 생성합니다.
- 커밋 메시지는 [Conventional Commits](https://www.conventionalcommits.org/) 스타일을 따릅니다. 한글 사용 가능.
  - `fix(auth): 브라우저 차단 해결`
  - `feat(portfolio): CSV 내보내기 추가`
  - `docs: README 업데이트`

## PR 가이드

1. `make test`와 `make lint`가 통과하는지 확인합니다.
2. PR 템플릿의 체크리스트를 채워주세요.
3. 거래(mutation) 관련 변경은 안전장치(config gate, confirm token 등)가 유지되는지 반드시 확인합니다.

## 프로젝트 구조

```
cmd/           # CLI 커맨드 정의
internal/      # 내부 패키지
auth-helper/   # Python 기반 브라우저 로그인 헬퍼
schemas/       # JSON 스키마
docs/          # 문서
```

## 메인테이너 되기

지속적으로 기여하다 보면 메인테이너로 승급될 수 있습니다. 진입 기준·권한·사임 규칙은 [`MAINTAINERS.md`](MAINTAINERS.md) 를 참고하세요.

## 주의사항

- 이 프로젝트는 토스증권 웹 내부 API에 의존합니다. API는 예고 없이 변경될 수 있습니다.
- 거래 기능은 기본 비활성입니다. 새로운 거래 기능 추가 시 동일한 opt-in 패턴을 따라주세요.
- 민감 정보(세션 토큰, 계좌 정보 등)가 코드나 로그에 노출되지 않도록 주의합니다.
