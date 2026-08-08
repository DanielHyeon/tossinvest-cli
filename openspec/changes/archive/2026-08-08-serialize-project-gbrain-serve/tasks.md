## 1. Reproduction and contract

- [x] 1.1 Capture the change base, validate the OpenSpec artifacts, and record proposal-freeze review
- [x] 1.2 Add RED subprocess tests for duplicate singleton rejection, crash release, and legacy live-owner detection
- [x] 1.3 Add RED SDD sync tests proving GBrain busy is advisory while other GBrain failures remain incomplete

## 2. Project GBrain ownership

- [x] 2.1 Implement inheritable kernel flock acquisition and minimal owner metadata in `gbrain_project.py`
- [x] 2.2 Detect live pre-wrapper PGLite owners without deleting their lock and return exit 75 busy diagnostics
- [x] 2.3 Route all GBrain SDD sync commands through the project wrapper and isolate verified busy contention

## 3. Recovery and documentation

- [x] 3.1 Document singleton ownership, busy behavior, and safe duplicate recovery in `docs/WORKFLOW.md`
- [x] 3.2 Revalidate PID, command, and `GBRAIN_HOME`, then terminate only the current non-owner duplicate GBrain child
- [x] 3.3 Verify focused tests, live busy latency, `make sdd-sync`, `make sdd-check`, strict OpenSpec validation, and project gate

## 4. Review and retention

- [x] 4.1 Record Function Logic Map: not-applicable for Python SDD tooling and run an independent diff/test review
- [x] 4.2 Retain the verified duplicate-process/PGLite ownership lesson in episodic memory
