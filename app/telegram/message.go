package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/v03413/bepusdt/app/config"
	"github.com/v03413/bepusdt/app/model"
	"strconv"
	"strings"
	"time"
)

func SendTradeSuccMsg(order model.TradeOrders) {
	var adminChatId, err = strconv.ParseInt(config.GetTGBotAdminId(), 10, 64)
	if err != nil {

		return
	}
	var text = `
✅有新的交易支付成功
---
🚦商户订单：｜%v｜
💰请求金额：｜%v｜ CNY(%v)
💲支付数额：%v USDT.TRC20
🪧收款地址：｜%s｜
⏱️创建时间：%s
️🎯️支付时间：%s
`
	text = fmt.Sprintf(strings.ReplaceAll(text, "｜", "`"), order.OrderId, order.Money, order.UsdtRate, order.Amount, order.Address,
		order.CreatedAt.Format(time.DateTime), order.UpdatedAt.Format(time.DateTime))
	var msg = tgbotapi.NewMessage(adminChatId, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonURL("📝查看交易明细", "https://tronscan.org/#/transaction/"+order.TradeHash),
			},
		},
	}

	_, _ = botApi.Send(msg)
}

func SendOtherNotify(text string) {
	var adminChatId, err = strconv.ParseInt(config.GetTGBotAdminId(), 10, 64)
	if err != nil {

		return
	}

	var msg = tgbotapi.NewMessage(adminChatId, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	_, _ = botApi.Send(msg)
}

func SendWelcome(version string) {
	var text = `
👋 欢迎使用 Bepusdt，一款更好用的个人USDT收款网关，如果您看到此消息，说明机器人已经启动成功

📌当前版本：` + version + `
📝发送命令 /start 可以开始使用
🎉开源地址 https://github.com/v03413/bepusdt
---
`
	var msg = tgbotapi.NewMessage(0, text)

	SendMsg(msg)
}
