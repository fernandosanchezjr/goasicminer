package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/epiclabs-io/elastic"
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type ConfigureResponse struct {
	VersionRolling     bool
	VersionRollingMask utils.Version
}

func NewConfigureResponse(reply *Reply) (*ConfigureResponse, error) {
	cr := &ConfigureResponse{}
	if err := reply.HasError(); err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := elastic.Set(&result, reply.Result); err != nil {
		return nil, err
	}
	if err := elastic.Set(&cr.VersionRolling, result["version-rolling"]); err != nil {
		return nil, err
	}
	if !cr.VersionRolling {
		return cr, nil
	}
	var versionRollingMask string
	if err := elastic.Set(&versionRollingMask, result["version-rolling.mask"]); err != nil {
		return nil, err
	}
	if data, err := hex.DecodeString(versionRollingMask); err != nil {
		return nil, err
	} else {
		cr.VersionRollingMask = utils.Version(binary.BigEndian.Uint32(data))
	}
	return cr, nil
}
