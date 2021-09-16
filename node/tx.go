package node

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/stevenroose/go-bitcoin-core-rpc/btcjson"
)

func ToMsgTx(hexData string) (*wire.MsgTx, error) {
	tx := &wire.MsgTx{}
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return nil, err
	}
	deserializeErr := tx.Deserialize(bytes.NewBuffer(data))
	if deserializeErr != nil {
		return nil, deserializeErr
	}
	return tx, nil
}

func MsgTxToString(tx *wire.MsgTx) (string, error) {
	buf := &bytes.Buffer{}
	if bufErr := tx.Serialize(buf); bufErr != nil {
		return "", bufErr
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func GetMerkleTree(
	coinbase *btcutil.Tx,
	sourceTxns []btcjson.GetBlockTemplateResultTx,
) (chainhash.Hash, []*btcutil.Tx, error) {
	transactions := make([]*btcutil.Tx, 0, len(sourceTxns)+1)
	transactions = append(transactions, coinbase)
	for _, templateTx := range sourceTxns {
		msgTx, msgTxErr := ToMsgTx(templateTx.Data)
		if msgTxErr != nil {
			return chainhash.Hash{}, nil, msgTxErr
		}
		transactions = append(transactions, btcutil.NewTx(msgTx))
	}
	merkles := blockchain.BuildMerkleTreeStore(transactions, false)
	return *merkles[len(merkles)-1], transactions, nil
}
