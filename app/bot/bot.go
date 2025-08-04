package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/v03413/bepusdt/app/conf"
	"github.com/v03413/bepusdt/app/log"
)

var api *bot.Bot
var err error

func Init() error {
	var opts = []bot.Option{
		//bot.WithDebug(),
		bot.WithCheckInitTimeout(time.Minute),
		bot.WithMiddlewares([]bot.Middleware{updateFilter}...),
		bot.WithDefaultHandler(defaultHandle),
	}

	api, err = bot.New(conf.BotToken(), opts...)

	return err
}

func Start() {
	var ctx, cancel = context.WithCancel(context.Background())

	defer cancel()

	var me, err2 = api.GetMe(ctx)
	if err2 != nil {
		panic(err2)
	}

	{
		api.RegisterHandler(bot.HandlerTypeMessageText, cmdGetId, bot.MatchTypeCommand, cmdGetIdHandle)
		api.RegisterHandler(bot.HandlerTypeMessageText, cmdStart, bot.MatchTypeCommand, cmdStartHandle)
		api.RegisterHandler(bot.HandlerTypeMessageText, cmdState, bot.MatchTypeCommand, cmdStateHandle)
		api.RegisterHandler(bot.HandlerTypeMessageText, cmdOrder, bot.MatchTypeCommand, cmdOrderHandle)

		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbOrderDetail, bot.MatchTypePrefix, cbOrderDetailAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbWallet, bot.MatchTypePrefix, cbWalletAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAddress, bot.MatchTypePrefix, cbAddressAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAddressAdd, bot.MatchTypePrefix, cbAddressAddAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAddressType, bot.MatchTypePrefix, cbAddressTypeAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAddressDelete, bot.MatchTypePrefix, cbAddressDelAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAddressBack, bot.MatchTypePrefix, cbAddressBackAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAddressEnable, bot.MatchTypePrefix, cbAddressEnableAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAddressDisable, bot.MatchTypePrefix, cbAddressDisableAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAddressOtherNotify, bot.MatchTypePrefix, cbAddressOtherNotifyAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbMarkNotifySucc, bot.MatchTypePrefix, cbMarkNotifySuccAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbOrderNotifyRetry, bot.MatchTypePrefix, dbOrderNotifyRetryAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbMarkOrderSucc, bot.MatchTypePrefix, dbMarkOrderSuccAction)
		api.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbOrderList, bot.MatchTypePrefix, cbOrderListAction)
	}

	_, err = api.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: cmdGetId, Description: "获取ID"},
			{Command: cmdStart, Description: "开始使用"},
			{Command: cmdState, Description: "收款状态"},
			{Command: cmdOrder, Description: "订单列表"},
		},
	})
	if err != nil {
		panic("SetMyCommandsParams Error: " + err.Error())
	}
	_, err = api.DeleteWebhook(ctx, &bot.DeleteWebhookParams{DropPendingUpdates: true})
	if err != nil {
		panic("DeleteWebhook Error: " + err.Error())
	}

	SendMessage(&bot.SendMessageParams{
		ChatID: conf.BotNotifyTarget(),
		Text:   Welcome(),
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "📢 关注频道", URL: "https://t.me/BEpusdtChannel"},
					{Text: "💬 社区交流", URL: "https://t.me/BEpusdtChat"},
				},
			},
		},
	})

	fmt.Printf("Bot UserName: %s %s%s\n", me.Username, me.FirstName, me.LastName)

	api.Start(ctx)
}

func SendMessage(p *bot.SendMessageParams) {
	if p.ChatID == nil {
		p.ChatID = conf.BotAdminID()
	}

	var ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := api.SendMessage(ctx, p)
	if err != nil {

		log.Warn("Bot Send Message Error:", err.Error())
	}
}

func DeleteMessage(ctx context.Context, b *bot.Bot, p *bot.DeleteMessageParams) {
	_, err := b.DeleteMessage(ctx, p)
	if err != nil {

		log.Warn("Bot Delete Message Error:", err.Error())
	}
}

func EditMessageText(ctx context.Context, b *bot.Bot, p *bot.EditMessageTextParams) {
	_, err := b.EditMessageText(ctx, p)
	if err != nil {

		log.Warn("BotEditMessageText Error:", err.Error())
	}
}

func EditMessageReplyMarkup(ctx context.Context, b *bot.Bot, p *bot.EditMessageReplyMarkupParams) {
	_, err := b.EditMessageReplyMarkup(ctx, p)
	if err != nil {

		log.Warn("BotEditMessageReplyMarkup Error:", err.Error())
	}
}
