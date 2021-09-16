module github.com/fernandosanchezjr/goasicminer

go 1.14

replace github.com/ziutek/ftdi => ../ftdi

replace github.com/stevenroose/go-bitcoin-core-rpc => ../go-bitcoin-core-rpc

require (
	github.com/ReneKroon/ttlcache v1.7.0
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/epiclabs-io/elastic v0.0.0-20200226000247-178868363452
	github.com/fernandosanchezjr/sha256-simd v0.1.4
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-echarts/go-echarts v1.0.0
	github.com/howeyc/crc16 v0.0.0-20171223171357-2b2a61e366a6
	github.com/julienschmidt/httprouter v1.3.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stevenroose/go-bitcoin-core-rpc v0.0.0-20181021223752-1f5e57e12ba1 // indirect
	github.com/stianeikeland/go-rpio/v4 v4.4.1-0.20200705092735-acc952dac3eb
	github.com/valyala/gorpc v0.0.0-20160519171614-908281bef774
	github.com/ziutek/ftdi v0.0.2
	go.etcd.io/bbolt v1.3.5
	gonum.org/v1/gonum v0.8.1
	gopkg.in/yaml.v2 v2.3.0
)
