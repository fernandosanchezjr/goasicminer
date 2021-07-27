package generators

import "flag"

var ReuseExtraNonce2 = false

func init() {
	flag.BoolVar(&ReuseExtraNonce2, "reuse-extra-nonce", ReuseExtraNonce2, "reuse extra nonce and find highest difficulty")
}
