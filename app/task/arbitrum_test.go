package task

import (
	"context"
	"testing"

	"github.com/smallnest/chanx"
	"github.com/v03413/bepusdt/app/conf"
	"github.com/v03413/bepusdt/app/log"
)

func TestArbitrum(t *testing.T) {
	log.Init()
	unitTestMode = true
	ctx := context.Background()

	arb := evm{
		Type:     conf.Arbitrum,
		Endpoint: conf.GetArbitrumRpcEndpoint(),
		Block: block{
			InitStartOffset: -600,
			ConfirmedOffset: 40,
		},
		blockScanQueue: chanx.NewUnboundedChan[evmBlock](ctx, 30),
	}
	go arb.blockDispatch(ctx)

	arb.blockRoll(ctx)

}
