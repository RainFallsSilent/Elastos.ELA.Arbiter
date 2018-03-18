package arbitratorgroup

import (
	"Elastos.ELA.Arbiter/arbitration/net"
	main "Elastos.ELA.Arbiter/arbitration/mainchain"
	side "Elastos.ELA.Arbiter/arbitration/sidechain"
	comp "Elastos.ELA.Arbiter/arbitration/complain"
	"Elastos.ELA.Arbiter/crypto"
	"Elastos.ELA.Arbiter/arbitration/base"
	"Elastos.ELA.Arbiter/common"
)

type ArbitratorMain interface {
	main.MainChain
}

type ArbitratorSide interface {
	side.SideChainManager
}

type Arbitrator interface {
	ArbitratorMain
	ArbitratorSide
	net.ArbitrationNetListener
	comp.ComplainListener

	GetArbitrationNet() net.ArbitrationNet
	GetComplainSolving() comp.ComplainSolving

	IsOnDuty() bool
	GetArbitratorGroup() ArbitratorGroup
}

type ArbitratorImpl struct {

}

func (ar *ArbitratorImpl) GetArbitrationNet() net.ArbitrationNet {
	return nil
}

func (ar *ArbitratorImpl) GetComplainSolving() comp.ComplainSolving {
	return nil
}

func (ar *ArbitratorImpl) IsOnDuty() bool {
	return true
}

func (ar *ArbitratorImpl) GetArbitratorGroup() ArbitratorGroup {
	return &ArbitratorGroupSingleton
}

func (ar *ArbitratorImpl) CreateWithdrawTransaction(withdrawBank *crypto.PublicKey, target *crypto.PublicKey) *base.TransactionInfo {
	return nil
}

func (ar *ArbitratorImpl) parseSideChainKey(uint256 *common.Uint256) *crypto.PublicKey {
	return nil
}

func (ar *ArbitratorImpl) parseUserSidePublicKey(uint256 *common.Uint256) *crypto.PublicKey {
	return nil
}

func (ar *ArbitratorImpl) OnUTXOChanged(transactionHash *common.Uint256) error {
	return nil
}

func (ar *ArbitratorImpl) IsValid(information *base.SpvInformation) (bool, error) {
	return false, nil
}

func (ar *ArbitratorImpl) GenerateSpvInformation(transaction *common.Uint256) *base.SpvInformation {
	return nil
}

func (ar *ArbitratorImpl) Add(chain side.SideChain) error {
	return nil
}

func (ar *ArbitratorImpl) Remove(key *crypto.PublicKey) error {
	return nil
}

func (ar *ArbitratorImpl) GetChain(key *crypto.PublicKey) (side.SideChain, error) {
	return nil, nil
}

func (ar *ArbitratorImpl) GetAllChains() ([]side.SideChain, error) {
	return nil, nil
}

func (ar *ArbitratorImpl) OnReceived(buf []byte, arbitrator Arbitrator) {

}

func (ar *ArbitratorImpl) OnComplainFeedback([]byte) {

}