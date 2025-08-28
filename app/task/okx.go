package task

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"github.com/v03413/bepusdt/app/conf"
	"github.com/v03413/bepusdt/app/log"
	"github.com/v03413/bepusdt/app/task/rate"
)

func init() {
	d := getExchangeRateUpdateInterval()
	register(task{duration: d, callback: OkxUsdtRateStart})
	register(task{duration: d, callback: OkxUsdcRateStart})
	register(task{duration: d, callback: OkxTrxRateStart})
}

func getExchangeRateUpdateInterval() time.Duration {
	var exchangeRateUpdateInterval time.Duration
	v := os.Getenv(`BEPUSDT_EXCHANGE_RATE_UPDATE_INTERVAL`)
	if len(v) > 0 {
		var err error
		exchangeRateUpdateInterval, err = time.ParseDuration(v)
		if err != nil {
			log.Errorf(`解析环境变量 BEPUSDT_EXCHANGE_RATE_UPDATE_INTERVAL 的值“%s”失败: %v`, v, err)
		}
	}
	if exchangeRateUpdateInterval <= 0 {
		exchangeRateUpdateInterval = time.Minute * 30
	}
	return exchangeRateUpdateInterval
}

// OkxUsdtRateStart Okx USDT_CNY 汇率监控
func OkxUsdtRateStart(ctx context.Context) {
	var rawRate, err = getOkxUsdTokenCnySellPrice(ctx, "USDT")
	if err != nil {
		log.Error("Okx USDT_CNY 汇率获取失败", err)
	} else {
		rate.SetOkxUsdtCnyRate(conf.GetUsdtRate(), rawRate)
	}

	log.Debug("当前 USDT_CNY 计算汇率：", rate.GetUsdtCalcRate())
}

// OkxUsdcRateStart Okx USDC_CNY 汇率监控
func OkxUsdcRateStart(ctx context.Context) {
	var rawRate, err = getOkxUsdTokenCnySellPrice(ctx, "USDC")
	if err != nil {
		log.Error("Okx USDC_CNY 汇率获取失败", err)
	} else {
		rate.SetOkxUsdcCnyRate(conf.GetUsdcRate(), rawRate)
	}

	log.Debug("当前 USDC_CNY 计算汇率：", rate.GetUsdcCalcRate())
}

// OkxTrxRateStart  Okx TRX_CNY 汇率监控
func OkxTrxRateStart(ctx context.Context) {
	var price, err = getOkxTrxCnyMarketPrice(ctx)
	if err != nil {
		log.Error("Okx TRX_USDT 汇率获取失败", err)
	} else {
		rate.SetOkxTrxCnyRate(conf.GetTrxRate(), price)
	}

	log.Debug("当前 TRX_CNY 计算汇率：", rate.GetTrxCalcRate())
}

// getOkxUsdtCnySellPrice  Okx  C2C快捷交易 USDT出售 实时汇率
func getOkxUsdTokenCnySellPrice(ctx context.Context, crypto string) (float64, error) {
	if crypto != "USDT" && crypto != "USDC" {
		return 0, errors.New("unsupported crypto:" + crypto)
	}

	t := strconv.Itoa(int(time.Now().Unix()))
	okxApi := fmt.Sprintf(
		"https://www.okx.com/v4/c2c/express/price?crypto=%s&fiat=CNY&side=sell&t=%s",
		crypto, t,
	)

	var c = &http.Client{Timeout: time.Second * 30, Transport: &http.Transport{TLSClientConfig: &tls.Config{NextProtos: []string{"http/1.1"}}}}

	req, err := http.NewRequestWithContext(ctx, "GET", okxApi, nil)
	if err != nil {
		return 0, fmt.Errorf(`okx creating request error: %v`, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
	resp, err := c.Do(req)
	if err != nil {

		return 0, errors.New("okx resp error:" + err.Error())
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {

		return 0, errors.New("okx resp status code:" + strconv.Itoa(resp.StatusCode))
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {

		return 0, errors.New("okx resp read error:" + err.Error())
	}

	result := gjson.ParseBytes(all)
	if result.Get("error_code").Int() != 0 {

		return 0, errors.New("json parse error:" + result.Get("error_message").String())
	}

	if result.Get("data.price").Exists() {
		var _ret = result.Get("data.price").Float()
		if _ret <= 0 {
			return 0, errors.New("okx resp json data.price <= 0")
		}

		return cast.ToFloat64(_ret), nil
	}

	return 0, errors.New("okx resp json data.price not found")
}

// getOkxTrxCnyMarketPrice 获取 Trx/Cny 市场价格 https://www.okx.com/zh-hans/convert/trx-to-cny
func getOkxTrxCnyMarketPrice(ctx context.Context) (float64, error) {
	var t = strconv.Itoa(int(time.Now().Unix()))
	var okxApi = "https://www.okx.com/priapi/v3/growth/convert/currency-pair-market-movement?baseCurrency=TRX&quoteCurrency=CNY&bar=4H&limit=1&t=" + t

	req, err := http.NewRequestWithContext(ctx, "GET", okxApi, nil)
	if err != nil {
		return 0, fmt.Errorf(`okx creating request errror: %v`, err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("app-type", "web")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://www.okx.com/zh-hans/convert/trx-to-cny")
	req.Header.Set("sec-ch-ua", "\"Google Chrome\";v=\"131\", \"Chromium\";v=\"131\", \"Not_A Brand\";v=\"24\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"macOS\"")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("x-locale", "zh_CN")
	req.Header.Set("x-utc", "8")
	req.Header.Set("x-zkdex-env", "0")

	var c = &http.Client{Timeout: time.Second * 30, Transport: &http.Transport{TLSClientConfig: &tls.Config{NextProtos: []string{"http/1.1"}}}}

	resp, err := c.Do(req)
	if err != nil {

		return 0, errors.New("okx resp error:" + err.Error())
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {

		return 0, errors.New("okx resp status code:" + strconv.Itoa(resp.StatusCode))
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {

		return 0, errors.New("okx resp read error:" + err.Error())
	}

	result := gjson.ParseBytes(all)
	if result.Get("error_code").Int() != 0 {

		return 0, errors.New("json parse error:" + result.Get("error_message").String())
	}

	var list = result.Get("data.datapointList").Array()
	if len(list) == 0 {

		return 0, errors.New("okx resp json data.datapointList not found")
	}

	return list[0].Get("price").Float(), nil
}
