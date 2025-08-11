package task

import (
	"context"
	"os"
	"testing"

	"github.com/smallnest/chanx"
	"github.com/v03413/bepusdt/app/conf"
	"github.com/v03413/bepusdt/app/log"
)

func TestArbitrum(t *testing.T) {
	os.Setenv(`BEPUSDT_LOG_OUTPUT_CONSOLE`, `1`)
	log.Init()
	unitTestMode = true
	ctx := context.Background()

	arb := evm{
		Type:     conf.Arbitrum,
		Endpoint: conf.GetArbitrumRpcEndpoint(),
		Block: block{
			//InitStartOffset: -600,
			ConfirmedOffset: 40,
		},
		blockScanQueue: chanx.NewUnboundedChan[evmBlock](ctx, 30),
		debug:          true,
	}
	go arb.blockDispatch(ctx)

	arb.blockRoll(ctx)

	//arb.getBlockByNumber(evmBlock{From: 367218642, To: 367218642})

}
