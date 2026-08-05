# Branch Test Map: `Console.decoratePositionRows`

12개 분기 전부와, 분기로는 드러나지 않는 **계약** 넷(C1–C4)을 덮는다. a081이
바꾸는 것은 B1·B2가 읽는 값의 **출처**이므로 그 둘과 그 아래 fallback 경로
(B7·B10·B11)가 판정의 중심이다. B4–B6·B9·B12는 손대지 않으며 기존 테스트가
이미 덮고 있다는 것을 회귀로 확인한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 참: 커맨더 배선됨 — 간격 안에서 여러 번 렌더해도 엔진 읽기는 각각 1회 / 거짓: 커맨더 nil — 엔진 읽기 0회, `runtimeAttempted` false로 렌더 | 2.2, 5.2 | yes | yes |
| B2 | 참: 목록 획득 성공 — `policyByID`가 채워지고 `엔진 관리`가 표시된다 / 거짓: 실패 — 캐시가 실패를 실패로 서빙하고 행은 `관리 여부 불명`이 된다 | 5.1, 5.2, 8.5 | yes | yes |
| B3 | 상태 목록을 PositionID로 색인한다 (공백 trim). 빈 목록이면 0회 | 5.1 | no | yes |
| B4 | 참: Settings 배선됨 / 거짓: 미배선 — `Designated`·`Excluded` zero value | 기존 | no | yes |
| B5 | 참: `Load` 성공 / 거짓: 실패 — 두 목록 모두 미표시 | 기존 | no | yes |
| B6 | 행별 desired include/exclude 표시. 빈 rows면 0회 | 기존 | no | yes |
| B7 | 참: `runtimeAttempted` — lifecycle·management 투영이 돈다 / 거짓: 그 아래 전체를 건너뛴다 | 5.1, 5.2 | yes | yes |
| B8 | 행 순회 — 행마다 독립적으로 투영된다. 빈 rows면 0회 | 5.1 | no | yes |
| B9 | 참: `InJournal` 행만 lifecycle 증명을 요구한다 / 거짓: 요구하지 않는다 | 기존 | no | yes |
| B10 | 참: 목록에 없음(또는 목록 자체가 nil) → `journalKnown=false` | 5.2 | no | yes |
| B11 | 거짓 갈래(else): 목록에 있음 → status·generation·released가 캐시된 값에서 온다 | 5.1, 4.4 | no | yes |
| B12 | 참: 대사 차단이 있으면 block view를 붙인다 / 거짓: 붙이지 않는다 | 기존 | no | yes |
| C1 | 상한 — 렌더 N회(N > 허용 읽기 수)에도 `Runtime`·`List` 각각 1회 이하. 자체 무의미 통과 검사 포함 | 2.2, 2.3 | yes | yes |
| C2 | 공유 — 두 화면이 간격을 공유한다. 화면마다가 아니라 간격마다 1벌 | 2.4 | yes | yes |
| C3 | 방향 — journal보다 오래된 목록은 판정을 보류할 뿐 만들어내지 않는다 (초안의 "한 시점의 짝"은 철회, 8.3) | 8.3 | yes | yes |
| C4 | 무효화 — 성공한 **정책** mutation은 캐시를 버리고, 실패한 것은 버리지 않는다. 격리 해제는 이 캐시를 지나지 않아 대상이 아니다 (8.4) | 4.1, 4.3 | yes | yes |
| C5 | 오염 금지 — 렌더를 버린 브라우저의 취소가 공유 reading에 기록되지 않는다 | 8.1 | yes | yes |
| C6 | 동시성 — 같이 도착한 렌더 여럿이 읽기 1벌만 만든다 | 8.6 | yes | yes |
| C7 | 절반별 간격 — 대사 차단은 lifecycle 간격을 기다리지 않는다 | 8.2 | yes | yes |

C1–C3은 분기가 아니라 이 함수가 **엔진에 거는 부하와 일관성**의 계약이다.
분기표만으로는 "몇 번 불렀는가"를 판정할 수 없어 따로 세운다 — a080의 예산
테스트가 F1을 놓친 이유가 정확히 이것이었다(하네스가 이 경로를 실행하지 않아
분기 커버리지는 초록이었다).

C4는 `decoratePositionRows` 바깥(Apply·격리 해제 핸들러)에서 발생하지만 이
함수가 다음 렌더에서 보는 값을 결정하므로 같은 map에 둔다.

RED은 캐시 없는 HEAD에서 관측한 것과 변이 검증 7종에서 관측한 것을 함께 센다.
B1·B7·C1·C2는 캐시 도입 전에 렌더 수만큼 엔진에 닿는 것을 그대로 관측했고, C3~C7은
각각 stale 목록 부활·무효화 제거·요청 컨텍스트 사용·늦은 stamp·간격 통합 변이에서
관측했다. GREEN은 패키지 전체가 통과한 실행에서 확인했다.

**B2의 RED은 초안에 없었다.** 픽스처의 `PositionID`가 journal과 어긋나 lifecycle
절반이 한 번도 실행되지 않았고, 그 상태로 stale 목록 부활 변이가 전 테스트를
통과했다 (review.md 3차). 지금은 `TestAFailedLifecycleReadingIsNotMaskedByThePreviousSuccess`가
그 변이를 잡는다.
