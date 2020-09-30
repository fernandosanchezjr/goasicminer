package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/epiclabs-io/elastic"
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type SetVersionMask struct {
	VersionRollingMask utils.Version
}

func NewSetVersionMask(reply *Reply) (*SetVersionMask, error) {
	svm := &SetVersionMask{}
	if err := reply.HasError(); err != nil {
		return nil, err
	}
	if len(reply.Params) != 1 {
		return nil, errors.New("invalid SetVersionMask parameters")
	}
	var versionRollingMask string
	if err := elastic.Set(&versionRollingMask, reply.Params[0]); err != nil {
		return nil, err
	}
	if data, err := hex.DecodeString(versionRollingMask); err != nil {
		return nil, err
	} else {
		svm.VersionRollingMask = utils.Version(binary.BigEndian.Uint32(data))
	}
	return svm, nil
}
