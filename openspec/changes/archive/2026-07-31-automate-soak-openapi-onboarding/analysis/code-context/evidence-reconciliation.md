# Evidence reconciliation

## Agreements

- Current HEAD and CodeGraph agree that no browser Open API credential setup
  route exists.
- Both agree that soak restart currently reports detached spawn success without
  checking the credential prerequisite.
- Both agree that the official client already refreshes expired access tokens.
- Both agree that environment credentials outrank the credential file.

## Reconciled risks

- Handler-local body limits are too late because `mutating` parses the form
  first. The new route must place the limiter outside `mutating`.
- The normal token file cannot validate replacement credentials because it may
  contain a valid token for the old pair. Submission validation must use a
  unique temporary cache.
- Console and setup clicks can overlap. One console mutex must serialize the
  credential-to-restart transaction.

No unresolved conflict remains. Production editing may proceed after Function
Logic Maps, Branch Test Maps, risk reports, and the Pre-Edit declaration are
complete.
