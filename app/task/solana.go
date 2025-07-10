package task

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/panjf2000/ants/v2"
	"github.com/shopspring/decimal"
	"github.com/smallnest/chanx"
	"github.com/tidwall/gjson"
	"github.com/v03413/bepusdt/app/conf"
	"github.com/v03413/bepusdt/app/log"
	"github.com/v03413/bepusdt/app/model"
	"io"
	"math/big"
	"time"
)

// 参考文档
//  - https://solana.com/zh/docs/rpc
//  - https://github.com/solana-program/token/blob/main/program/src/instruction.rs

type solana struct {
	slotConfirmedOffset int64
	slotInitStartOffset int64
	lastSlotNum         int64
	slotQueue           *chanx.UnboundedChan[int64]
}

var sol solana

func init() {
	sol = newSolana()
	register(task{callback: sol.slotDispatch})
	register(task{callback: sol.slotRoll, duration: time.Second * 5})
}

func newSolana() solana {
	return solana{
		slotConfirmedOffset: 60,
		slotInitStartOffset: -600,
		lastSlotNum:         0,
		slotQueue:           chanx.NewUnboundedChan[int64](context.Background(), 30),
	}
}

func (s *solana) slotRoll(context.Context) {
	if rollBreak(conf.Solana) {

		return
	}

	post := []byte(`{"jsonrpc":"2.0","id":1,"method":"getSlot"}`)

	resp, err := client.Post(conf.GetSolanaRpcEndpoint(), contentType, bytes.NewBuffer(post))
	if err != nil {
		log.Warn("slotRoll Error sending request:", err)

		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Warn("slotRoll Error response status code:", resp.StatusCode)

		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warn("slotRoll Error reading response body:", err)

		return
	}

	now := gjson.GetBytes(body, "result").Int()
	if now <= 0 {
		log.Warn("slotRoll Error: invalid slot number:", now)

		return
	}

	if conf.GetTradeIsConfirmed() {

		now = now - s.slotConfirmedOffset
	}

	if now-s.lastSlotNum > conf.BlockHeightMaxDiff { // 区块高度变化过大，强制丢块重扫
		s.lastSlotNum = now
		s.slotInitOffset(now)
	}

	if now == s.lastSlotNum { // 区块高度没有变化

		return
	}

	for n := s.lastSlotNum + 1; n <= now; n++ {
		// 待扫描区块入列

		s.slotQueue.In <- n
	}

	s.lastSlotNum = now
}

func (s *solana) slotDispatch(ctx context.Context) {
	p, err := ants.NewPoolWithFunc(3, s.slotParse)
	if err != nil {
		panic(err)

		return
	}

	defer p.Release()

	for {
		select {
		case slot := <-s.slotQueue.Out:
			if err := p.Invoke(slot); err != nil {
				s.slotQueue.In <- slot
				log.Warn("slotDispatch Error invoking process slot:", err)
			}
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				log.Warn("slotDispatch context done:", err)
			}

			return
		}
	}
}

// slotInitOffset 初始化区块高度偏移，往回扫一定数量的区块
func (s *solana) slotInitOffset(now int64) {
	if now == 0 || s.lastSlotNum != 0 {

		return
	}

	go func() {
		ticker := time.NewTicker(time.Millisecond * 300)
		defer ticker.Stop()

		for num := now; num >= now+s.slotInitStartOffset; num-- {
			if rollBreak(conf.Solana) {

				return
			}

			s.slotQueue.In <- num

			<-ticker.C
		}
	}()
}

func (s *solana) slotParse(n any) {
	slot := n.(int64)
	post := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"getBlock","params":[%d,{"encoding":"json","maxSupportedTransactionVersion":0,"transactionDetails":"full","rewards":false}]}`, slot))
	network := conf.Solana

	conf.SetBlockTotal(network)
	resp, err := client.Post(conf.GetSolanaRpcEndpoint(), contentType, bytes.NewBuffer(post))
	if err != nil {
		conf.SetBlockFail(network)
		log.Warn("slotParse Error sending request:", err)

		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		conf.SetBlockFail(network)
		log.Warn("slotParse Error response status code:", resp.StatusCode)

		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		conf.SetBlockFail(network)
		s.slotQueue.In <- slot
		log.Warn("slotParse Error reading response body:", err)

		return
	}

	timestamp := time.Unix(gjson.GetBytes(body, "result.blockTime").Int(), 0)

	for _, trans := range gjson.GetBytes(body, "result.transactions").Array() {
		hash := trans.Get("transaction.signatures.0").String()

		// 解析账号索引
		accountKeys := make([]string, 0)
		for _, key := range trans.Get("transaction.message.accountKeys").Array() {
			accountKeys = append(accountKeys, key.String())
		}
		for _, v := range []string{"readonly", "writable"} {
			for _, key := range trans.Get("meta.loadedAddresses." + v).Array() {

				accountKeys = append(accountKeys, key.String())
			}
		}

		// 查找SPL Token索引
		splTokenIndex := int64(-1)
		for i, v := range accountKeys {
			if v == conf.SolSplToken {
				splTokenIndex = int64(i)

				break
			}
		}

		// SPL Token的Mint地址，即不包含USDT交易信息
		if splTokenIndex == -1 {

			continue
		}

		// 解析 USDT Token 账户 【Token Address => Owner Address】
		usdtTokenAccountMap := make(map[string]string)
		for _, v := range []string{"postTokenBalances", "preTokenBalances"} {
			for _, itm := range trans.Get("meta." + v).Array() {
				if itm.Get("mint").String() != conf.UsdtSolana || itm.Get("programId").String() != conf.SolSplToken {

					continue
				}

				usdtTokenAccountMap[accountKeys[itm.Get("accountIndex").Int()]] = itm.Get("owner").String()
			}
		}

		transArr := make([]transfer, 0)

		// 解析外部指令
		for _, instr := range trans.Get("transaction.message.instructions").Array() {
			if instr.Get("programIdIndex").Int() != splTokenIndex {

				continue
			}

			transArr = append(transArr, s.parseTransfer(instr, accountKeys, usdtTokenAccountMap))
		}

		// 解析内部指令
		innerInstructions := trans.Get("meta.innerInstructions").Array()
		if len(innerInstructions) == 0 {

			continue
		}

		for _, itm := range innerInstructions {
			for _, instr := range itm.Get("instructions").Array() {
				if instr.Get("programIdIndex").Int() != splTokenIndex {
					// 不是SPL Token的指令，即不会包含合约代表 transfer 的指令

					continue
				}

				transArr = append(transArr, s.parseTransfer(instr, accountKeys, usdtTokenAccountMap))
			}
		}

		// 过滤无关交易
		result := make([]transfer, 0)
		for _, t := range transArr {
			if t.FromAddress == "" || t.RecvAddress == "" || t.Amount.IsZero() {

				continue
			}

			t.TxHash = hash
			t.Network = conf.Solana
			t.BlockNum = slot
			t.Timestamp = timestamp
			t.TradeType = model.OrderTradeTypeUsdtSolana

			result = append(result, t)
		}

		if len(result) > 0 {
			transferQueue.In <- result
		}
	}

	log.Info("区块扫描完成", slot, conf.GetBlockSuccRate(network), network)
}

func (s *solana) parseTransfer(instr gjson.Result, accountKeys []string, usdtTokenAccountMap map[string]string) transfer {
	accounts := instr.Get("accounts").Array()
	trans := transfer{}
	if len(accounts) < 3 { // from to singer，至少存在3个账户索引，如果是多签则 > 3

		return trans
	}

	data := base58.Decode(instr.Get("data").String())
	dLen := len(data)
	isTransfer := data[0] == 3 && dLen == 9
	isTransferChecked := data[0] == 12 && dLen == 10
	if !isTransfer && !isTransferChecked {

		return trans
	}

	var exp int32 = -6
	if isTransferChecked {
		exp = int32(data[9]) * -1
	}

	from, ok := usdtTokenAccountMap[accountKeys[accounts[0].Int()]]
	if !ok {

		return trans
	}

	trans.FromAddress = from
	trans.RecvAddress = usdtTokenAccountMap[accountKeys[accounts[1].Int()]]
	if isTransferChecked {
		trans.RecvAddress = usdtTokenAccountMap[accountKeys[accounts[2].Int()]]
	}

	buf := make([]byte, 8)
	copy(buf[:], data[1:9])
	number := binary.LittleEndian.Uint64(buf)
	b := new(big.Int)
	b.SetUint64(number)
	trans.Amount = decimal.NewFromBigInt(b, exp) // USDT的精度是6位小数

	return trans
}
