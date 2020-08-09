package stratum

import (
	"encoding/json"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
)

var subscribeExample = "{\"id\":1,\"method\":\"\",\"params\":null,\"result\":[[[\"mining.set_difficulty\",\"1\"]," +
	"[\"mining.notify\",\"1\"]],\"2a6502002aa65f\",8],\"error\":null}"
var setDifficultyExample = "{\"id\":0,\"method\":\"mining.set_difficulty\",\"params\":[8192],\"result\":null," +
	"\"error\":null}"
var notifyExample = "{\"id\":0,\"method\":\"mining.notify\",\"params\":[\"9b289d93\",\"bd3e4f2c6d8b14c9d677cb428a124d" +
	"cae58c5530000f791f0000000000000000\",\"0100000001000000000000000000000000000000000000000000000000000000000000000" +
	"0ffffffff4a031ccb09fabe6d6df183ff6cbf2a1e8198b6679b7cef3e1cce0431da353154caa55e04ca3f66b3a60100000000000000\"," +
	"\"939d289b2f736c7573682f000000000443ca3c26000000001976a9147c154ed1dc59609e3d26abb2df2ea3d587cd8c4188ac0000000000" +
	"0000002c6a4c2952534b424c4f434b3a0ec82b00b353ab052014b472cb3ee39bb32431be99b7db757171f716002750190000000000000000" +
	"296a4c266a24b9e11b6d8f8cc50f47dc5e8537a9e300984ee50eefd8eb7917b4b83a28287fe15e80c9820000000000000000266a24aa21a9" +
	"ed7aee68d448839eba918f66147bf31b096fe443c60175a53878c8e052cdd799f700000000\",[\"f3dbd0071549db620a9e0969e54d9bec" +
	"3d22093817b48e7d1cb0e02edbf698d4\",\"c6e9ebbf95ac8ec33d0349f6a05b71e428a4bf3ce74e1e2b1774504bc1f68f39\",\"be214b" +
	"bc6c1b82ddc69c440bdfaaa63e5be2760cc9327868f0336050f77b0928\",\"08d41fb5297c248b58682ea7e967ddbd712bd3985113a2c55" +
	"6a57ef464b604fc\",\"c95ef4a3995bfe6728e7214a720851d4d14ebee3518d2d4aa6b026f44e16413d\",\"1a589249b64cae830be976d" +
	"537ee5d7cf9776cc8589d5c83d0dbad909f0fa002\",\"3903aeea64743ca14bb2fae0f82d61b116f6b048d063e978102e2a1c015b6501\"" +
	",\"9581e992663090ced5c5cbbef8cd95028f804c45a6361ecff9ec699b31922573\",\"7de724b8c18fa8030979d89aa3adc5dc08f59132" +
	"005b1b39173e05bf88b4fe3e\",\"b309c157d95cf434627e7f0ea5d87d4cbf287cb4a06fbc55609db4e3b7bbe3d0\",\"77f11484dbd994" +
	"09257933d3da53b9b742223239378f290bf051a2ee4521fde0\"],\"20000000\",\"1710b4f8\",\"5f260659\",true],\"result" +
	"\":null,\"error\":null}"
var setVersionMaskTest = "{\"id\":0,\"method\":\"mining.set_version_mask\",\"params\":[\"1fffe000\"],\"result\":null," +
	"\"error\":null}"

func UnmarshalTestWork() (*PoolWork, error) {
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
	pw := NewPoolWork(sr, cf, sd, n, nil)
	return pw, nil
}
