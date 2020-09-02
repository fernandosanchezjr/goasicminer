package stratum

import (
	"encoding/json"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
)

var subscribeExample = "{\"id\":1,\"result\":[[[\"mining.set_difficulty\",\"1\"],[\"mining.notify\",\"1\"]]," +
	"\"2c65030277fdb3\",8],\"error\":null}"
var setDifficultyExample = "{\"id\":null,\"method\":\"mining.set_difficulty\",\"params\":[512]}"
var notifyExample = "{\"id\":null,\"method\":\"mining.notify\",\"params\":[\"670b4c525\",\"ea2bc5140f45747839fce96b74bafe832804ed98000c81720000000000000000\",\"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4b03a7db09fabe6d6db63ccf784702ba45dd61cbb11d131ec33e62fd16e388467d4a2544b331fe7cbf0100000000000000\",\"25c5b470062f736c7573682f00000000041387a126000000001976a9147c154ed1dc59609e3d26abb2df2ea3d587cd8c4188ac00000000000000002c6a4c2952534b424c4f434b3ab73f1400b5697194e9e6a0523919071738f3a64cfe9f2d9d29289d25002894330000000000000000296a4c266a24b9e11b6d6f56392d1ff3101bfff34a5ec7ba38797e6317fcea6a3cfd3195c43e52cb22830000000000000000266a24aa21a9eda32de84b30634cd207648e108480f8a4975d6fab461757cb93023c9b566e082000000000\",[\"a1418e505588c28c81a6bf25e45266d6bea31317148df2f6f14b1f0099a8cb0a\",\"77ea80020b774ee28bf61c46f7c3a1d09da21fbd7566fcb0561c7e0d23eb8bdc\",\"b77725351418639da840d85f9a8a2818d7d06d953d47cb98855c23d71bf5dddb\",\"39ad1032bc10efdf9dec376602dc185990d091b40f866db0816f992b9d5a21b4\",\"5800d0da2834472aca4b969060afd182afb635b9c0f4b4af5a7893c5c51ffd7b\",\"a2741e286758116693ad7adf6c5cb43602a9229c0c33882280952c1a7f86fa56\",\"aaf8da527b487e079b11384ef5745d44457c5bd253352ba9675ebccad3ca42bd\",\"208d8e571c7a8edf4a444b02a6257a0664f028db47bd7d29f9c111192043df03\",\"26e1245cfa3c69fe064bb538147fe20c4f43b368df17df03f78d23759dc23624\",\"b5bd185d4aff198d591553d48d7f297c154aeffc74801a7032091aaf95694f58\",\"d6eae325ea4b01ca070e4e2551dabe5fd22f495472df4fd04ce56117735f6954\"],\"20000000\",\"171007ea\",\"5f4c4275\",false]}"
var setVersionMaskTest = "{\"id\":null,\"method\":\"mining.set_version_mask\",\"params\":[\"1fffe000\"]}"

func UnmarshalTestWork() (*Work, error) {
	var reply *protocol.Reply
	var sr *protocol.SubscribeResponse
	var sd *protocol.SetDifficulty
	var n *protocol.Notify
	var svm *protocol.SetVersionMask
	cf := &protocol.ConfigureResponse{}
	if err := json.Unmarshal([]byte(subscribeExample), &reply); err != nil {
		return nil, err
	} else {
		if sr, err = protocol.NewSubscribeResponse(reply); err != nil {
			return nil, err
		}
	}
	if err := json.Unmarshal([]byte(setDifficultyExample), &reply); err != nil {
		return nil, err
	} else {
		if sd, err = protocol.NewSetDifficulty(reply); err != nil {
			return nil, err
		}
	}
	if err := json.Unmarshal([]byte(notifyExample), &reply); err != nil {
		return nil, err
	} else {
		if n, err = protocol.NewNotify(reply); err != nil {
			return nil, err
		}
	}
	if err := json.Unmarshal([]byte(setVersionMaskTest), &reply); err != nil {
		return nil, err
	} else {
		if svm, err = protocol.NewSetVersionMask(reply); err != nil {
			return nil, err
		} else {
			cf.VersionRolling = true
			cf.VersionRollingMask = svm.VersionRollingMask
		}
	}
	pw := NewWork(sr, cf, sd, n, nil)
	return pw, nil
}
