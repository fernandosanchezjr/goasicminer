package node

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"github.com/stevenroose/go-bitcoin-core-rpc/btcjson"
	"time"
)

func (n *Node) GetBlock(removedTransactions int) (*btcutil.Block, error) {
	if n.blockTemplate == nil {
		return nil, errors.New("no Block template available")
	}
	template := n.blockTemplate
	//coinbase, coinbaseErr := n.GenerateCoinbase(
	//	int32(template.Height),
	//	math.MaxInt64,
	//	template.CoinbaseValue,
	//)
	coinbase, coinbaseErr := n.GenerateCoinbase(
		int32(template.Height),
		utils.MaskedRandomInt64(),
		template.CoinbaseValue,
	)
	if coinbaseErr != nil {
		return nil, coinbaseErr
	}
	var rawTransactions = append([]btcjson.GetBlockTemplateResultTx{}, template.Transactions...)
	removedTransactions = utils.Min(removedTransactions, len(rawTransactions)-1)
	if removedTransactions > 0 {
		rawTransactions = rawTransactions[:len(rawTransactions)-removedTransactions]
	}
	merkleRoot, transactions, merkleErr := GetMerkleTree(coinbase, rawTransactions)

	//merkleRoot, transactions, merkleErr := GetMerkleTree(coinbase, []btcjson.GetBlockTemplateResultTx{})
	if merkleErr != nil {
		return nil, merkleErr
	}
	previousHash, previousHashErr := chainhash.NewHashFromStr(template.PreviousHash)
	if previousHashErr != nil {
		return nil, previousHashErr
	}
	var nBits uint32
	if data, err := hex.DecodeString(template.Bits); err != nil {
		return nil, err
	} else {
		nBits = binary.BigEndian.Uint32(data)
	}
	var msgBlock wire.MsgBlock
	msgBlock.Header = wire.BlockHeader{
		Version:    template.Version,
		PrevBlock:  *previousHash,
		MerkleRoot: merkleRoot,
		Timestamp:  time.Unix(int64(template.CurTime), 0),
		Bits:       nBits,
	}
	for _, tx := range transactions {
		if err := msgBlock.AddTransaction(tx.MsgTx()); err != nil {
			return nil, err
		}
	}
	block := btcutil.NewBlock(&msgBlock)
	block.SetHeight(int32(template.Height))
	return block, nil
}

func (n *Node) Submit(block *btcutil.Block) error {
	n.mtx.Lock()
	defer n.mtx.Unlock()
	if n.status == Disconnected {
		return nil
	}
	client, err := n.getClient()
	if err != nil {
		return err
	}
	n.log.WithField("height", block.Height()).Println("Submitting Block")
	return client.SubmitBlock(block, nil)
}
