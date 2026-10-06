# TypeSafe System One API 계약 동결 (a128 task 1.1)

- 읽은 순간: 2026-10-07 (KST), `curl -sL` 로 받은 Markdown 원문.
- 읽은 페이지(모집단): `https://docs.typesafe.ai/llms.txt`(인덱스), `api.md`, `primitives/noul.md`,
  `models.md`, `concepts/state.md`, `sdk/python/api/constants.md`, `sdk/python/api/retries.md`.
  그 밖의 페이지는 읽지 않았다 — 아래 "문서 미기재"는 **이 일곱 페이지 범위에서** 미기재라는 뜻이다.
- `noul.md` 는 JSX(`TypesafeExample` 컴포넌트·압축 함수)가 섞여 있어 산문과 코드 블록만 인용한다.
- 인용은 원문 그대로(영문). 해석은 각 항목의 "프로브 적용" 줄에만 쓴다.

## 1. 엔드포인트·인증

`api.md` "Evaluation endpoint":

> ```http
> POST https://api.typesafe.ai/v1/systemone
> Authorization: Bearer <API_KEY>
> Content-Type: application/json
> ```

`sdk/python/api/constants.md`:

> `API_KEY_ENV = 'TYPESAFE_API_KEY'` — "Environment variable for the API key."
> `DEFAULT_BASE_URL = 'https://api.typesafe.ai'` — "Default API base URL."
> `DEFAULT_TIMEOUT = 10.0` — "Default timeout in seconds for each HTTP operation."

프로브 적용: `POST {base}/v1/systemone`, base 기본 `https://api.typesafe.ai`(시험은 httptest 로 교체),
헤더 `Authorization: Bearer $TYPESAFE_API_KEY` · `Content-Type: application/json`, 시도당 타임아웃 10s.

## 2. 요청 본문

`api.md` "Request body":

> `state` (string | object | array, required) — "The content to evaluate. A plain string for text,
> or structured data (object/array) for things like chat logs, records, or the current state of your
> application."
>
> `model` (string, required) — "The model that handles the request. Use `"jev-latest"`, TypeSafe's
> flagship model."
>
> `questions` (map<string, Question>, required) — "A map of typed Question objects. You choose each
> key; answers come back under the same keys."
>
> question id — "A key you choose. The matching Answer is returned under this same id. The key is not
> sent to the underlying model and is not used in inference."

### Noul 질문

`api.md` "Noul":

> "A yes/no question. Returns the probability the answer is yes."
>
> `type` ("noul", required)
>
> `instructions` (string | object | array, required) — "The yes/no question to evaluate."
>
> `criteria` (object, optional) — "Optional descriptions of what a yes and a no mean."
> `true` — "What a yes (value near 1) means." / `false` — "What a no (value near 0) means."

예시 원문(`api.md`):

> ```json
> {
>   "state": "Help! My payouts have been failing for 3 days.",
>   "model": "jev-latest",
>   "questions": {
>     "is_urgent": {
>       "type": "noul",
>       "instructions": "Does this convey urgency?",
>       "criteria": {
>         "true": "Explicitly time-sensitive",
>         "false": "No urgency expressed"
>       }
>     }
>   }
> }
> ```

`noul.md` "Writing a Noul question":

> "Ask one yes/no question per Noul. If a question has two conditions, … the model has to judge both
> at once and the value means less."
>
> "Phrase the question so that a high value means yes."

`noul.md` "Good practice: ask more than one question per call":

> "Questions are evaluated in parallel, so adding Nouls barely changes the response time."

`concepts/state.md`:

> "Each request evaluates one state against one or more questions. All questions see the same state
> and are evaluated independently."
>
> "Use an object for most requests so each part of the state has a descriptive name and its
> relationships remain clear."

프로브 적용: J1·J2 를 **한 요청**의 두 question id(`j1_up_within_horizon`, `j2_hold_off_red_flag`)로
보낸다(같은 state, 독립 평가). state 는 JSON object. 질문은 영어(아래 4 절 근거).

## 3. 응답 본문 — p 의 경로

`api.md` "Response body":

> `model` (string, required) — "The model that performed the evaluation."
>
> `answers` (map<string, Answer>, required) — "One Answer per question, keyed by the same ids you used
> in questions."
>
> `usage` (object, required) — "Token usage for the request." 속성 `input_tokens`(integer),
> `output_tokens`(integer).

`api.md` "Noul answer":

> `type` ("noul", required)
>
> `noul` (number, required) — "The yes/no answer on a scale from 0 (no) to 1 (yes)."

> ```json
> {
>   "model": "jev-1.13.0",
>   "answers": {
>     "is_urgent": {
>       "type": "noul",
>       "noul": 0.95
>     }
>   },
>   "usage": { "input_tokens": 307, "output_tokens": 20 }
> }
> ```

`noul.md`:

> "There is no separate `confidence` value for a Noul, unlike a Choice or a Score."
>
> "A Noul value runs from 0 to 1, but it's not a scale of the thing you asked about. It is the
> probability that the answer is yes."

프로브 적용: **p = `answers.<question id>.noul`**. `answers.<id>.type != "noul"`, `noul` 부재,
0~1 밖, 또는 id 부재면 계약 위반으로 그 판단을 버리고 원장에 결손(`gap`)으로 적는다(고쳐 읽지 않음).
응답 `model`(답한 버전 ID)을 판단 행에 그대로 기록한다.

## 4. 모델·버전·언어

`models.md`:

> `jev-latest` → `jev-1.13.0` — "The most recent stable, official release."
>
> "An alias moves when a new release ships, so the answers behind it can change without a change on
> your side. The response's `model` field reports the versioned ID that answered, so you can log which
> model produced each result. If you have tuned confidence thresholds against a specific version, pin
> that version's ID instead of the alias and move to the new one on your own schedule."
>
> "Versioned IDs such as `jev-1.13.0` are accepted by the `model` field whether or not they appear in
> the list."
>
> "English is the primary training language and where accuracy is currently best. Other languages,
> including CJK scripts, are handled but not equally well"
>
> Context length: "64k tokens per request; 32k tokens for `state` plus the longest question"
>
> Price: "\$42 / \$0.042" (per Btok / per Mtok), "Charged per input token. Output tokens are free."

프로브 적용: 기본 `--model jev-1.13.0`(버전 핀 — 측정 중 alias 이동으로 모집단이 바뀌지 않게).
질문·state 키는 영어. 종목명(한글)은 값으로만 들어간다.

## 5. 오류·재시도

`api.md` "Errors":

> | `401 Unauthorized` | Missing or invalid API key. Check the `Authorization` header. |
> | `422 Unprocessable Entity` | The request body failed validation … The body details the offending field. |
> | `429 Too Many Requests` | You have exceeded your rate limit. Back off and retry after a short delay. |
> | `529 Overloaded` | TypeSafe is temporarily overloaded. Retry after a short delay. |
>
> "When you receive a `429 Too Many Requests` or `529 Overloaded` response, retry the request with
> exponential backoff instead of retrying immediately."

`models.md`:

> "Rate limits: … 100K tokens per second / 80 requests per second" · "A request over either limit
> returns `429 Too Many Requests`. Our client SDKs retry with backoff by default and honor the
> `retry-after` header when the response carries one."

`sdk/python/api/retries.md`(SDK 기본값 — HTTP 계약이 아니라 참고):

> `RetryPolicy(max_retries: int = 2, backoff_initial: float = 0.5, backoff_max: float = 5.0,
> backoff_jitter: float = 0.25, http_statuses: set[int] = {408, 429, *range(500, 600)},
> respect_retry_after: bool = True, …)`
>
> backoff_initial — "First backoff delay in seconds, doubled each attempt up to `backoff_max`"
>
> respect_retry_after — "Whether to honor `Retry-After` and `retry-after-ms` response headers."

프로브 적용: 재시도 2회(총 3 시도), 대상 = 408·429·5xx(529 포함)·전송 오류, 지연 0.5s 에서 2배씩
최대 5s, `Retry-After`(초) 가 있으면 그것을 쓴다(상한 30s). 401·422 등 그 밖의 4xx 는 재시도하지 않는다.

## 6. 문서 미기재 (위 일곱 페이지 범위)

지어내지 않고 미기재로 남긴 것. 프로브는 이것들에 의존하지 않는다.

1. **오류 응답 JSON 본문의 필드 구조** — "a JSON body describing what went wrong" 이라고만 한다.
   프로브는 본문을 해석하지 않고 상태 코드 + 본문 앞 200 바이트만 오류 문구에 싣는다.
2. **`Retry-After` 헤더의 형식**(초 정수 / HTTP-date) — HTTP API 페이지에는 헤더 이름조차 없고
   SDK 페이지가 이름(`Retry-After`, `retry-after-ms`)만 말한다. 프로브는 초 정수만 해석하고 그 밖은 무시.
3. **요청 크기 상한(바이트)·question 개수 상한** — 토큰 예산(64k/32k)만 있다. state 는 수 KB 수준이라 무관.
4. **Noul `noul` 값의 소수 자릿수·정밀도** — "number" 로만 기재. 프로브는 float64 그대로 기록.
5. **멱등성 키·요청 ID 헤더** — 기재 없음. 판단 호출은 side effect 없는 평가라 재시도 중복은 비용만 늘린다.
6. **동일 입력의 결정성(같은 state·질문에 같은 p 인가)** — 이 범위에서 기재 없음
   (`cookbooks/consistency_noul_cookbook.md` 가 다룬다고 인덱스에 있으나 읽지 않았다). 원장은 state digest 와
   질문 digest 를 남겨 사후에 잴 수 있게 한다.
7. **응답 `usage` 이외의 과금 메타(요청당 비용)** — 가격표(per-token)만 있다. 비용은 `usage.input_tokens`
   로 사후 계산하도록 판단 행에 기록한다.
