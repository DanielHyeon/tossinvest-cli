# Review: a050-add-strategy-optimization

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability, UI/UX

## Findings and decisions

1. StockOS 최신 lane-console의 화면 단위 navigation, changed-key-only save와 effective refetch를 참고한다. 구형 slider matrix는 복제하지 않는다.
2. 모든 조작은 stable preset/radio/select/chip/toggle/discrete-step/current-row action이다. text/number/textarea/contenteditable/free symbol/typed phrase/free reason은 0개다.
3. a041 `internal/settingmeta`를 조합하며 owner descriptor가 유일한 label/default/help source다. a050은 category/lifecycle만 소유한다.
4. console/httpapi는 journal read-only이고 durable commander를 공유한다. engine만 trading journal state를 쓴다.
5. high-risk apply/rollback은 desired를 보존할 수 있으나 manifest가 바뀌면 effective entry는 OFF다. rollback은 current constraints를 검증한 새 version이다.
6. evidence 부족은 자동 추천을 막지만 검증된 보수적 server preset의 human selection 자체를 막지 않는다.

## Verification evidence

- OpenSpec strict validation: pass.
- StockOS reference inspected: latest lane-console shell/tabs/human-only control; old full-page slider matrix rejected.

## Verdict

a049 뒤 구현을 승인한다. input-free static/render tests, CAS/fault tests, independent UI/security review가 gate 조건이다. LIVE/activation authority는 범위 밖이다.
