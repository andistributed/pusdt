package conf

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

const (
	Bsc      = "bsc" // Binance Smart Chain
	Tron     = "tron"
	Aptos    = "aptos"
	Solana   = "solana"
	Xlayer   = "xlayer"
	Polygon  = "polygon"
	Arbitrum = "arbitrum"
	Ethereum = "ethereum"
)

var (
	cfg  Conf
	path string
)

func Init() error {
	flag.StringVar(&path, "conf", "./conf.toml", "config file path")
	flag.Parse()

	data, err := os.ReadFile(path)
	if err != nil {

		return fmt.Errorf("配置文件加载失败：%w", err)
	}

	if err = toml.Unmarshal(data, &cfg); err != nil {

		return fmt.Errorf("配置数据解析失败：%w", err)
	}

	cfg.setDefaults()

	if BotToken() == "" || BotAdminID() == 0 {

		return errors.New("telegram bot 参数 admin_id 或 token 均不能为空")
	}

	return nil
}

func GetUsdtRate() string {

	return cfg.Pay.UsdtRate
}

func GetUsdcRate() string {

	return cfg.Pay.UsdcRate
}

func GetTrxRate() string {

	return cfg.Pay.TrxRate
}

func GetUsdtAtomicity() (decimal.Decimal, int) {
	var val = defaultUsdtAtomicity
	if cfg.Pay.UsdtAtom != 0 {

		val = cfg.Pay.UsdtAtom
	}

	var atom = decimal.NewFromFloat(val)

	return atom, cast.ToInt(math.Abs(float64(atom.Exponent())))
}

func GetUsdcAtomicity() (decimal.Decimal, int) {
	var val = defaultUsdcAtomicity
	if cfg.Pay.UsdcAtom != 0 {

		val = cfg.Pay.UsdcAtom
	}

	var atom = decimal.NewFromFloat(val)

	return atom, cast.ToInt(math.Abs(float64(atom.Exponent())))
}

func GetTrxAtomicity() (decimal.Decimal, int) {
	var val = defaultTrxAtomicity
	if cfg.Pay.TrxAtom != 0 {

		val = cfg.Pay.TrxAtom
	}

	var atom = decimal.NewFromFloat(val)

	return atom, cast.ToInt(math.Abs(float64(atom.Exponent())))
}

func GetExpireTime() time.Duration {
	if cfg.Pay.ExpireTime == 0 {

		return time.Duration(defaultExpireTime)
	}

	return time.Duration(cfg.Pay.ExpireTime)
}

func GetExpireSeconds() time.Duration {
	return GetExpireTime() * time.Second
}

func IsExpired(ts int64) bool {
	return time.Now().After(time.Unix(ts, 0).Add(GetExpireSeconds()))
}

func GetAuthToken() string {
	if cfg.AuthToken == "" {

		return defaultAuthToken
	}

	return cfg.AuthToken
}

func GetAppUri(host string) string {
	if cfg.AppUri != "" {

		return cfg.AppUri
	}

	return host
}

func GetStaticPath() string {

	return cfg.StaticPath
}

func GetSqlitePath() string {
	if cfg.SqlitePath != "" {

		return cfg.SqlitePath
	}

	return filepath.Join(executeDir(), defaultSqlitePath)
}

func executeDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) { // go run
		return `.`
	}
	return filepath.Dir(os.Args[0])
}

func GetOutputLog() string {
	if cfg.OutputLog != "" {

		return cfg.OutputLog
	}

	return filepath.Join(executeDir(), defaultOutputLog)
}

func GetListen() string {
	if cfg.Listen != "" {

		return cfg.Listen
	}

	return defaultListen
}

func BotToken() string {
	var token = strings.TrimSpace(os.Getenv("BOT_TOKEN"))
	if token != "" {

		return token
	}

	return cfg.Bot.Token
}

func BotAdminID() int64 {
	var id = strings.TrimSpace(os.Getenv("BOT_ADMIN_ID"))
	if id != "" {

		return cast.ToInt64(id)
	}

	return cfg.Bot.AdminID
}

func BotNotifyTarget() string {
	if cfg.Bot.GroupID != "" {

		return cfg.Bot.GroupID
	}

	return cast.ToString(cfg.Bot.AdminID)
}

func GetWalletAddress() []string {

	return cfg.Pay.WalletAddress
}

func GetTradeIsConfirmed() bool {

	return cfg.Pay.TradeIsConfirmed
}

func GetPaymentAmountMin() decimal.Decimal {
	var val = defaultPaymentMinAmount
	if cfg.Pay.PaymentAmountMin != 0 {

		val = cfg.Pay.PaymentAmountMin
	}

	return decimal.NewFromFloat(val)
}

func GetPaymentAmountMax() decimal.Decimal {
	var val float64 = defaultPaymentMaxAmount
	if cfg.Pay.PaymentAmountMax != 0 {

		val = cfg.Pay.PaymentAmountMax
	}

	return decimal.NewFromFloat(val)
}

func GetWebhookUrl() string {

	return cfg.WebhookUrl
}

func GetConfig() Conf {
	return cfg
}

func GetDebug() bool {
	if v := os.Getenv(`BEPUSDT_DEBUG`); len(v) > 0 {
		cfg.Debug, _ = strconv.ParseBool(v)
	}
	return cfg.Debug
}

func GetLogLevel() string {
	if v := os.Getenv(`BEPUSDT_LOG_LEVEL`); len(v) > 0 {
		return v
	}
	return ``
}

func GetLogOutputConsole() bool {
	if v := os.Getenv(`BEPUSDT_LOG_OUTPUT_CONSOLE`); len(v) > 0 {
		y, _ := strconv.ParseBool(v)
		return y
	}
	return false
}

func GetAppName() string {

	return cfg.AppName
}

func GetHomeURL() string {

	return cfg.HomeURL
}
