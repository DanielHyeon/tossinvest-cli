package strategyruntime

import (
	"crypto/sha256"
	"errors"
)

type Lineage struct {
	CandidateID        string
	CandidateDigest    string
	EvidenceDigest     string
	RouterDecisionID   string
	LaneID             string
	LaneVersion        string
	CampaignID         string
	LegID              string
	RiskPolicyDigest   string
	ReservationID      string
	GuardianDecisionID string
	AttemptID          string
	seal               [32]byte
}

type lineageInput Lineage

func newLineage(input lineageInput) (Lineage, error) {
	lineage := Lineage(input)
	lineage.seal = [32]byte{}
	if !completeLineage(lineage) {
		return Lineage{}, errors.New("strategyruntime: incomplete lineage")
	}
	lineage.seal = lineageSeal(lineage)
	return lineage, nil
}

func completeLineage(lineage Lineage) bool {
	for _, value := range []string{lineage.CandidateID, lineage.RouterDecisionID, lineage.LaneID, lineage.LaneVersion, lineage.CampaignID, lineage.LegID, lineage.ReservationID, lineage.GuardianDecisionID, lineage.AttemptID} {
		if !validIdentity(value) {
			return false
		}
	}
	for _, value := range []string{lineage.CandidateDigest, lineage.EvidenceDigest, lineage.RiskPolicyDigest} {
		if !validDigest(value) {
			return false
		}
	}
	return true
}

func lineageSeal(lineage Lineage) [32]byte {
	return hashStrings(lineage.CandidateID, lineage.CandidateDigest, lineage.EvidenceDigest, lineage.RouterDecisionID, lineage.LaneID,
		lineage.LaneVersion, lineage.CampaignID, lineage.LegID, lineage.RiskPolicyDigest, lineage.ReservationID, lineage.GuardianDecisionID, lineage.AttemptID)
}

func validLineage(lineage Lineage) bool {
	return completeLineage(lineage) && lineage.seal == lineageSeal(lineage)
}

type generationBinding struct {
	Generation uint64
	Digest     string
}

func validGenerationBinding(binding generationBinding) bool {
	return binding.Generation > 0 && validDigest(binding.Digest)
}

type ProtectionState string

const ProtectionWired ProtectionState = "WIRED"

type protectionBinding struct {
	Generation        uint64
	AttestationSerial uint64
	Digest            string
	State             ProtectionState
}

type authoritySnapshotInput struct {
	AccountID      string
	Market         Market
	Symbol         string
	Activation     generationBinding
	Calendar       generationBinding
	Protection     protectionBinding
	Reconciliation generationBinding
	Risk           generationBinding
	Guardian       generationBinding
	Build          generationBinding
}

type authoritySnapshot struct {
	accountID      string
	market         Market
	symbol         string
	activation     generationBinding
	calendar       generationBinding
	protection     protectionBinding
	reconciliation generationBinding
	risk           generationBinding
	guardian       generationBinding
	build          generationBinding
	seal           [32]byte
}

func newAuthoritySnapshot(input authoritySnapshotInput) (authoritySnapshot, error) {
	if !validIdentity(input.AccountID) || !validMarket(input.Market) || !validIdentity(input.Symbol) || !validGenerationBinding(input.Activation) || !validGenerationBinding(input.Calendar) ||
		input.Protection.Generation == 0 || input.Protection.AttestationSerial == 0 || !validDigest(input.Protection.Digest) || input.Protection.State != ProtectionWired ||
		!validGenerationBinding(input.Reconciliation) || !validGenerationBinding(input.Risk) || !validGenerationBinding(input.Guardian) || !validGenerationBinding(input.Build) {
		return authoritySnapshot{}, errors.New("strategyruntime: incomplete authority snapshot")
	}
	snapshot := authoritySnapshot{accountID: input.AccountID, market: input.Market, symbol: input.Symbol, activation: input.Activation, calendar: input.Calendar,
		protection: input.Protection, reconciliation: input.Reconciliation, risk: input.Risk, guardian: input.Guardian, build: input.Build}
	snapshot.seal = authoritySeal(snapshot)
	return snapshot, nil
}

func authoritySeal(authority authoritySnapshot) [32]byte {
	hash := sha256.New()
	for _, value := range []string{authority.accountID, string(authority.market), authority.symbol} {
		writeString(hash, value)
	}
	for _, binding := range []generationBinding{authority.activation, authority.calendar, authority.reconciliation, authority.risk, authority.guardian, authority.build} {
		writeUint64(hash, binding.Generation)
		writeString(hash, binding.Digest)
	}
	writeUint64(hash, authority.protection.Generation)
	writeUint64(hash, authority.protection.AttestationSerial)
	writeString(hash, authority.protection.Digest)
	writeString(hash, string(authority.protection.State))
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func validAuthority(authority authoritySnapshot) bool {
	return validIdentity(authority.accountID) && validMarket(authority.market) && validIdentity(authority.symbol) && authority.seal == authoritySeal(authority) &&
		validGenerationBinding(authority.activation) && validGenerationBinding(authority.calendar) && authority.protection.Generation > 0 &&
		authority.protection.AttestationSerial > 0 && validDigest(authority.protection.Digest) && authority.protection.State == ProtectionWired &&
		validGenerationBinding(authority.reconciliation) && validGenerationBinding(authority.risk) && validGenerationBinding(authority.guardian) && validGenerationBinding(authority.build)
}

func (authority authoritySnapshot) Market() Market { return authority.market }

type ownerFence struct {
	epoch uint64
	token string
	seal  [32]byte
}

func newOwnerFence(epoch uint64, token string) (ownerFence, error) {
	if epoch == 0 || !validIdentity(token) {
		return ownerFence{}, errors.New("strategyruntime: invalid owner fence")
	}
	owner := ownerFence{epoch: epoch, token: token}
	owner.seal = ownerFenceSeal(owner)
	return owner, nil
}

func ownerFenceSeal(owner ownerFence) [32]byte {
	hash := sha256.New()
	writeUint64(hash, owner.epoch)
	writeString(hash, owner.token)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func validOwnerFence(owner ownerFence) bool {
	return owner.epoch > 0 && validIdentity(owner.token) && owner.seal == ownerFenceSeal(owner)
}

func restartOwner(current ownerFence, nextToken string) (ownerFence, error) {
	if !validOwnerFence(current) || current.epoch == ^uint64(0) || !validIdentity(nextToken) || nextToken == current.token {
		return ownerFence{}, errors.New("strategyruntime: invalid owner restart")
	}
	return newOwnerFence(current.epoch+1, nextToken)
}
