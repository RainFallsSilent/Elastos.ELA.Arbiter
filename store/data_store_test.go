package store

import (
	"os"
	"testing"

	"github.com/elastos/Elastos.ELA.Arbiter/config"
)

func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

func setup() {
	config.InitMockConfig()
}

func TestDataStoreImpl_AddSideChainTx(t *testing.T) {
	datastore, err := OpenDataStore()
	if err != nil {
		t.Error("Open database error.")
	}

	genesisBlockAddress := "testAddress"
	txHash := "testHash"

	ok, err := datastore.HashSideChainTx(txHash)
	if err != nil {
		t.Error("Get side chain transaction error.")
	}
	if ok {
		t.Error("Should not have specified transaction.")
	}

	if err := datastore.AddSideChainTx(txHash, genesisBlockAddress); err != nil {
		t.Error("Add side chain transaction error.")
	}

	ok, err = datastore.HashSideChainTx(txHash)
	if err != nil {
		t.Error("Get side chain transaction error.")
	}
	if !ok {
		t.Error("Should have specified transaction.")
	}

	datastore.ResetDataStore()
}