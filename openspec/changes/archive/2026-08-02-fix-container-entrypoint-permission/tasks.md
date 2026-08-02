## 1. Contract and implementation

- [x] 1.1 Capture exit 126 logs and identify source/image mode mismatch.
- [x] 1.2 Pin entrypoint mode to 0755 in the Dockerfile and document the invariant.
- [x] 1.3 Add a static regression test for the executable copy instruction.

## 2. Verification and release

- [x] 2.1 Run focused tests, build the image without relying on stale layers, and inspect the image entrypoint mode.
- [x] 2.2 Run SDD/OpenSpec gates and independent review.
- [x] 2.3 Commit, push remote main, redeploy Compose, and pass HTTPS HTTP/2 canaries.
