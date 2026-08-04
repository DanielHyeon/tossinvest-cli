CREATE TABLE strategy_dispatch_owner_epochs (
 owner_epoch INTEGER PRIMARY KEY CHECK(owner_epoch>0),
 fencing_token TEXT NOT NULL UNIQUE,
 owner_instance TEXT NOT NULL,
 acquired_at TEXT NOT NULL,
 UNIQUE(owner_epoch,fencing_token)
) STRICT;
CREATE TABLE strategy_dispatch_owner_current (
 owner_key TEXT PRIMARY KEY CHECK(owner_key='CENTRAL'),
 owner_epoch INTEGER NOT NULL,
 fencing_token TEXT NOT NULL,
 owner_instance TEXT NOT NULL,
 revision INTEGER NOT NULL CHECK(revision>0),
 acquired_at TEXT NOT NULL,
 FOREIGN KEY(owner_epoch,fencing_token) REFERENCES strategy_dispatch_owner_epochs(owner_epoch,fencing_token)
) STRICT;
CREATE TABLE strategy_dispatch_market_authorities (
 authority_id TEXT PRIMARY KEY,
 account_ref TEXT NOT NULL,
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 symbol TEXT NOT NULL,
 activation_generation INTEGER NOT NULL CHECK(activation_generation>0),
 activation_digest TEXT NOT NULL,
 calendar_generation INTEGER NOT NULL CHECK(calendar_generation>0),
 protection_generation INTEGER NOT NULL CHECK(protection_generation>0),
 protection_serial TEXT NOT NULL,
 protection_digest TEXT NOT NULL,
 reconciliation_generation INTEGER NOT NULL CHECK(reconciliation_generation>0),
 risk_policy_generation INTEGER NOT NULL CHECK(risk_policy_generation>0),
 risk_policy_digest TEXT NOT NULL,
 guardian_generation INTEGER NOT NULL CHECK(guardian_generation>0),
 guardian_digest TEXT NOT NULL,
 build_digest TEXT NOT NULL,
 revision INTEGER NOT NULL CHECK(revision>0),
 record_digest TEXT NOT NULL UNIQUE,
 updated_at TEXT NOT NULL,
 UNIQUE(account_ref,market,symbol)
) STRICT;
CREATE INDEX idx_strategy_dispatch_authority_market
 ON strategy_dispatch_market_authorities(market,account_ref,symbol,revision);
CREATE TABLE strategy_dispatch_leases (
 lease_id TEXT PRIMARY KEY,
 operation_id TEXT NOT NULL UNIQUE,
 account_ref TEXT NOT NULL,
 market TEXT NOT NULL CHECK(market IN ('KR','US')),
 symbol TEXT NOT NULL,
 candidate_id TEXT NOT NULL,
 evidence_digest TEXT NOT NULL,
 router_id TEXT NOT NULL,
 router_version TEXT NOT NULL,
 lane_id TEXT NOT NULL,
 lane_version TEXT NOT NULL,
 campaign_id TEXT NOT NULL,
 leg_id TEXT NOT NULL,
 risk_reservation_id TEXT NOT NULL UNIQUE,
 guardian_decision_id TEXT NOT NULL,
 owner_epoch INTEGER NOT NULL,
 fencing_token TEXT NOT NULL,
 authority_revision INTEGER NOT NULL CHECK(authority_revision>0),
 authority_digest TEXT NOT NULL,
 issued_at TEXT NOT NULL,
 expires_at TEXT NOT NULL,
 state TEXT NOT NULL CHECK(state IN ('ISSUED','CLAIMED','SUBMITTING','SUBMITTED','AMBIGUOUS','REFUSED')),
 disposition TEXT NOT NULL CHECK(disposition IN ('RESERVED','RELEASED','TRANSFERRED','HELD')),
 revision INTEGER NOT NULL CHECK(revision>0),
 transport_started_at TEXT,
 refusal_code TEXT NOT NULL DEFAULT '',
 outcome_code TEXT NOT NULL DEFAULT '',
 broker_order_id TEXT NOT NULL DEFAULT '',
 query_digest TEXT NOT NULL DEFAULT '',
 outcome_observed_at TEXT,
 lease_digest TEXT NOT NULL UNIQUE,
 created_at TEXT NOT NULL,
 updated_at TEXT NOT NULL,
 FOREIGN KEY(owner_epoch,fencing_token) REFERENCES strategy_dispatch_owner_epochs(owner_epoch,fencing_token),
 FOREIGN KEY(account_ref,market,symbol) REFERENCES strategy_dispatch_market_authorities(account_ref,market,symbol),
 FOREIGN KEY(risk_reservation_id) REFERENCES risk_reservations(id),
 FOREIGN KEY(guardian_decision_id) REFERENCES risk_bucket_final_decisions(decision_id),
 CHECK(expires_at>issued_at),
 CHECK((state IN ('ISSUED','CLAIMED','SUBMITTING') AND disposition='RESERVED')
    OR (state='SUBMITTED' AND disposition='TRANSFERRED' AND broker_order_id<>'')
    OR (state='REFUSED' AND disposition='RELEASED')
    OR (state='AMBIGUOUS' AND disposition IN ('HELD','RELEASED','TRANSFERRED'))),
 CHECK((state IN ('ISSUED','CLAIMED') AND transport_started_at IS NULL)
    OR (state='REFUSED')
    OR (state IN ('SUBMITTING','SUBMITTED','AMBIGUOUS') AND transport_started_at IS NOT NULL))
) STRICT;
CREATE INDEX idx_strategy_dispatch_lease_market_state
 ON strategy_dispatch_leases(market,state,account_ref,symbol);
CREATE INDEX idx_strategy_dispatch_lease_owner
 ON strategy_dispatch_leases(owner_epoch,fencing_token,state);
CREATE INDEX idx_strategy_dispatch_lease_recovery
 ON strategy_dispatch_leases(issued_at,lease_id) WHERE state IN ('ISSUED','CLAIMED','SUBMITTING');
CREATE UNIQUE INDEX uq_strategy_dispatch_account_broker_order
 ON strategy_dispatch_leases(account_ref,broker_order_id) WHERE broker_order_id<>'';
CREATE TABLE strategy_dispatch_outcomes (
 outcome_id TEXT PRIMARY KEY,
 lease_id TEXT NOT NULL REFERENCES strategy_dispatch_leases(lease_id),
 from_state TEXT NOT NULL,
 to_state TEXT NOT NULL,
 from_disposition TEXT NOT NULL,
 to_disposition TEXT NOT NULL,
 expected_revision INTEGER NOT NULL CHECK(expected_revision>0),
 next_revision INTEGER NOT NULL CHECK(next_revision=expected_revision+1),
 transition_code TEXT NOT NULL,
 operation_identity TEXT NOT NULL,
 broker_order_id TEXT NOT NULL DEFAULT '',
 query_digest TEXT NOT NULL DEFAULT '',
 observed_at TEXT NOT NULL,
 record_digest TEXT NOT NULL UNIQUE,
 UNIQUE(lease_id,next_revision)
) STRICT;
CREATE INDEX idx_strategy_dispatch_outcome_lease
 ON strategy_dispatch_outcomes(lease_id,next_revision);
CREATE TRIGGER strategy_dispatch_owner_epochs_no_update BEFORE UPDATE ON strategy_dispatch_owner_epochs BEGIN SELECT RAISE(ABORT,'strategy dispatch owner epochs are immutable'); END;
CREATE TRIGGER strategy_dispatch_owner_epochs_no_delete BEFORE DELETE ON strategy_dispatch_owner_epochs BEGIN SELECT RAISE(ABORT,'strategy dispatch owner epochs are immutable'); END;
CREATE TRIGGER strategy_dispatch_owner_current_monotonic BEFORE UPDATE ON strategy_dispatch_owner_current
WHEN NEW.owner_epoch<=OLD.owner_epoch OR NEW.revision<>OLD.revision+1 OR NEW.fencing_token=OLD.fencing_token
BEGIN SELECT RAISE(ABORT,'strategy dispatch owner fence must advance'); END;
CREATE TRIGGER strategy_dispatch_owner_current_no_delete BEFORE DELETE ON strategy_dispatch_owner_current BEGIN SELECT RAISE(ABORT,'strategy dispatch current owner cannot be deleted'); END;
CREATE TRIGGER strategy_dispatch_authority_revision BEFORE UPDATE ON strategy_dispatch_market_authorities
WHEN NEW.revision<>OLD.revision+1 BEGIN SELECT RAISE(ABORT,'strategy dispatch authority revision must advance'); END;
CREATE TRIGGER strategy_dispatch_authority_no_delete BEFORE DELETE ON strategy_dispatch_market_authorities BEGIN SELECT RAISE(ABORT,'strategy dispatch authority cannot be deleted'); END;
CREATE TRIGGER strategy_dispatch_lease_qfinal_authority BEFORE INSERT ON strategy_dispatch_leases
WHEN NOT EXISTS (
 SELECT 1
 FROM risk_bucket_final_decisions q
 JOIN decisions d ON d.id=q.decision_id
 JOIN risk_reservations aggregate_hold
   ON aggregate_hold.id=q.existing_reservation_id
  AND aggregate_hold.decision_id=d.id
 JOIN risk_bucket_owners owner
   ON owner.account_ref=q.account_ref
  AND owner.market=q.market
  AND owner.symbol=q.symbol
  AND owner.prospective_generation=q.owner_prospective_generation
  AND owner.lane_id=q.owner_lane_id
  AND owner.campaign_id=q.owner_campaign_id
  AND owner.released_at IS NULL
 JOIN strategy_dispatch_market_authorities authority
   ON authority.account_ref=q.account_ref
  AND authority.market=q.market
  AND authority.symbol=q.symbol
  AND authority.revision=NEW.authority_revision
  AND authority.record_digest=NEW.authority_digest
 JOIN strategy_dispatch_owner_current dispatch_owner
   ON dispatch_owner.owner_key='CENTRAL'
  AND dispatch_owner.owner_epoch=NEW.owner_epoch
  AND dispatch_owner.fencing_token=NEW.fencing_token
 WHERE q.decision_id=NEW.guardian_decision_id
   AND q.existing_reservation_id=NEW.risk_reservation_id
   AND q.account_ref=NEW.account_ref
   AND q.market=NEW.market
   AND q.symbol=NEW.symbol
   AND q.owner_lane_id=NEW.lane_id
   AND q.owner_campaign_id=NEW.campaign_id
   AND d.account_ref=NEW.account_ref
   AND d.safety_class='EXPOSURE_RAISING'
   AND d.preimage_kind='RISK_INTENT'
   AND d.issued_at<=NEW.issued_at
   AND d.expires_at>=NEW.expires_at
   AND q.created_at<=NEW.issued_at
   AND aggregate_hold.account_ref=NEW.account_ref
   AND aggregate_hold.state='HELD'
   AND (SELECT count(*) FROM risk_bucket_reservations monetary
        WHERE monetary.decision_id=q.decision_id
          AND monetary.existing_reservation_id=q.existing_reservation_id
          AND monetary.account_ref=q.account_ref
          AND monetary.market=q.market
          AND monetary.symbol=q.symbol
          AND monetary.owner_prospective_generation=q.owner_prospective_generation
          AND monetary.state='HELD'
          AND monetary.held_minor=monetary.reserved_minor)=5
   AND (SELECT count(DISTINCT monetary.bucket_dimension) FROM risk_bucket_reservations monetary
        WHERE monetary.decision_id=q.decision_id
          AND monetary.bucket_dimension IN ('horizon','market','strategy','sector','symbol'))=5
)
BEGIN SELECT RAISE(ABORT,'strategy dispatch lease requires exact q_final authority'); END;
CREATE TRIGGER strategy_dispatch_lease_transition BEFORE UPDATE ON strategy_dispatch_leases
WHEN NEW.revision<>OLD.revision+1 OR NOT (
 (OLD.state='ISSUED' AND NEW.state IN ('CLAIMED','REFUSED'))
 OR (OLD.state='CLAIMED' AND NEW.state IN ('SUBMITTING','REFUSED'))
 OR (OLD.state='SUBMITTING' AND NEW.state IN ('SUBMITTED','AMBIGUOUS','REFUSED'))
 OR (OLD.state='AMBIGUOUS' AND OLD.disposition='HELD' AND NEW.state='AMBIGUOUS' AND NEW.disposition IN ('RELEASED','TRANSFERRED')))
BEGIN SELECT RAISE(ABORT,'invalid irreversible strategy dispatch lease transition'); END;
CREATE TRIGGER strategy_dispatch_lease_no_delete BEFORE DELETE ON strategy_dispatch_leases BEGIN SELECT RAISE(ABORT,'strategy dispatch leases cannot be deleted'); END;
CREATE TRIGGER strategy_dispatch_outcome_no_update BEFORE UPDATE ON strategy_dispatch_outcomes BEGIN SELECT RAISE(ABORT,'strategy dispatch outcomes are immutable'); END;
CREATE TRIGGER strategy_dispatch_outcome_no_delete BEFORE DELETE ON strategy_dispatch_outcomes BEGIN SELECT RAISE(ABORT,'strategy dispatch outcomes are immutable'); END;
