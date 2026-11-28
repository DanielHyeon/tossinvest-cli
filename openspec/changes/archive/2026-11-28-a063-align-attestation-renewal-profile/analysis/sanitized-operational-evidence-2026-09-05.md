# Sanitized operational evidence — 2026-09-05

## Collection boundaries

- Collection was read-only: user-systemd inspection, file metadata, one attestation expiry field, and the local-only JSON form of `soak status`.
- No unit was installed, reloaded, enabled, started, stopped, or edited. No survey was started and no trading control or order command was invoked.
- Record and attestation contents, account values, credentials, and raw service output were not retained here.

## Current machine state

| Item | Sanitized observation |
|---|---|
| attestation service | loaded; inactive/dead after its last run; systemd result is `success` with exit status 0 |
| attestation timer | loaded, enabled, active/waiting; last trigger 2026-09-05 22:13:01 KST |
| config-profile soak record | present; modified 2026-09-05 22:30:09 KST |
| legacy data-profile soak record | present; last modified 2026-07-30 23:27:26 KST |
| config-profile attestation | present; modified 2026-09-05 22:13:01 KST; expiry 2026-10-05T13:13:01Z |

The installed unit includes the explicit config profile. Its command appends a message after a failed `soak attest` using shell `||`, so systemd success cannot prove that renewal itself succeeded. This is the concrete failure-masking evidence; the unit is not repository-backed at HEAD.

## Local qualification summary

The masked local JSON status reported ready. Its current streak was 21 days against the required three days; the active window began 2026-08-16T00:00:00Z, its latest cycle was 2026-09-05T13:29:06Z, it recorded 47 token refreshes in the active window, and it reported zero active-window completeness failures.

The 21 qualifying dates in that active window were 2026-08-16 through 2026-09-05 inclusive. The survey and timer predated this analysis; this analysis did not activate, restart, or otherwise operate either process.

## Evidence limits

- The proposal's 2026-08-29 deadline is historical. It is not the current attestation expiry.
- A current attestation's expiry and current summary do not retain the result or reasons of the last refused timer attempt.
- Current `Summary.Evaluate` reasons may contain record-derived detail. They must not be persisted wholesale or rendered as a service-error transcript.
