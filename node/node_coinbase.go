package node

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

const (
	// CoinbaseFlags is added to the coinbase script of a generated Block
	// and is used to monitor BIP16 support as well as blocks that are
	// generated via btcd.
	CoinbaseFlags = "/P2SH/btcd/"
)

func (n *Node) GenerateCoinbaseScript(nextBlockHeight int32, extraNonce int64) ([]byte, error) {
	return txscript.NewScriptBuilder().AddInt64(int64(nextBlockHeight)).
		AddInt64(extraNonce).AddData([]byte(CoinbaseFlags)).
		Script()
}

func (n *Node) GenerateCoinbase(nextBlockHeight int32, extraNonce int64, coinbaseValue int64) (*btcutil.Tx, error) {
	script, scriptErr := n.GenerateCoinbaseScript(nextBlockHeight, extraNonce)
	if scriptErr != nil {
		return nil, scriptErr
	}
	pkScript, pkScriptErr := txscript.PayToAddrScript(n.walletAddress)
	if pkScriptErr != nil {
		return nil, pkScriptErr
	}
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(&wire.TxIn{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOutPoint: *wire.NewOutPoint(&chainhash.Hash{},
			wire.MaxPrevOutIndex),
		SignatureScript: script,
		Sequence:        wire.MaxTxInSequenceNum,
	})
	tx.AddTxOut(&wire.TxOut{
		Value:    coinbaseValue,
		PkScript: pkScript,
	})
	return btcutil.NewTx(tx), nil
}
