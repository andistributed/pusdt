package monitor

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"github.com/v03413/bepusdt/app/config"
	"github.com/v03413/bepusdt/app/help"
	"github.com/v03413/bepusdt/app/log"
	"github.com/v03413/bepusdt/app/model"
	"github.com/v03413/bepusdt/app/notify"
	"github.com/v03413/bepusdt/app/telegram"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const tronScanApi = "https://apilist.tronscanapi.com/"
const usdtToken = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"

func TradeStart() {
	log.Info("交易监控启动.")

	for range time.Tick(time.Second * 5) {
		var _lock, err = getAllPendingOrders()
		if err != nil {
			log.Error(err.Error())

			continue
		}

		for _, _row := range model.GetAvailableAddress() {
			var result, err = searchTransaction(_row.Address)
			if err != nil {
				log.Error(err.Error())

				continue
			}

			recentTransferTotal := result.Get("total").Num
			log.Info(fmt.Sprintf("recent transfer total: %s(%v)", _row.Address, recentTransferTotal))
			if recentTransferTotal <= 0 { // 没有交易记录

				continue
			}

			handlePaymentTransaction(_lock, _row.Address, result)
			handleOtherNotify(_row.Address, result)
		}
	}
}

// 列出所有等待支付的交易订单
func getAllPendingOrders() (map[string]model.TradeOrders, error) {
	tradeOrders, err := model.GetTradeOrderByStatus(model.OrderStatusWaiting)
	if err != nil {

		return nil, fmt.Errorf("待支付订单获取失败: %w", err)
	}

	var _lock = make(map[string]model.TradeOrders) // 当前所有正在等待支付的订单 Lock Key
	for _, order := range tradeOrders {
		if time.Now().Unix() >= order.ExpiredAt.Unix() { // 订单过期
			err := order.OrderSetExpired()
			if err != nil {
				log.Error("订单过期标记失败：", err, order.OrderId)
			} else {
				log.Info("订单过期：", order.OrderId)
			}

			continue
		}

		_lock[order.Address+order.Amount] = order
	}

	return _lock, nil
}

// 处理支付交易
func handlePaymentTransaction(_lock map[string]model.TradeOrders, _toAddress string, _data gjson.Result) {
	for _, transfer := range _data.Get("data").Array() {
		if transfer.Get("to_address").String() != _toAddress {
			// 不是接收地址

			continue
		}

		// 计算交易金额
		var _amount = parseTransAmount(transfer.Get("amount").Float())

		_order, ok := _lock[_toAddress+_amount]
		if !ok || transfer.Get("contractRet").String() != "SUCCESS" {
			// 订单不存在或交易失败

			continue
		}

		// 判断时间是否有效
		var _createdAt = time.UnixMilli(transfer.Get("date_created").Int())
		if _createdAt.Unix() < _order.CreatedAt.Unix() || _createdAt.Unix() > _order.ExpiredAt.Unix() {
			// 失效交易

			continue
		}

		// 判断交易是否需要等待广播确认
		var _confirmed = transfer.Get("confirmed").Bool()
		var _tradeHash = transfer.Get("hash").String()
		var _tradeIsConfirmed = config.GetTradeConfirmed()
		var _fromAddress = transfer.Get("from_address").String()

		if (_tradeIsConfirmed && _confirmed) || !_tradeIsConfirmed {
			if _order.OrderSetSucc(_fromAddress, _tradeHash, _createdAt) == nil {
				// 通知订单支付成功
				go notify.OrderNotify(_order)

				// TG发送订单信息
				go telegram.SendTradeSuccMsg(_order)
			}
		}
	}
}

// 非订单交易通知
func handleOtherNotify(_toAddress string, result gjson.Result) {
	for _, transfer := range result.Get("data").Array() {
		if !model.GetOtherNotify(_toAddress) {

			break
		}

		var _amount = parseTransAmount(transfer.Get("amount").Float())
		var _created = time.UnixMilli(transfer.Get("date_created").Int())
		var _txid = transfer.Get("hash").String()
		var _detailUrl = "https://tronscan.org/#/transaction/" + _txid
		if !model.IsNeedNotifyByTxid(_txid) {
			// 不需要额外通知

			continue
		}

		var title = "收入"
		if transfer.Get("to_address").String() != _toAddress {
			title = "支出"
		}

		var text = fmt.Sprintf(
			"#账户%s #非订单交易\n---\n```\n💲交易数额：%v USDT.TRC20\n⏱️交易时间：%v\n✅接收地址：%v\n🅾️发送地址：%v```\n",
			title,
			_amount,
			_created.Format(time.DateTime),
			help.MaskAddress(transfer.Get("to_address").String()),
			help.MaskAddress(transfer.Get("from_address").String()),
		)

		var chatId, err = strconv.ParseInt(config.GetTgBotNotifyTarget(), 10, 64)
		if err != nil {

			continue
		}

		var msg = tgbotapi.NewMessage(chatId, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonURL("📝查看交易明细", _detailUrl),
				},
			},
		}

		var _record = model.NotifyRecord{Txid: _txid}
		model.DB.Create(&_record)

		go telegram.SendMsg(msg)
	}
}

// 搜索交易记录
func searchTransaction(_toAddress string) (gjson.Result, error) {
	var client = &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest("GET", tronScanApi+"api/multi/search", nil)
	if err != nil {

		return gjson.Result{}, fmt.Errorf("处理请求创建错误: %w", err)
	}

	// 构建请求参数
	req.URL.RawQuery = buildSearchParams(_toAddress)

	// 请求交易记录
	resp, err := client.Do(req)
	if err != nil {

		return gjson.Result{}, fmt.Errorf("请求交易记录错误: %w", err)
	}

	// 获取响应记录
	all, err := io.ReadAll(resp.Body)
	if err != nil {

		return gjson.Result{}, fmt.Errorf("读取交易记录错误: %w", err)
	}

	// 释放响应请求
	_ = resp.Body.Close()

	// 解析响应记录
	return gjson.ParseBytes(all), nil
}

// 构建搜索参数
func buildSearchParams(toAddress string) string {
	var params = url.Values{}
	var now = time.Now()
	var start = now.Add(-time.Hour) // 当前时间向前推 3 小时
	var end = now.Add(time.Hour)    // 当前时间向后推 1 小时

	params.Add("limit", "50")
	params.Add("start", "0")
	params.Add("type", "transfer")
	params.Add("secondType", "20")
	params.Add("start_timestamp", strconv.FormatInt(start.UnixMilli(), 10)) // 起始时间
	params.Add("end_timestamp", strconv.FormatInt(end.UnixMilli(), 10))     // 截止时间
	params.Add("toAddress", toAddress)                                      // 接收地址
	params.Add("fromAddress", toAddress)                                    // 发送地址
	params.Add("token", usdtToken)                                          // USDT 通证
	params.Add("relation", "or")

	return params.Encode()
}

// 解析交易金额
func parseTransAmount(amount float64) string {
	var _decimalAmount = decimal.NewFromFloat(amount)
	var _decimalDivisor = decimal.NewFromFloat(1000000)
	return _decimalAmount.Div(_decimalDivisor).String()
}
