package stratum

import (
	"encoding/hex"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"log"
	"testing"
)

func TestWork(t *testing.T) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pw.SetExtraNonce2(4)
	coinbase := pw.Coinbase()
	hexCoinbase := hex.EncodeToString(coinbase)
	expectedCoinbase := "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4b03a7db09fabe6d6db63ccf784702ba45dd61cbb11d131ec33e62fd16e388467d4a2544b331fe7cbf01000000000000002c65030277fdb3040000000000000025c5b470062f736c7573682f00000000041387a126000000001976a9147c154ed1dc59609e3d26abb2df2ea3d587cd8c4188ac00000000000000002c6a4c2952534b424c4f434b3ab73f1400b5697194e9e6a0523919071738f3a64cfe9f2d9d29289d25002894330000000000000000296a4c266a24b9e11b6d6f56392d1ff3101bfff34a5ec7ba38797e6317fcea6a3cfd3195c43e52cb22830000000000000000266a24aa21a9eda32de84b30634cd207648e108480f8a4975d6fab461757cb93023c9b566e082000000000"
	if hexCoinbase != expectedCoinbase {
		for i := 0; i < len(hexCoinbase); i++ {
			if len(expectedCoinbase) <= i {
				log.Println("Failed at character", i)
				break
			}
			if hexCoinbase[i] != expectedCoinbase[i] {
				log.Println("Difference at", i, string(hexCoinbase[i]), "vs", string(expectedCoinbase[i]))
				break
			}
		}
		t.Fatal(hexCoinbase)
	}
	coinbaseHash := utils.DoubleHash(coinbase)
	hexCoinbaseHash := hex.EncodeToString(coinbaseHash[:])
	if hexCoinbaseHash != "07763980e4e1f1dc66632da08bcd3beb19fb871af1cdd6ae9bf19b4e624b9caf" {
		t.Fatal(hexCoinbaseHash)
	}
	merkleRoot := pw.MerkleRoot()
	hexMerkleRoot := hex.EncodeToString(merkleRoot)
	if hexMerkleRoot != "0ba65b67c3e86a8021a899f4cb1872e000b70fc61c565f935c2d7bfca03faa69" {
		t.Fatal(hexMerkleRoot)
	}
	plainHeader := pw.PlainHeader()
	hexPlainHeader := hex.EncodeToString(plainHeader)
	expectedPlainHeader := "20000000ea2bc5140f45747839fce96b74bafe832804ed98000c817200000000000000000ba65b67c3e86a8021a899f4cb1872e000b70fc61c565f935c2d7bfca03faa695f4c4275171007ea00000000"
	if hexPlainHeader != expectedPlainHeader {
		for i := 0; i < len(hexPlainHeader); i++ {
			if len(expectedPlainHeader) <= i {
				log.Println("Failed at character", i)
				break
			}
			if hexPlainHeader[i] != expectedPlainHeader[i] {
				log.Println("Difference at", i, string(hexPlainHeader[i]), "vs", string(expectedPlainHeader[i]))
				break
			}
		}
		t.Fatal(hexPlainHeader)
	}
	doubleHash := utils.DoubleHash(plainHeader)
	if hex.EncodeToString(doubleHash[:]) != "71ebc0d2147cb602520ed0a7286bc2aa9df9c3e329489c7067738c6e877e814d" {
		t.Fatal(hex.EncodeToString(doubleHash[:]))
	}
	versions := pw.Versions(4)
	if len(versions) != 4 {
		t.Fatal()
	}
	clone := pw.Clone()
	clone.Version = 0
	if clone.Version == pw.Version {
		t.Fatal()
	}
	pw.SetExtraNonce2(5)
	if hex.EncodeToString(pw.PlainHeader()) != "20000000ea2bc5140f45747839fce96b74bafe832804ed98000c8172000000000000000035fa2ab24ca9ffcda72268bb459859ddb4af1da7d7d347415c6a3d49dc82438b5f4c4275171007ea00000000" {
		t.Fatal(hex.EncodeToString(pw.PlainHeader()))
	}
}

func BenchmarkWork_PlainHeader(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pw.PlainHeader()
	}
	b.StopTimer()
}

func BenchmarkWork_Versions(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = pw.Versions(4)
	}
	b.StopTimer()
}

func BenchmarkWork_Clone(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	var clone *Work
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		clone = pw.Clone()
	}
	b.StopTimer()
	// method call to avoid unused variable error
	_ = clone.String()
}

func BenchmarkWork_Midstate(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	header := pw.PlainHeader()
	var ms [32]byte
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ms = utils.Midstate(header[:64])
	}
	b.StopTimer()
	// call to avoid unused variable error
	_ = ms
}

func BenchmarkWork_DoubleHash(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	header := pw.PlainHeader()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ms := utils.DoubleHash(header)
		_ = ms
	}
	b.StopTimer()
}
