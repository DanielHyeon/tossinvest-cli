# SDD sync timeout repair evidence

Baseline `sync` Python AST: `/tmp/a145-sync-before-ast.json`; Go-only extractor is not applicable to Python. Original `sync`: successful hard CodeGraph enters success set; advisory stats timeout raised before final state recording.

Branch map: hard success + stats timeout => advisory failure, record hard success; hard failure + stats timeout => both failures, empty success set, no hard stamp. Normal stats branches unchanged. Negative fixture in `test_context_stats_timeout_preserves_only_successful_hard_index`; RED `/tmp/a145-sync-red.log`, GREEN `/tmp/a145-sync-green.log`. No index-state fabrication or gate bypass.
