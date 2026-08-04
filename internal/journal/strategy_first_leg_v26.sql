CREATE TABLE strategy_first_leg_bindings (
 decision_id TEXT PRIMARY KEY REFERENCES risk_bucket_final_decisions(decision_id),
 aggregate_reservation_id TEXT NOT NULL UNIQUE REFERENCES risk_reservations(id),
 entry_decision_identity TEXT NOT NULL UNIQUE REFERENCES strategy_decision_lineage(entry_decision_identity),
 attempt_id TEXT NOT NULL UNIQUE REFERENCES strategy_attempt_lineage(attempt_id),
 campaign_id TEXT NOT NULL UNIQUE REFERENCES position_campaigns(id),
 leg_sequence INTEGER NOT NULL CHECK(leg_sequence=1),
 leg_plan_id TEXT NOT NULL,
 account_ref TEXT NOT NULL,
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 symbol TEXT NOT NULL,
 candidate_id TEXT NOT NULL,
 evidence_digest TEXT NOT NULL,
 lane_id TEXT NOT NULL,
 lane_version TEXT NOT NULL,
 router_id TEXT NOT NULL CHECK(length(router_id) BETWEEN 1 AND 256),
 router_version TEXT NOT NULL CHECK(length(router_version) BETWEEN 1 AND 256),
 prospective_token TEXT NOT NULL UNIQUE CHECK(length(prospective_token)=64 AND prospective_token NOT GLOB '*[^0-9a-f]*'),
 q_final INTEGER NOT NULL CHECK(q_final>0),
 request_digest TEXT NOT NULL UNIQUE CHECK(length(request_digest)=64 AND request_digest NOT GLOB '*[^0-9a-f]*'),
 record_digest TEXT NOT NULL UNIQUE CHECK(length(record_digest)=64 AND record_digest NOT GLOB '*[^0-9a-f]*'),
 created_at TEXT NOT NULL,
 FOREIGN KEY(campaign_id,leg_sequence) REFERENCES campaign_legs(campaign_id,sequence)
) STRICT;

CREATE TRIGGER strategy_first_leg_binding_insert_guard
BEFORE INSERT ON strategy_first_leg_bindings
WHEN NOT EXISTS (
 SELECT 1
 FROM risk_bucket_final_decisions q
 JOIN decisions risk_intent ON risk_intent.id=q.decision_id
 JOIN risk_reservations aggregate_hold
   ON aggregate_hold.id=q.existing_reservation_id
  AND aggregate_hold.decision_id=q.decision_id
 JOIN strategy_attempt_lineage attempt
   ON attempt.attempt_id=NEW.attempt_id
  AND attempt.risk_intent_id=q.decision_id
  AND attempt.guardian_decision_id=q.decision_id
 JOIN strategy_decision_lineage strategy
   ON strategy.entry_decision_identity=attempt.entry_decision_identity
 JOIN position_campaigns campaign
   ON campaign.id=NEW.campaign_id
  AND campaign.decision_id=q.decision_id
 JOIN position_campaign_claims claim
   ON claim.campaign_id=campaign.id
 JOIN campaign_legs leg
   ON leg.campaign_id=campaign.id
  AND leg.sequence=NEW.leg_sequence
 JOIN risk_bucket_owners owner
   ON owner.account_ref=q.account_ref
  AND owner.market=q.market
  AND owner.symbol=q.symbol
  AND owner.prospective_generation=q.owner_prospective_generation
  AND owner.released_at IS NULL
 WHERE q.decision_id=NEW.decision_id
   AND q.existing_reservation_id=NEW.aggregate_reservation_id
   AND q.account_ref=NEW.account_ref
   AND q.market=NEW.market
   AND q.symbol=NEW.symbol
   AND q.owner_lane_id=NEW.lane_id
   AND q.owner_campaign_id=NEW.campaign_id
   AND q.owner_prospective_generation=NEW.prospective_token
   AND q.q_final=NEW.q_final
   AND risk_intent.account_ref=NEW.account_ref
   AND risk_intent.safety_class='EXPOSURE_RAISING'
   AND risk_intent.preimage_kind='RISK_INTENT'
   AND aggregate_hold.account_ref=NEW.account_ref
   AND aggregate_hold.state='HELD'
   AND attempt.account_ref=NEW.account_ref
   AND attempt.entry_decision_identity=NEW.entry_decision_identity
   AND attempt.state='PLANNED'
   AND attempt.activation_manifest_digest=strategy.activation_manifest_digest
   AND attempt.client_order_id=risk_intent.client_order_id
   AND upper(strategy.market)=NEW.market
   AND strategy.symbol=NEW.symbol
   AND strategy.candidate_life_id=NEW.candidate_id
   AND strategy.evidence_digest=NEW.evidence_digest
   AND strategy.lane_id=NEW.lane_id
   AND strategy.lane_version=NEW.lane_version
   AND strategy.quantity=CAST(NEW.q_final AS TEXT)
   AND campaign.account_ref=NEW.account_ref
   AND lower(campaign.market)=lower(NEW.market)
   AND campaign.symbol=NEW.symbol
   AND campaign.lane_id=NEW.lane_id
   AND campaign.lane_version=NEW.lane_version
   AND campaign.evidence_digest=NEW.evidence_digest
   AND campaign.prospective_token=NEW.prospective_token
   AND campaign.state='PLANNED'
   AND campaign.entry_blocked=0
   AND claim.account_ref=NEW.account_ref
   AND lower(claim.market)=lower(NEW.market)
   AND claim.symbol=NEW.symbol
   AND claim.position_generation=campaign.expected_position_generation
   AND claim.position_version=campaign.expected_position_version
   AND claim.prospective_token=NEW.prospective_token
   AND leg.plan_id=NEW.leg_plan_id
   AND leg.requested_quantity=CAST(NEW.q_final AS TEXT)
   AND leg.residual_quantity=CAST(NEW.q_final AS TEXT)
   AND leg.filled_quantity='0'
   AND leg.state='PLANNED'
   AND owner.lane_id=NEW.lane_id
   AND owner.campaign_id=NEW.campaign_id
   AND (SELECT count(*) FROM risk_bucket_reservations monetary
        WHERE monetary.decision_id=q.decision_id
          AND monetary.existing_reservation_id=q.existing_reservation_id
          AND monetary.account_ref=q.account_ref
          AND monetary.market=q.market
          AND monetary.symbol=q.symbol
          AND monetary.owner_prospective_generation=q.owner_prospective_generation
          AND monetary.state='HELD'
          AND monetary.held_minor=monetary.reserved_minor)=5
)
BEGIN SELECT RAISE(ABORT,'first-leg binding requires exact atomic authority'); END;

CREATE TRIGGER strategy_first_leg_bindings_no_update
BEFORE UPDATE ON strategy_first_leg_bindings
BEGIN SELECT RAISE(ABORT,'first-leg bindings are immutable'); END;

CREATE TRIGGER strategy_first_leg_bindings_no_delete
BEFORE DELETE ON strategy_first_leg_bindings
BEGIN SELECT RAISE(ABORT,'first-leg bindings are immutable'); END;

CREATE TRIGGER strategy_dispatch_lease_first_leg_binding
BEFORE INSERT ON strategy_dispatch_leases
WHEN NOT EXISTS (
 SELECT 1 FROM strategy_first_leg_bindings binding
 JOIN strategy_attempt_lineage attempt ON attempt.attempt_id=binding.attempt_id
 JOIN strategy_decision_lineage strategy ON strategy.entry_decision_identity=binding.entry_decision_identity
 JOIN decisions risk_intent ON risk_intent.id=binding.decision_id
 JOIN position_campaigns campaign ON campaign.id=binding.campaign_id
 JOIN position_campaign_claims claim ON claim.campaign_id=binding.campaign_id
 JOIN campaign_legs leg ON leg.campaign_id=binding.campaign_id AND leg.sequence=binding.leg_sequence
 JOIN risk_bucket_owners owner
   ON owner.account_ref=binding.account_ref
  AND owner.market=binding.market
  AND owner.symbol=binding.symbol
  AND owner.prospective_generation=binding.prospective_token
  AND owner.released_at IS NULL
 WHERE binding.decision_id=NEW.guardian_decision_id
   AND binding.aggregate_reservation_id=NEW.risk_reservation_id
   AND binding.account_ref=NEW.account_ref
   AND binding.market=NEW.market
   AND binding.symbol=NEW.symbol
   AND binding.candidate_id=NEW.candidate_id
   AND binding.evidence_digest=NEW.evidence_digest
   AND binding.lane_id=NEW.lane_id
   AND binding.lane_version=NEW.lane_version
   AND binding.router_id=NEW.router_id
   AND binding.router_version=NEW.router_version
   AND binding.campaign_id=NEW.campaign_id
   AND binding.leg_plan_id=NEW.leg_id
   AND attempt.risk_intent_id=binding.decision_id
   AND attempt.guardian_decision_id=binding.decision_id
   AND attempt.account_ref=binding.account_ref
   AND attempt.entry_decision_identity=binding.entry_decision_identity
   AND attempt.state='PLANNED'
   AND attempt.activation_manifest_digest=strategy.activation_manifest_digest
   AND attempt.client_order_id=risk_intent.client_order_id
   AND NEW.operation_id=attempt.client_order_id
   AND risk_intent.account_ref=binding.account_ref
   AND risk_intent.safety_class='EXPOSURE_RAISING'
   AND risk_intent.preimage_kind='RISK_INTENT'
   AND upper(strategy.market)=binding.market
   AND strategy.symbol=binding.symbol
   AND strategy.candidate_life_id=binding.candidate_id
   AND strategy.evidence_digest=binding.evidence_digest
   AND strategy.lane_id=binding.lane_id
   AND strategy.lane_version=binding.lane_version
   AND strategy.quantity=CAST(binding.q_final AS TEXT)
   AND campaign.decision_id=binding.decision_id
   AND campaign.account_ref=binding.account_ref
   AND lower(campaign.market)=lower(binding.market)
   AND campaign.symbol=binding.symbol
   AND campaign.lane_id=binding.lane_id
   AND campaign.lane_version=binding.lane_version
   AND campaign.evidence_digest=binding.evidence_digest
   AND campaign.prospective_token=binding.prospective_token
   AND campaign.state='PLANNED'
   AND campaign.entry_blocked=0
   AND claim.account_ref=binding.account_ref
   AND lower(claim.market)=lower(binding.market)
   AND claim.symbol=binding.symbol
   AND claim.position_generation=campaign.expected_position_generation
   AND claim.position_version=campaign.expected_position_version
   AND claim.prospective_token=binding.prospective_token
   AND leg.plan_id=binding.leg_plan_id
   AND leg.requested_quantity=CAST(binding.q_final AS TEXT)
   AND leg.residual_quantity=CAST(binding.q_final AS TEXT)
   AND leg.filled_quantity='0'
   AND leg.state='PLANNED'
   AND owner.lane_id=binding.lane_id
   AND owner.campaign_id=binding.campaign_id
)
BEGIN SELECT RAISE(ABORT,'strategy dispatch lease requires exact first-leg binding'); END;
