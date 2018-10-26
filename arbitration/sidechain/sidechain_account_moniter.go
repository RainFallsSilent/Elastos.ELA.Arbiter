package sidechain

import (
	"errors"
	"sync"
	"time"

	"github.com/elastos/Elastos.ELA.Arbiter/arbitration/arbitrator"
	. "github.com/elastos/Elastos.ELA.Arbiter/arbitration/base"
	"github.com/elastos/Elastos.ELA.Arbiter/config"
	"github.com/elastos/Elastos.ELA.Arbiter/log"
	. "github.com/elastos/Elastos.ELA.Arbiter/rpc"
	"github.com/elastos/Elastos.ELA.Arbiter/store"
	"github.com/elastos/Elastos.ELA.Utility/common"
)

type SideChainAccountMonitorImpl struct {
	mux sync.Mutex

	ParentArbitrator   arbitrator.Arbitrator
	accountListenerMap map[string]AccountListener
}

func (monitor *SideChainAccountMonitorImpl) tryInit() {
	if monitor.accountListenerMap == nil {
		monitor.accountListenerMap = make(map[string]AccountListener)
	}
}

func (monitor *SideChainAccountMonitorImpl) AddListener(listener AccountListener) {
	monitor.tryInit()
	monitor.accountListenerMap[listener.GetAccountAddress()] = listener
}

func (monitor *SideChainAccountMonitorImpl) RemoveListener(account string) error {
	if monitor.accountListenerMap == nil {
		return nil
	}

	if _, ok := monitor.accountListenerMap[account]; !ok {
		return errors.New("Do not exist listener.")
	}
	delete(monitor.accountListenerMap, account)
	return nil
}

func (monitor *SideChainAccountMonitorImpl) fireUTXOChanged(txinfos []*TransactionInfo, genesisBlockAddress string, blockHeight uint32) error {
	if monitor.accountListenerMap == nil {
		return nil
	}

	item, ok := monitor.accountListenerMap[genesisBlockAddress]
	if !ok {
		return errors.New("Fired unknown listener.")
	}

	return item.OnUTXOChanged(txinfos, blockHeight)
}

func (monitor *SideChainAccountMonitorImpl) SyncChainData(sideNode *config.SideNodeConfig) {
	for {
		chainHeight, currentHeight, needSync := monitor.needSyncBlocks(sideNode.GenesisBlockAddress, sideNode.Rpc)

		if needSync {
			log.Info("currentHeight:", currentHeight, " chainHeight:", chainHeight)
			for currentHeight < chainHeight {
				if currentHeight >= 6 {
					transactions, err := GetDestroyedTransactionByHeight(currentHeight+1-6, sideNode.Rpc)
					if err != nil {
						log.Error("Get destoryed transaction at height:", currentHeight+1-6, "failed\n"+
							"rpc:", sideNode.Rpc.IpAddress, ":", sideNode.Rpc.HttpJsonPort, "\n"+
							"error:", err)
						break
					}
					monitor.processTransactions(transactions, sideNode.GenesisBlockAddress, currentHeight+1-6)
				}
				// Update wallet height
				currentHeight = store.DbCache.SideChainStore.CurrentSideHeight(sideNode.GenesisBlockAddress, currentHeight+1)
				log.Info(" [SyncSideChain] Side chain [", sideNode.GenesisBlockAddress, "] height: ", currentHeight)
			}

			arbitrator.ArbitratorGroupSingleton.SyncFromMainNode()
			if arbitrator.ArbitratorGroupSingleton.GetCurrentArbitrator().IsOnDutyOfMain() {
				sideChain, ok := arbitrator.ArbitratorGroupSingleton.GetCurrentArbitrator().GetSideChainManager().GetChain(sideNode.GenesisBlockAddress)
				if ok {
					sideChain.StartSideChainMining()
					log.Info("[SyncSideChain] Start side chain mining, genesis address: [", sideNode.GenesisBlockAddress, "]")
				}
			}
		}

		time.Sleep(time.Millisecond * config.Parameters.SideChainMonitorScanInterval)
	}
}

func (monitor *SideChainAccountMonitorImpl) needSyncBlocks(genesisBlockAddress string, config *config.RpcConfig) (uint32, uint32, bool) {

	chainHeight, err := GetCurrentHeight(config)
	if err != nil {
		return 0, 0, false
	}

	currentHeight := store.DbCache.SideChainStore.CurrentSideHeight(genesisBlockAddress, store.QueryHeightCode)

	if currentHeight >= chainHeight {
		return chainHeight, currentHeight, false
	}

	return chainHeight, currentHeight, true
}

func (monitor *SideChainAccountMonitorImpl) processTransactions(transactions *BlockTransactions, genesisAddress string, blockHeight uint32) {
	var txInfos []*TransactionInfo
	for _, txn := range transactions.Transactions {
		for _, output := range txn.Outputs {
			if output.Address == DESTROY_ADDRESS {
				txnBytes, err := common.HexStringToBytes(txn.Hash)
				if err != nil {
					log.Warn("Find output to destroy address, but transaction hash to transaction bytes failed")
					continue
				}
				reversedTxnBytes := common.BytesReverse(txnBytes)
				reversedTxnHash := common.BytesToHexString(reversedTxnBytes)
				txn.Hash = reversedTxnHash
				if ok, err := store.DbCache.SideChainStore.HasSideChainTx(txn.Hash); err != nil || !ok {
					txInfos = append(txInfos, txn)
					break
				}
			}
		}
	}
	if len(txInfos) != 0 {
		monitor.fireUTXOChanged(txInfos, genesisAddress, blockHeight)
	}
}
