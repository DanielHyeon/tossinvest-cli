def expected_setup_identity:
  {
    "algorithm":"SHA-256",
    "domain":"tossos.breakout.setup.v1",
    "preimage_encoding":"UTF-8",
    "canonical_preimage_fields":["market","symbol","session_id","calendar_version","opening_range_first_bar_id","opening_range_last_bar_id","lane_id","lane_version","config_digest"],
    "separator":{"byte":0,"notation":"NUL (0x00)"},
    "trailing_delimiter":false,
    "field_value_rejections":["EMPTY","LEADING_OR_TRAILING_WHITESPACE","NUL"],
    "forbidden_pre_hash_operations":["TRIM","CASE_FOLD","UNICODE_NORMALIZATION","JSON_SERIALIZATION"],
    "output_format":{"prefix":"sha256:","hex_case":"lowercase","hex_length":64},
    "known_vector":{
      "market":"KR",
      "symbol":"005930",
      "session_id":"KRX:2026-08-18",
      "calendar_version":"krx-calendar-v1",
      "opening_range_first_bar_id":"bar-20260818-0900",
      "opening_range_last_bar_id":"bar-20260818-0914",
      "lane_id":"kr_short_breakout_retest_v1",
      "lane_version":"v1",
      "config_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "expected_setup_id":"sha256:d2d5e2da006b841e45a1d2991624c5316b520a3900bb780653c79f455eeef04c"
    },
    "bar_revision_location":"snapshot/evidence digest only; never setup_id",
    "proposal_identity":"idempotent per setup/snapshot/config",
    "first_leg":"at most one shared-admission authority; no scale-in authority"
  };

.schema_version == 1 and
.change_id == "a112-run-four-strategy-families-independently" and
.frozen_base == "016da6245feb60e13971388be386c2c2041469a8" and
.setup_identity == expected_setup_identity
