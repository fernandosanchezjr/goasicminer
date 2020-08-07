package stratum

import (
	"encoding/hex"
	"encoding/json"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"io/ioutil"
	"log"
	"testing"
)

func unmarshalFile(fileName string, value interface{}) error {
	if data, err := ioutil.ReadFile(fileName); err != nil {
		return err
	} else {
		if err := json.Unmarshal(data, value); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalTestWork() (*PoolWork, error) {
	var reply *protocol.Reply
	var sr *protocol.SubscribeResponse
	var sd *protocol.SetDifficulty
	var n *protocol.Notify
	var svm *protocol.SetVersionMask
	cf := &protocol.ConfigureResponse{}
	if err := unmarshalFile("subscribe_test.json", &reply); err != nil {
		return nil, err
	} else {
		if sr, err = protocol.NewSubscribeResponse(reply); err != nil {
			return nil, err
		}
	}
	if err := unmarshalFile("set_difficulty_test.json", &reply); err != nil {
		return nil, err
	} else {
		if sd, err = protocol.NewSetDifficulty(reply); err != nil {
			return nil, err
		}
	}
	if err := unmarshalFile("notify_test.json", &reply); err != nil {
		return nil, err
	} else {
		if n, err = protocol.NewNotify(reply); err != nil {
			return nil, err
		}
	}
	if err := unmarshalFile("set_version_mask_test.json", &reply); err != nil {
		return nil, err
	} else {
		if svm, err = protocol.NewSetVersionMask(reply); err != nil {
			return nil, err
		} else {
			cf.VersionRolling = true
			cf.VersionRollingMask = svm.VersionRollingMask
		}
	}
	pw := NewPoolWork(sr, cf, sd, n, nil)
	return pw, nil
}

func TestPoolWork(t *testing.T) {
	pw, err := unmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	coinbase := pw.Coinbase()
	hex_coinbase := hex.EncodeToString(coinbase)
	expected_coinbase := "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4a031ccb0" +
		"9fabe6d6df183ff6cbf2a1e8198b6679b7cef3e1cce0431da353154caa55e04ca3f66b3a60100000000000000002a6502002aa65f000" +
		"0000000000000939d289b2f736c7573682f000000000443ca3c26000000001976a9147c154ed1dc59609e3d26abb2df2ea3d587cd8c4" +
		"188ac00000000000000002c6a4c2952534b424c4f434b3a0ec82b00b353ab052014b472cb3ee39bb32431be99b7db757171f71600275" +
		"0190000000000000000296a4c266a24b9e11b6d8f8cc50f47dc5e8537a9e300984ee50eefd8eb7917b4b83a28287fe15e80c98200000" +
		"00000000000266a24aa21a9ed7aee68d448839eba918f66147bf31b096fe443c60175a53878c8e052cdd799f700000000"
	if hex_coinbase != expected_coinbase {
		for i := 0; i < len(hex_coinbase); i++ {
			if len(expected_coinbase) <= i {
				log.Println("Failed at character", i)
				break
			}
			if hex_coinbase[i] != expected_coinbase[i] {
				log.Println("Difference at", i, string(hex_coinbase[i]), "vs", string(expected_coinbase[i]))
				break
			}
		}
		t.Fail()
	}
	coinbase_hash := utils.DoubleHash(coinbase)
	hex_coinbase_hash := hex.EncodeToString(coinbase_hash[:])
	if hex_coinbase_hash != "2d749fc9eeea345bd91241187f92318442f48fca3c2537c242d2c6c917d7dca6" {
		t.Fail()
	}
	merkle_root := pw.MerkleRoot()
	hex_merkle_root := hex.EncodeToString(merkle_root)
	if hex_merkle_root != "8de8f457cffef502d75ada232b2e68be61724c35f48432c7d0cac77d7b1dde50" {
		t.Fail()
	}
	plain_header := pw.PlainHeader()
	hex_plain_header := hex.EncodeToString(plain_header)
	expected_plain_header := "20000000bd3e4f2c6d8b14c9d677cb428a124dcae58c5530000f791f00000000000000008de8f457cffef50" +
		"2d75ada232b2e68be61724c35f48432c7d0cac77d7b1dde505f2606591710b4f800000000"
	if hex_plain_header != expected_plain_header {
		for i := 0; i < len(hex_plain_header); i++ {
			if len(expected_plain_header) <= i {
				log.Println("Failed at character", i)
				break
			}
			if hex_plain_header[i] != expected_plain_header[i] {
				log.Println("Difference at", i, string(hex_plain_header[i]), "vs", string(expected_plain_header[i]))
				break
			}
		}
		t.Fail()
	}
	doubleHash := utils.DoubleHash(plain_header)
	if hex.EncodeToString(doubleHash[:]) != "93ce397668878d409b5a3d3aaa9ad6b58442449c3c407969750d244137ce4abd" {
		t.Fatal("invalid double hash")
	}
	versions := pw.Versions(4)
	if len(versions) != 4 {
		t.Fail()
	}
	for _, v := range versions {
		log.Printf("%x", v&pw.VersionRollingMask)
	}
	clone := pw.Clone()
	clone.Version = 0
	if clone.Version == pw.Version {
		t.Fail()
	}
}

func BenchmarkPoolWork_PlainHeader(b *testing.B) {
	pw, err := unmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pw.PlainHeader()
	}
	b.StopTimer()
}

func BenchmarkPoolWork_Versions(b *testing.B) {
	pw, err := unmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = pw.Versions(4)
	}
	b.StopTimer()
}

func BenchmarkPoolWork_Clone(b *testing.B) {
	pw, err := unmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	var clone *PoolWork
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		clone = pw.Clone()
	}
	b.StopTimer()
	// method call to avoid unused variable error
	_ = clone.String()
}

func BenchmarkPoolWork_Midstate(b *testing.B) {
	pw, err := unmarshalTestWork()
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

func BenchmarkPoolWork_DoubleHash(b *testing.B) {
	pw, err := unmarshalTestWork()
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
	// call to avoid unused variable error
	//log.Println(hex.EncodeToString(ms[:]))
}
