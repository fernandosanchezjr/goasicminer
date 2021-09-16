package analytics

import (
	"encoding/json"
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"log"
	"os"
	"path"
	"testing"
)

func Test_Analytics_NTimeVersion(t *testing.T) {
	t.SkipNow()
	cfg, err := config.LoadConfig()
	cfg.Node.ClientOnly = true
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Node == nil {
		t.Fatal("No node in cfg file")
	}
	node := node.NewNode(cfg.Node)
	if err := node.Connect(); err != nil {
		t.Fatal(err)
	}
	info, infoErr := node.GetInfo()
	if infoErr != nil {
		t.Fatal(err)
	}
	var usedNTimes = map[utils.NTime]map[utils.Version]uint64{}
	for i := info.Blocks; i > 0; i-- {
		header, headerErr := node.GetBlockHeader(i)
		if headerErr != nil {
			t.Fatal(headerErr)
		}
		var nTime = utils.NTime(header.Timestamp.Unix()) & 0xff
		var version = utils.Version(header.Version)
		var usedNTime, found = usedNTimes[nTime]
		if !found {
			usedNTime = map[utils.Version]uint64{version: 1}
		} else {
			var usedVersion, versionFound = usedNTime[version]
			if !versionFound {
				usedNTime[version] = 1
			} else {
				usedNTime[version] = usedVersion + 1
			}
		}
		usedNTimes[nTime] = usedNTime
		if i%100 == 0 {
			log.Println(i)
		}
	}
	log.Println("Complete")
	analyticsFolder := utils.GetSubFolder("analytics")
	f, fileError := os.OpenFile(path.Join(analyticsFolder, "ntimeVersion.json"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if fileError != nil {
		t.Fatal(fileError)
	}
	var encoder = json.NewEncoder(f)
	if encoderErr := encoder.Encode(usedNTimes); encoderErr != nil {
		t.Fatal(encoderErr)
	}
	if closeErr := f.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

}

func Test_Analytics_ReadNTimeVersion(t *testing.T) {
	cfg, err := config.LoadConfig()
	cfg.Node.ClientOnly = true
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Node == nil {
		t.Fatal("No node in cfg file")
	}
	node := node.NewNode(cfg.Node)
	if err := node.Connect(); err != nil {
		t.Fatal(err)
	}
	template, templateErr := node.GetBlockTemplate()
	if templateErr != nil {
		t.Fatal(templateErr)
	}
	var unts, loadErr = LoadRawUsedNtimes()
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	unts.FilterVersions(utils.Version(template.Version))
}
