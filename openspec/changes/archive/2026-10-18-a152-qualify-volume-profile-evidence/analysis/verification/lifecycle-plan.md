# Lifecycle preparation

After independent and gstack implementation reviews and Manager pre-gate verification, freeze source, tasks and evidence; execute hard SDD sync/check and the separate Python test suite, then the aggregate gate. Preserve actual logs and exit codes. The aggregate gate does not discover the Python suite. Gate success is not implied by this plan.

After actual gate success and Manager approval, use formal `openspec archive` with main-spec synchronization. The delta modifies `candidate-discovery`; move Story `STORY-TOS-a152` paths to the archive location and regenerate/check PM trackers. The inspected active a153–a156 documents do not contain direct references to the a152 change path. Recheck direct references and origin/main immediately before lifecycle; do not force-push, alter the original checkout or overwrite existing archives/evidence. Commit only reviewed a152 artifacts and associated lifecycle changes.
