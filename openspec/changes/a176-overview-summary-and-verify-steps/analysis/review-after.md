# a176 UI·접근성 검토

## 도구 상태

- `design-critic`: PATH에서 찾지 못함. a174 선례와 같이 수동 렌더 검토 및 `impeccable detect`로 대체함.
- `fixing-accessibility`: 새 상호작용은 네이티브 `details`·`summary`만 사용하도록 검토함. 사용자 정의 버튼·ARIA 상태·키보드 핸들러를 추가하지 않음.

## 확인 결과

- `details.step > summary`는 네이티브 접기 제어라 키보드 접근과 펼침 상태 전달을 브라우저에 맡김. CSS는 기존 `summary:focus-visible` 윤곽을 제거하지 않음.
- `details.step summary` 최소 높이: 데스크톱 44px, 모바일 66px. 모바일 CSS에서 `display: list-item`을 명시해 공개/접힘 표식을 복원함.
- 1280px·390px 라이트·다크 스크린샷 8장을 수동 확인함. 새 요약과 단계 목록의 가로 넘침은 각각 `1280/1280`, `390/390`으로 없음. `width: max-content` 신규 사용 없음.
- `/verify`의 시작·승인·중단 폼, CSRF 필드, 승인 프롬프트, 시장 링크는 변경하지 않았고 고정 렌더 fixture 시험으로 확인함.
- 시각 검토에서 잘린 텍스트·중복 제어·버튼 추가를 발견하지 못함. `/verify`의 1800px 목표는 머리말 `<pre>` 보존 때문에 미달이며, 내용 은닉이나 압축으로 해소하지 않음.
