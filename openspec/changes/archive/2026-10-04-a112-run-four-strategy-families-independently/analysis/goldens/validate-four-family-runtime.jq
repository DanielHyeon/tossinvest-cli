def expected_descriptors:
  [
    {"market":"KR","family":"CONTINUATION","lane_id":"kr_short_flow_continuation_v1","lane_version":"v1","horizon":"SHORT","desired":"OFF","effective":"OFF","runtime":"UNOBSERVED"},
    {"market":"US","family":"CONTINUATION","lane_id":"us_short_participation_continuation_v1","lane_version":"v1","horizon":"SHORT","desired":"OFF","effective":"OFF","runtime":"UNOBSERVED"},
    {"market":"KR","family":"REVERSAL","lane_id":"kr_short_absorption_reversal_v1","lane_version":"v1","horizon":"SHORT","desired":"OFF","effective":"OFF","runtime":"UNOBSERVED"},
    {"market":"US","family":"REVERSAL","lane_id":"us_short_dislocation_reversal_v1","lane_version":"v1","horizon":"SHORT","desired":"OFF","effective":"OFF","runtime":"UNOBSERVED"},
    {"market":"KR","family":"WEEKLY_VALUE","lane_id":"kr_weekly_disclosure_value_v1","lane_version":"v1","horizon":"WEEKLY","desired":"OFF","effective":"OFF","runtime":"UNOBSERVED"},
    {"market":"US","family":"WEEKLY_VALUE","lane_id":"us_weekly_disclosure_value_v1","lane_version":"v1","horizon":"WEEKLY","desired":"OFF","effective":"OFF","runtime":"UNOBSERVED"},
    {"market":"KR","family":"BREAKOUT_RETEST","lane_id":"kr_short_breakout_retest_v1","lane_version":"v1","horizon":"SHORT","desired":"OFF","effective":"OFF","runtime":"UNOBSERVED"},
    {"market":"US","family":"BREAKOUT_RETEST","lane_id":"us_short_breakout_retest_v1","lane_version":"v1","horizon":"SHORT","desired":"OFF","effective":"OFF","runtime":"UNOBSERVED"}
  ];

.schema_version == 1 and
.change_id == "a112-run-four-strategy-families-independently" and
.frozen_base == "016da6245feb60e13971388be386c2c2041469a8" and
.families == ["CONTINUATION","REVERSAL","WEEKLY_VALUE","BREAKOUT_RETEST"] and
.markets == ["KR","US"] and
.descriptors == expected_descriptors and
.worker_key_fields == ["market","family","lane_id","lane_version"] and
.coordinator_key_fields == ["market"] and
.worker_count == 8 and .coordinator_count == 2 and
.shared_authorities.owner_key_fields == ["account","market","symbol","position_generation"] and
.shared_authorities.family_or_horizon_in_owner_key == false and
.shared_authorities.account_scoped_guardian_count == 1 and
.shared_authorities.official_execution_gateway_count == 1 and
.states.allowed_transitions == [
  {"from":"DISCOVERED","to":"RANGE_LOCKED"},{"from":"RANGE_LOCKED","to":"BREAKOUT_CLOSED"},
  {"from":"BREAKOUT_CLOSED","to":"RETEST_WAIT"},{"from":"RETEST_WAIT","to":"RECLAIMED"},
  {"from":"RECLAIMED","to":"ARMED"},{"from":"ARMED","to":"PROPOSED"},
  {"from":"DISCOVERED","to":"INVALIDATED"},{"from":"RANGE_LOCKED","to":"INVALIDATED"},
  {"from":"BREAKOUT_CLOSED","to":"INVALIDATED"},{"from":"RETEST_WAIT","to":"INVALIDATED"},
  {"from":"RECLAIMED","to":"INVALIDATED"},{"from":"ARMED","to":"INVALIDATED"},
  {"from":"RETEST_WAIT","to":"TIMED_OUT"},{"from":"PROPOSED","to":"CONSUMED"}
] and
.states.terminal_non_resurrection == true and
.refusal_enums.arbitration == ["ARBITRATION_UNCALIBRATED","ARBITRATION_TIE","ARBITRATION_MULTIPLE_OWNER","ARBITRATION_STALE_OWNER","ARBITRATION_STALE_ENVELOPE","ARBITRATION_SEAL_MISMATCH"] and
.refusal_enums.quote_fx_sizing == ["QUOTE_STALE","SPREAD_TOO_WIDE","ENTRY_DRIFT_EXCEEDED","FX_MISSING","FX_STALE","FX_CURRENCY_MISMATCH","FX_INVALID_RATE","SIZING_OVERFLOW","NON_PROTECTIVE_STOP","NON_PROTECTIVE_TARGET","ZERO_QUANTITY"] and
.refusal_enums.scheduler == ["BUDGET_DEFERRED"] and
.refusal_enums.diagnostic_not_refusal == ["CORRECTION_AFTER_PROPOSAL"] and
.arbitration.score_ppm == {"integer":true,"minimum":0,"maximum":1000000} and
.arbitration.singleton_without_calibration_refusal == "ARBITRATION_UNCALIBRATED" and
.arbitration.singleton_without_calibration_dispatch_handoff == false and
.queue.dedup_key_fields == ["account","market","symbol","position_generation","family","lane_id","lane_version","snapshot_digest"] and
.queue.market_wide_single_proposal_assumption_forbidden == true and
.operator_compatibility.additive_children == ["lanes[8] fixed deterministic order","coordinators[2] fixed deterministic order"]
