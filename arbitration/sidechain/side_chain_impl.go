package sidechain

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	. "github.com/elastos/Elastos.ELA.Arbiter/arbitration/arbitrator"
	. "github.com/elastos/Elastos.ELA.Arbiter/arbitration/base"
	"github.com/elastos/Elastos.ELA.Arbiter/common"
	"github.com/elastos/Elastos.ELA.Arbiter/common/config"
	tx "github.com/elastos/Elastos.ELA.Arbiter/core/transaction"
	"github.com/elastos/Elastos.ELA.Arbiter/core/transaction/payload"
	"github.com/elastos/Elastos.ELA.Arbiter/rpc"
	spvdb "github.com/elastos/Elastos.ELA.SPV/interface"
	spvWallet "github.com/elastos/Elastos.ELA.SPV/spvwallet"
)

type SideChainImpl struct {
	AccountListener
	key string

	currentConfig *config.SideNodeConfig
}

func (sc *SideChainImpl) GetKey() string {
	return sc.key
}

func (sc *SideChainImpl) getCurrentConfig() *config.SideNodeConfig {
	if sc.currentConfig == nil {
		for _, sideConfig := range config.Parameters.SideNodeList {
			if sc.GetKey() == sideConfig.GenesisBlockAddress {
				sc.currentConfig = sideConfig
				break
			}
		}
	}
	return sc.currentConfig
}

func (sc *SideChainImpl) GetRage() float32 {
	return sc.getCurrentConfig().Rate
}

func (sc *SideChainImpl) GetCurrentHeight() (uint32, error) {
	return rpc.GetCurrentHeight(sc.getCurrentConfig().Rpc)
}

func (sc *SideChainImpl) GetBlockByHeight(height uint32) (*BlockInfo, error) {
	return rpc.GetBlockByHeight(height, sc.getCurrentConfig().Rpc)
}

func (sc *SideChainImpl) SendTransaction(info *TransactionInfo) error {
	infoDataReader := new(bytes.Buffer)
	err := info.Serialize(infoDataReader)
	if err != nil {
		return err
	}
	content := common.BytesToHexString(infoDataReader.Bytes())

	result, err := rpc.CallAndUnmarshal("sendrawtransaction", rpc.Param("Data", content), sc.currentConfig.Rpc)
	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}

func (sc *SideChainImpl) GetAccountAddress() string {
	return sc.GetKey()
}

func (sc *SideChainImpl) OnUTXOChanged(txinfo *TransactionInfo) error {

	txn, err := txinfo.ToTransaction()
	if err != nil {
		return err
	}
	withdrawInfo, err := sc.ParseUserWithdrawTransactionInfo(txn)
	if err != nil {
		return err
	}
	for _, info := range withdrawInfo {
		currentArbitrator := ArbitratorGroupSingleton.GetCurrentArbitrator()
		if err != nil {
			return err
		}

		rateFloat := sc.GetRage()
		rate := common.Fixed64(rateFloat * 10000)
		amount := info.Amount * 10000 / rate
		withdrawTransaction, err := currentArbitrator.CreateWithdrawTransaction(
			sc.GetKey(), info.TargetAddress, amount, txinfo.Hash)
		if err != nil {
			return err
		}
		if withdrawTransaction == nil {
			return errors.New("Created an empty withdraw transaction.")
		}
		currentArbitrator.BroadcastWithdrawProposal(withdrawTransaction)
	}

	return nil
}

func (sc *SideChainImpl) CreateDepositTransaction(target string, proof spvdb.Proof, amount common.Fixed64) (*TransactionInfo, error) {
	var totalOutputAmount = amount // The total amount will be spend
	var txOutputs []TxoutputInfo   // The outputs in transaction

	assetID := spvWallet.SystemAssetId
	txOutput := TxoutputInfo{
		AssetID:    assetID.String(),
		Value:      totalOutputAmount.String(),
		Address:    target,
		OutputLock: uint32(0),
	}
	txOutputs = append(txOutputs, txOutput)

	spvInfo := new(bytes.Buffer)
	err := proof.Serialize(spvInfo)
	if err != nil {
		return nil, err
	}

	// Create payload
	txPayloadInfo := new(IssueTokenInfo)
	txPayloadInfo.Proof = common.BytesToHexString(spvInfo.Bytes())

	// Create attributes
	txAttr := TxAttributeInfo{tx.Nonce, strconv.FormatInt(rand.Int63(), 10)}
	attributesInfo := make([]TxAttributeInfo, 0)
	attributesInfo = append(attributesInfo, txAttr)

	// Create program
	program := ProgramInfo{}
	return &TransactionInfo{
		TxType:        tx.IssueToken,
		Payload:       txPayloadInfo,
		Attributes:    attributesInfo,
		UTXOInputs:    []UTXOTxInputInfo{},
		BalanceInputs: []BalanceTxInputInfo{},
		Outputs:       txOutputs,
		Programs:      []ProgramInfo{program},
		LockTime:      uint32(0),
	}, nil
}

func (sc *SideChainImpl) ParseUserWithdrawTransactionInfo(txn *tx.Transaction) ([]*WithdrawInfo, error) {

	var result []*WithdrawInfo

	switch payloadObj := txn.Payload.(type) {
	case *payload.TransferCrossChainAsset:
		for address, index := range payloadObj.AddressesMap {
			info := &WithdrawInfo{
				TargetAddress: address,
				Amount:        txn.Outputs[index].Value,
			}
			result = append(result, info)
		}
	default:
		return nil, errors.New("Invalid payload")
	}

	return result, nil
}
