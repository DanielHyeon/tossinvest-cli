# a173 측정 하네스

격리 콘솔(빈 설정 폴더·자격증명 0·포트 38321, 운영 37085 금지)에서 GET 만 한다. 버튼을 누르지 않는다.

```
node shoot.cjs <playwright-모듈-경로> http://127.0.0.1:38321 <세션토큰> <출력폴더> <metrics.json>
python3 outline.py <출력폴더> count   # 상자·식별자 수(JSON). 인자 outline 이면 제목·폼 개요
```
세션 토큰은 콘솔 기동 로그의 `?session=` 값. 스크린샷은 `<출력폴더>/{light,dark}-{d,m}-<화면>.png`, HTML 은 light 데스크톱만.
수치 JSON(숫자만)은 `baseline-<utc>.json` / `after-<utc>.json` 으로 이 폴더에 커밋한다.
