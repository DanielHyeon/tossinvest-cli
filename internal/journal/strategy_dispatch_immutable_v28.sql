CREATE TRIGGER strategy_dispatch_lease_authority_immutable_v28
BEFORE UPDATE ON strategy_dispatch_leases
WHEN NEW.lease_id IS NOT OLD.lease_id
  OR NEW.operation_id IS NOT OLD.operation_id
  OR NEW.account_ref IS NOT OLD.account_ref
  OR NEW.market IS NOT OLD.market
  OR NEW.symbol IS NOT OLD.symbol
  OR NEW.candidate_id IS NOT OLD.candidate_id
  OR NEW.evidence_digest IS NOT OLD.evidence_digest
  OR NEW.router_id IS NOT OLD.router_id
  OR NEW.router_version IS NOT OLD.router_version
  OR NEW.lane_id IS NOT OLD.lane_id
  OR NEW.lane_version IS NOT OLD.lane_version
  OR NEW.campaign_id IS NOT OLD.campaign_id
  OR NEW.leg_id IS NOT OLD.leg_id
  OR NEW.risk_reservation_id IS NOT OLD.risk_reservation_id
  OR NEW.guardian_decision_id IS NOT OLD.guardian_decision_id
  OR NEW.owner_epoch IS NOT OLD.owner_epoch
  OR NEW.fencing_token IS NOT OLD.fencing_token
  OR NEW.authority_revision IS NOT OLD.authority_revision
  OR NEW.authority_digest IS NOT OLD.authority_digest
  OR NEW.issued_at IS NOT OLD.issued_at
  OR NEW.expires_at IS NOT OLD.expires_at
  OR NEW.lease_digest IS NOT OLD.lease_digest
  OR NEW.created_at IS NOT OLD.created_at
BEGIN
 SELECT RAISE(ABORT,'strategy dispatch lease authority is immutable');
END;
