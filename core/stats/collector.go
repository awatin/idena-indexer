package stats

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/idena-network/idena-go/blockchain"
	"github.com/idena-network/idena-go/blockchain/types"
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/common/math"
	"github.com/idena-network/idena-go/core/appstate"
	"github.com/idena-network/idena-go/core/state"
	"github.com/idena-network/idena-go/stats/collector"
	statsTypes "github.com/idena-network/idena-go/stats/types"
	"github.com/idena-network/idena-indexer/core/conversion"
	"github.com/idena-network/idena-indexer/db"
	"github.com/idena-network/idena-indexer/log"
	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
)

type statsCollector struct {
	stats        *Stats
	statsEnabled bool
	pending      *pending
}

type pending struct {
	invitationRewardsByAddrAndType  map[common.Address]map[RewardType]*RewardStats
	balanceUpdates                  []*db.BalanceUpdate
	epochRewardBalanceUpdatesByAddr map[common.Address]*db.BalanceUpdate
	identityStates                  []state.IdentityState
	tx                              *pendingTx
}

type pendingTx struct {
	tx                            *types.Transaction
	factEvidenceContractDeploy    *db.FactEvidenceContract
	factEvidenceContractCallStart *db.FactEvidenceContractCallStart
}

func NewStatsCollector() collector.StatsCollector {
	return &statsCollector{}
}

func (c *statsCollector) EnableCollecting() {
	c.stats = &Stats{}
	c.pending = &pending{}
}

func (c *statsCollector) initRewardStats() {
	if c.stats.RewardsStats != nil {
		return
	}
	c.stats.RewardsStats = &RewardsStats{}
}

func (c *statsCollector) initInvitationRewardsByAddrAndType() {
	if c.pending.invitationRewardsByAddrAndType != nil {
		return
	}
	c.pending.invitationRewardsByAddrAndType = make(map[common.Address]map[RewardType]*RewardStats)
}

func (c *statsCollector) SetValidation(validation *statsTypes.ValidationStats) {
	c.stats.ValidationStats = validation
}

func (c *statsCollector) SetMinScoreForInvite(score float32) {
	c.stats.MinScoreForInvite = &score
}

func (c *statsCollector) SetValidationResults(validationResults *types.ValidationResults) {
	c.initRewardStats()
	c.stats.RewardsStats.ValidationResults = validationResults
}

func (c *statsCollector) SetTotalReward(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Total = amount
}

func (c *statsCollector) SetTotalValidationReward(amount *big.Int, share *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Validation = amount
	c.stats.RewardsStats.ValidationShare = share
}

func (c *statsCollector) SetTotalFlipsReward(amount *big.Int, share *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Flips = amount
	c.stats.RewardsStats.FlipsShare = share
}

func (c *statsCollector) SetTotalInvitationsReward(amount *big.Int, share *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.Invitations = amount
	c.stats.RewardsStats.InvitationsShare = share
}

func (c *statsCollector) SetTotalFoundationPayouts(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.FoundationPayouts = amount
}

func (c *statsCollector) SetTotalZeroWalletFund(amount *big.Int) {
	c.initRewardStats()
	c.stats.RewardsStats.ZeroWalletFund = amount
}

func (c *statsCollector) AddValidationReward(addr common.Address, age uint16, balance *big.Int, stake *big.Int) {
	c.addReward(addr, balance, stake, Validation)
	if c.stats.RewardsStats.AgesByAddress == nil {
		c.stats.RewardsStats.AgesByAddress = make(map[string]uint16)
	}
	c.stats.RewardsStats.AgesByAddress[conversion.ConvertAddress(addr)] = age + 1
}

func (c *statsCollector) AddFlipsReward(addr common.Address, balance *big.Int, stake *big.Int,
	rewardedStrongFlipCids [][]byte, rewardedWeakFlipCids [][]byte) {
	c.addReward(addr, balance, stake, Flips)
	c.addRewardedFlips(rewardedStrongFlipCids, rewardedWeakFlipCids)
}

func (c *statsCollector) addRewardedFlips(rewardedStrongFlipCids [][]byte, rewardedWeakFlipCids [][]byte) {
	if len(rewardedStrongFlipCids)+len(rewardedWeakFlipCids) == 0 {
		return
	}
	c.initRewardStats()
	for _, cidBytes := range rewardedStrongFlipCids {
		flipCid, _ := cid.Parse(cidBytes)
		c.stats.RewardsStats.RewardedFlipCids = append(c.stats.RewardsStats.RewardedFlipCids, flipCid.String())
	}
	for _, cidBytes := range rewardedWeakFlipCids {
		flipCid, _ := cid.Parse(cidBytes)
		c.stats.RewardsStats.RewardedFlipCids = append(c.stats.RewardsStats.RewardedFlipCids, flipCid.String())
	}
}

func (c *statsCollector) AddInvitationsReward(addr common.Address, balance *big.Int, stake *big.Int, age uint16,
	txHash *common.Hash, isSavedInviteWinner bool) {
	rewardType, err := determineInvitationsRewardType(age, isSavedInviteWinner)
	if err != nil {
		log.Warn(err.Error())
		return
	}
	c.addReward(addr, balance, stake, rewardType)
	c.addRewardedInvite(addr, txHash, rewardType)
}

func determineInvitationsRewardType(age uint16, isSavedInviteWinner bool) (RewardType, error) {
	switch age {
	case 0:
		if isSavedInviteWinner {
			return SavedInviteWin, nil
		}
		return SavedInvite, nil
	case 1:
		return Invitations, nil
	case 2:
		return Invitations2, nil
	case 3:
		return Invitations3, nil
	default:
		return 0, errors.Errorf("no invitations reward type for age: %v, isSavedInviteWinner: %v", age, isSavedInviteWinner)
	}
}

func (c *statsCollector) addRewardedInvite(addr common.Address, txHash *common.Hash, rewardType RewardType) {
	if rewardType == SavedInviteWin || rewardType == SavedInvite {
		if c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType == nil {
			c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType = make(map[common.Address]map[RewardType]uint8)
		}
		if _, ok := c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType[addr]; !ok {
			c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType[addr] = make(map[RewardType]uint8)
		}
		c.stats.RewardsStats.SavedInviteRewardsCountByAddrAndType[addr][rewardType]++
		return
	}
	if txHash == nil {
		log.Warn(fmt.Sprintf("wrong value txHash=nil for rewardType=%v", rewardType))
		return
	}
	c.stats.RewardsStats.RewardedInvites = append(c.stats.RewardsStats.RewardedInvites, &db.RewardedInvite{
		TxHash: conversion.ConvertHash(*txHash),
		Type:   byte(rewardType),
	})
}

func (c *statsCollector) AddFoundationPayout(addr common.Address, balance *big.Int) {
	c.addReward(addr, balance, nil, FoundationPayouts)
}

func (c *statsCollector) AddZeroWalletFund(addr common.Address, balance *big.Int) {
	c.addReward(addr, balance, nil, ZeroWalletFund)
}

func (c *statsCollector) addReward(addr common.Address, balance *big.Int, stake *big.Int, rewardType RewardType) {
	if (balance == nil || balance.Sign() == 0) && (stake == nil || stake.Sign() == 0) {
		return
	}
	c.initRewardStats()
	rewardsStats := &RewardStats{
		Address: addr,
		Balance: balance,
		Stake:   stake,
		Type:    rewardType,
	}
	if c.increaseInvitationRewardIfExists(rewardsStats) {
		return
	}
	c.stats.RewardsStats.Rewards = append(c.stats.RewardsStats.Rewards, rewardsStats)
}

func (c *statsCollector) increaseInvitationRewardIfExists(rewardsStats *RewardStats) bool {
	if rewardsStats.Type != Invitations && rewardsStats.Type != Invitations2 && rewardsStats.Type != Invitations3 &&
		rewardsStats.Type != SavedInvite && rewardsStats.Type != SavedInviteWin {
		return false
	}
	c.initInvitationRewardsByAddrAndType()
	addrInvitationRewardsByType, ok := c.pending.invitationRewardsByAddrAndType[rewardsStats.Address]
	if ok {
		if ir, ok := addrInvitationRewardsByType[rewardsStats.Type]; ok {
			ir.Balance.Add(ir.Balance, rewardsStats.Balance)
			ir.Stake.Add(ir.Stake, rewardsStats.Stake)
			return true
		}
	} else {
		addrInvitationRewardsByType = make(map[RewardType]*RewardStats)
	}
	addrInvitationRewardsByType[rewardsStats.Type] = rewardsStats
	c.pending.invitationRewardsByAddrAndType[rewardsStats.Address] = addrInvitationRewardsByType
	return false
}

func (c *statsCollector) AddProposerReward(addr common.Address, balance *big.Int, stake *big.Int) {
	c.addMiningReward(addr, balance, stake, true)
}

func (c *statsCollector) AddFinalCommitteeReward(addr common.Address, balance *big.Int, stake *big.Int) {
	c.addMiningReward(addr, balance, stake, false)
	c.stats.FinalCommittee = append(c.stats.FinalCommittee, addr)
}

func (c *statsCollector) addMiningReward(addr common.Address, balance *big.Int, stake *big.Int, isProposerReward bool) {
	c.stats.MiningRewards = append(c.stats.MiningRewards, &db.MiningReward{
		Address:  conversion.ConvertAddress(addr),
		Balance:  blockchain.ConvertToFloat(balance),
		Stake:    blockchain.ConvertToFloat(stake),
		Proposer: isProposerReward,
	})
}

func (c *statsCollector) AfterSubPenalty(addr common.Address, amount *big.Int, appState *appstate.AppState) {
	if amount == nil || amount.Sign() != 1 {
		return
	}
	c.detectAndCollectCompletedPenalty(addr, appState)
}

func (c *statsCollector) detectAndCollectCompletedPenalty(addr common.Address, appState *appstate.AppState) {
	updatedPenalty := appState.State.GetPenalty(addr)
	if updatedPenalty != nil && updatedPenalty.Sign() == 1 {
		return
	}
	c.initBurntPenaltiesByAddr()
	c.stats.BurntPenaltiesByAddr[addr] = updatedPenalty
}

func (c *statsCollector) BeforeClearPenalty(addr common.Address, appState *appstate.AppState) {
	c.detectAndCollectBurntPenalty(addr, appState)
}

func (c *statsCollector) BeforeSetPenalty(addr common.Address, appState *appstate.AppState) {
	c.detectAndCollectBurntPenalty(addr, appState)
}

func (c *statsCollector) detectAndCollectBurntPenalty(addr common.Address, appState *appstate.AppState) {
	curPenalty := appState.State.GetPenalty(addr)
	if curPenalty == nil || curPenalty.Sign() != 1 {
		return
	}
	c.initBurntPenaltiesByAddr()
	c.stats.BurntPenaltiesByAddr[addr] = curPenalty
}

func (c *statsCollector) initBurntPenaltiesByAddr() {
	if c.stats.BurntPenaltiesByAddr != nil {
		return
	}
	c.stats.BurntPenaltiesByAddr = make(map[common.Address]*big.Int)
}

func (c *statsCollector) AddMintedCoins(amount *big.Int) {
	if amount == nil {
		return
	}
	if c.stats.MintedCoins == nil {
		c.stats.MintedCoins = big.NewInt(0)
	}
	c.stats.MintedCoins.Add(c.stats.MintedCoins, amount)
}

func (c *statsCollector) addBurntCoins(addr common.Address, amount *big.Int, reason db.BurntCoinsReason, tx *types.Transaction) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	if c.stats.BurntCoins == nil {
		c.stats.BurntCoins = big.NewInt(0)
	}
	c.stats.BurntCoins.Add(c.stats.BurntCoins, amount)
	if c.stats.BurntCoinsByAddr == nil {
		c.stats.BurntCoinsByAddr = make(map[common.Address][]*db.BurntCoins)
	}
	var txHash string
	if tx != nil {
		txHash = tx.Hash().Hex()
	}
	c.stats.BurntCoinsByAddr[addr] = append(c.stats.BurntCoinsByAddr[addr], &db.BurntCoins{
		Amount: blockchain.ConvertToFloat(amount),
		Reason: reason,
		TxHash: txHash,
	})
}

func (c *statsCollector) AddPenaltyBurntCoins(addr common.Address, amount *big.Int) {
	c.addBurntCoins(addr, amount, db.PenaltyBurntCoins, nil)
}

func (c *statsCollector) AddInviteBurntCoins(addr common.Address, amount *big.Int, tx *types.Transaction) {
	c.addBurntCoins(addr, amount, db.InviteBurntCoins, tx)
}

func (c *statsCollector) AddFeeBurntCoins(addr common.Address, feeAmount *big.Int, burntRate float32, tx *types.Transaction) {
	if feeAmount == nil || feeAmount.Sign() == 0 {
		return
	}
	burntFee := decimal.NewFromBigInt(feeAmount, 0)
	burntFee = burntFee.Mul(decimal.NewFromFloat32(burntRate))
	c.addBurntCoins(addr, math.ToInt(burntFee), db.FeeBurntCoins, tx)
}

func (c *statsCollector) AddKilledBurntCoins(addr common.Address, amount *big.Int) {
	c.addBurntCoins(addr, amount, db.KilledBurntCoins, nil)
}

func (c *statsCollector) AddBurnTxBurntCoins(addr common.Address, tx *types.Transaction) {
	c.addBurntCoins(addr, tx.AmountOrZero(), db.BurnTxBurntCoins, tx)
}

func (c *statsCollector) afterBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.initBalanceUpdatesByAddr()
	c.stats.BalanceUpdateAddrs.Add(addr)
}

func (c *statsCollector) initBalanceUpdatesByAddr() {
	if c.stats.BalanceUpdateAddrs != nil {
		return
	}
	c.stats.BalanceUpdateAddrs = mapset.NewSet()
}

func (c *statsCollector) CompleteCollecting() {
	c.stats = nil
	c.pending = nil
}

func (c *statsCollector) AfterAddStake(addr common.Address, amount *big.Int, appState *appstate.AppState) {
	if appState.State.GetIdentityState(addr) == state.Killed {
		c.addBurntCoins(addr, amount, db.KilledBurntCoins, nil)
	}
}

func (c *statsCollector) AddActivationTxBalanceTransfer(tx *types.Transaction, amount *big.Int) {
	sender, _ := types.Sender(tx)
	if sender == *tx.To {
		return
	}
	if amount == nil || amount.Sign() == 0 {
		return
	}
	c.stats.ActivationTxTransfers = append(c.stats.ActivationTxTransfers, db.ActivationTxTransfer{
		TxHash:          conversion.ConvertHash(tx.Hash()),
		BalanceTransfer: blockchain.ConvertToFloat(amount),
	})
}

func (c *statsCollector) AddKillTxStakeTransfer(tx *types.Transaction, amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	c.stats.KillTxTransfers = append(c.stats.KillTxTransfers, db.KillTxTransfer{
		TxHash:        conversion.ConvertHash(tx.Hash()),
		StakeTransfer: blockchain.ConvertToFloat(amount),
	})
}

func (c *statsCollector) AddKillInviteeTxStakeTransfer(tx *types.Transaction, amount *big.Int) {
	if amount == nil || amount.Sign() == 0 {
		return
	}
	c.stats.KillInviteeTxTransfers = append(c.stats.KillInviteeTxTransfers, db.KillInviteeTxTransfer{
		TxHash:        conversion.ConvertHash(tx.Hash()),
		StakeTransfer: blockchain.ConvertToFloat(amount),
	})
}

func (c *statsCollector) BeginVerifiedStakeTransferBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.VerifiedStakeTransferReason, nil)
}

func (c *statsCollector) BeginTxBalanceUpdate(tx *types.Transaction, appState *appstate.AppState) {
	sender, _ := types.Sender(tx)
	txHash := tx.Hash()
	c.addPendingBalanceUpdate(sender, appState, db.TxReason, &txHash)
	if tx.To != nil && *tx.To != sender {
		c.addPendingBalanceUpdate(*tx.To, appState, db.TxReason, &txHash)
	}
}

func (c *statsCollector) BeginProposerRewardBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.ProposerRewardReason, nil)
}

func (c *statsCollector) BeginCommitteeRewardBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.CommitteeRewardReason, nil)
}

func (c *statsCollector) BeginEpochRewardBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.EpochRewardReason, nil)
}

func (c *statsCollector) BeginFailedValidationBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.FailedValidationReason, nil)
}

func (c *statsCollector) BeginPenaltyBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.PenaltyReason, nil)
}

func (c *statsCollector) BeginEpochPenaltyResetBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.EpochPenaltyResetReason, nil)
}

func (c *statsCollector) BeginDustClearingBalanceUpdate(addr common.Address, appState *appstate.AppState) {
	c.addPendingBalanceUpdate(addr, appState, db.DustClearingReason, nil)
}

func (c *statsCollector) CompleteBalanceUpdate(appState *appstate.AppState) {
	balanceUpdates := c.completeBalanceUpdates(appState)
	for _, balanceUpdate := range balanceUpdates {
		if !isBalanceChanged(balanceUpdate) {
			continue
		}
		if balanceUpdate.Reason == db.DustClearingReason {
			c.addBurntCoins(balanceUpdate.Address, balanceUpdate.BalanceOld, db.DustClearingBurntCoins, nil)
		}
		if balanceUpdate.Reason == db.EpochRewardReason {
			if c.pending.epochRewardBalanceUpdatesByAddr == nil {
				c.pending.epochRewardBalanceUpdatesByAddr = map[common.Address]*db.BalanceUpdate{
					balanceUpdate.Address: balanceUpdate,
				}
			} else if bu, ok := c.pending.epochRewardBalanceUpdatesByAddr[balanceUpdate.Address]; ok {
				bu.BalanceNew = balanceUpdate.BalanceNew
				bu.StakeNew = balanceUpdate.StakeNew
				bu.PenaltyNew = balanceUpdate.PenaltyNew
				continue
			} else {
				c.pending.epochRewardBalanceUpdatesByAddr[balanceUpdate.Address] = balanceUpdate
			}
		}
		c.stats.BalanceUpdates = append(c.stats.BalanceUpdates, balanceUpdate)
		c.afterBalanceUpdate(balanceUpdate.Address, appState)
	}
}

func isBalanceChanged(balanceUpdate *db.BalanceUpdate) bool {
	return balanceUpdate.BalanceOld.Cmp(balanceUpdate.BalanceNew) != 0 ||
		balanceUpdate.StakeOld.Cmp(balanceUpdate.StakeNew) != 0 ||
		valueOrZero(balanceUpdate.PenaltyOld).Cmp(valueOrZero(balanceUpdate.PenaltyNew)) != 0
}

func valueOrZero(v *big.Int) *big.Int {
	if v == nil {
		return common.Big0
	}
	return v
}

func (c *statsCollector) addPendingBalanceUpdate(
	addr common.Address,
	appState *appstate.AppState,
	reason db.BalanceUpdateReason,
	txHash *common.Hash,
) {
	c.pending.balanceUpdates = append(c.pending.balanceUpdates, &db.BalanceUpdate{
		Address:    addr,
		BalanceOld: appState.State.GetBalance(addr),
		StakeOld:   c.getStakeIfNotKilled(addr, appState),
		PenaltyOld: c.getPenaltyIfNotKilled(addr, appState),
		Reason:     reason,
		TxHash:     txHash,
	})
}

func (c *statsCollector) completeBalanceUpdates(appState *appstate.AppState) []*db.BalanceUpdate {
	for _, balanceUpdate := range c.pending.balanceUpdates {
		balanceUpdate.BalanceNew = appState.State.GetBalance(balanceUpdate.Address)
		balanceUpdate.StakeNew = c.getStakeIfNotKilled(balanceUpdate.Address, appState)
		balanceUpdate.PenaltyNew = c.getPenaltyIfNotKilled(balanceUpdate.Address, appState)
	}
	balanceUpdates := c.pending.balanceUpdates
	c.pending.balanceUpdates = nil
	return balanceUpdates
}

func (c *statsCollector) getStakeIfNotKilled(addr common.Address, appState *appstate.AppState) *big.Int {
	if appState.State.GetIdentityState(addr) == state.Killed {
		return common.Big0
	}
	return appState.State.GetStakeBalance(addr)
}

func (c *statsCollector) getPenaltyIfNotKilled(addr common.Address, appState *appstate.AppState) *big.Int {
	if appState.State.GetIdentityState(addr) == state.Killed {
		return common.Big0
	}
	return appState.State.GetIdentity(addr).Penalty // State.GetPenalty is not used since it may add new identity to state
}

func (c *statsCollector) SetCommitteeRewardShare(amount *big.Int) {
	c.stats.CommitteeRewardShare = amount
}

func (c *statsCollector) BeginApplyingTx(tx *types.Transaction, appState *appstate.AppState) {
	c.pending.tx = &pendingTx{
		tx: tx,
	}
	sender, _ := types.Sender(tx)
	senderState := appState.State.GetIdentityState(sender)
	c.pending.identityStates = []state.IdentityState{senderState}
	if tx.To != nil && *tx.To != sender {
		recipientState := appState.State.GetIdentityState(*tx.To)
		c.pending.identityStates = append(c.pending.identityStates, recipientState)
	}
}

func (c *statsCollector) CompleteApplyingTx(appState *appstate.AppState) {
	tx := c.pending.tx.tx
	var changesByAddress map[common.Address]*IdentityStateChange
	initChangesByAddress := func() {
		if changesByAddress != nil {
			return
		}
		changesByAddress = make(map[common.Address]*IdentityStateChange)
	}
	sender, _ := types.Sender(tx)
	senderState := appState.State.GetIdentityState(sender)
	if c.pending.identityStates[0] != senderState {
		initChangesByAddress()
		changesByAddress[sender] = &IdentityStateChange{
			PrevState: c.pending.identityStates[0],
			NewState:  senderState,
		}
	}
	if tx.To != nil && *tx.To != sender {
		recipientState := appState.State.GetIdentityState(*tx.To)
		if c.pending.identityStates[1] != recipientState {
			initChangesByAddress()
			changesByAddress[*tx.To] = &IdentityStateChange{
				PrevState: c.pending.identityStates[1],
				NewState:  recipientState,
			}
		}
	}
	if len(changesByAddress) > 0 {
		if c.stats.IdentityStateChangesByTxHashAndAddress == nil {
			c.stats.IdentityStateChangesByTxHashAndAddress = make(map[common.Hash]map[common.Address]*IdentityStateChange)
		}
		c.stats.IdentityStateChangesByTxHashAndAddress[tx.Hash()] = changesByAddress
	}
	c.pending.tx = nil
}

func (c *statsCollector) AddTxFee(feeAmount *big.Int) {
	tx := c.pending.tx.tx
	if c.stats.FeesByTxHash == nil {
		c.stats.FeesByTxHash = make(map[common.Hash]*big.Int)
	}
	c.stats.FeesByTxHash[tx.Hash()] = feeAmount
}

func (c *statsCollector) AddFactEvidenceContractDeploy(contractAddress common.Address, startTime uint64) {
	tx := c.pending.tx.tx
	c.pending.tx.factEvidenceContractDeploy = &db.FactEvidenceContract{
		TxHash:          tx.Hash(),
		ContractAddress: contractAddress,
		StartTime:       startTime,
	}
}

func (c *statsCollector) AddFactEvidenceContractCallStart(contractAddress common.Address, startBlock uint64) {
	tx := c.pending.tx.tx
	c.pending.tx.factEvidenceContractCallStart = &db.FactEvidenceContractCallStart{
		TxHash:          tx.Hash(),
		ContractAddress: contractAddress,
		StartHeight:     startBlock,
	}
}

func (c *statsCollector) AddTxReceipt(txReceipt *types.TxReceipt) {
	if txReceipt.Success {
		if c.pending.tx.factEvidenceContractDeploy != nil {
			c.stats.FactEvidenceContracts = append(c.stats.FactEvidenceContracts, c.pending.tx.factEvidenceContractDeploy)
		}
		if c.pending.tx.factEvidenceContractCallStart != nil {
			c.stats.FactEvidenceContractCallStarts = append(c.stats.FactEvidenceContractCallStarts, c.pending.tx.factEvidenceContractCallStart)
		}
	}
}

func (c *statsCollector) Disable() {
	c.statsEnabled = false
}

func (c *statsCollector) Enable() {
	c.statsEnabled = true
}

func (c *statsCollector) GetStats() *Stats {
	if !c.statsEnabled {
		return &Stats{}
	}
	return c.stats
}
