package notify

import (
	"testing"

	"github.com/v03413/bepusdt/app/help"
)

func TestSign(t *testing.T) {
	req := EpNotify{
		TradeId:            `abc`,
		OrderId:            `def`,
		Amount:             200.12,
		ActualAmount:       200.12,
		Token:              `0xabcdefghijkmln`,
		BlockTransactionId: `uvwxyz`,
		Signature:          ``,
		Status:             2,
		Nonce:              `iajfiefioahfoeflfbjpftnaoeof`,
	}
	data := req.ToMap()
	sign := help.EpusdtSign(data, `testsecret`)
	t.Logf(`sign: %s`, sign)
}
