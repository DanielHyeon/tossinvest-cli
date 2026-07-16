# Authenticated Sanitized Fixtures

Response shapes for authenticated endpoints. Every value here is synthetic — identifiers are
redacted and amounts, quantities, and product codes are generated.

Refresh them with both steps. `fetch_auth_fixtures.py` only redacts identifier keys
(`accountNo`, `name`, …); `scrub_fixture_values.py` replaces the remaining values so the
committed fixture is dummy data end to end:

```bash
python3 tools/fetch_auth_fixtures.py output/playwright/tossinvest-account-state.json fixtures/responses/auth-sanitized
python3 tools/scrub_fixture_values.py fixtures/responses/auth-sanitized
```

Rules:

- never commit raw storage state
- never commit raw account numbers, cookies, or tokens
- commit synthetic amounts, quantities, and product codes only — run the scrub step
- inspect diffs before commit when fixture values change

These double as a demo dataset: point `TOSSCTL_MOCK_DIR` here and account responses are served
from these files while everything else still hits the live API.

```bash
TOSSCTL_MOCK_DIR=fixtures/responses/auth-sanitized tossctl portfolio positions
```
