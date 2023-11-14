package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/v03413/bepusdt/app/model"
	"github.com/v03413/bepusdt/app/usdt"
)

const cmdGetId = "id"
const cmdStart = "start"
const cmdUsdt = "usdt"

const replayAddressText = "🚚 请发送一个合法的钱包地址"

func cmdGetIdHandle(_msg *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(_msg.Chat.ID, "您的ID: "+fmt.Sprintf("`%v`(点击复制)", _msg.Chat.ID))
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyToMessageID = _msg.MessageID
	_, _ = botApi.Send(msg)
}

func cmdStartHandle() {
	var msg = tgbotapi.NewMessage(0, "请点击钱包地址按照提示进行操作")
	var was []model.WalletAddress
	var inlineBtn [][]tgbotapi.InlineKeyboardButton
	if model.DB.Find(&was).Error == nil {
		for _, wa := range was {
			var _address = fmt.Sprintf("[✅已启用] %s", wa.Address)
			if wa.Status == model.StatusDisable {
				_address = fmt.Sprintf("[❌已禁用] %s", wa.Address)
			}

			inlineBtn = append(inlineBtn, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(_address, fmt.Sprintf("%s|%v", cbAddress, wa.Id))))
		}
	}

	inlineBtn = append(inlineBtn, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("👛 添加新的钱包地址", cbAddressAdd)))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineBtn...)

	SendMsg(msg)
}

func cmdUsdtHandle() {
	var msg = tgbotapi.NewMessage(0, fmt.Sprintf("🪧交易所最新基准汇率：`%v`\n✅订单实际计算浮动汇率：`%v`",
		usdt.GetOkxLastRate(), usdt.GetLatestRate()))
	msg.ParseMode = tgbotapi.ModeMarkdown

	SendMsg(msg)
}
