package web

import "strings"

// PaymentConfig 支付配置结构
type PaymentConfig struct {
	Coin            string // 币种名称 (USDT, USDC, TRX)
	Network         string // 网络标识 (ERC20, TRC20, BEP20, etc.)
	NetworkFullName string // 网络全名 (以太坊 (Ethereum), 波场 (TRON), etc.)
	WarningCoin     string // 警告币种 (ETH, TRX, BNB, etc.)
}

// GetPaymentConfig 根据交易类型获取支付配置
func GetPaymentConfig(tradeType string) PaymentConfig {
	// 解析交易类型，格式: coin.network (如 usdt.erc20)
	parts := strings.Split(tradeType, ".")
	if len(parts) != 2 {
		// 默认配置
		return PaymentConfig{
			Coin:            "USDT",
			Network:         "TRC20",
			NetworkFullName: "波场 (TRON)",
			WarningCoin:     "TRX",
		}
	}

	coin := strings.ToUpper(parts[0])
	network := strings.ToUpper(parts[1])

	// 网络配置映射
	networkConfigs := map[string]PaymentConfig{
		"usdt.erc20": {
			Coin:            "USDT",
			Network:         "ERC20",
			NetworkFullName: "以太坊 (Ethereum)",
			WarningCoin:     "ETH",
		},
		"usdt.trc20": {
			Coin:            "USDT",
			Network:         "TRC20",
			NetworkFullName: "波场 (TRON)",
			WarningCoin:     "TRX",
		},
		"usdt.bep20": {
			Coin:            "USDT",
			Network:         "BEP20",
			NetworkFullName: "币安智能链 (BSC)",
			WarningCoin:     "BNB",
		},
		"usdt.polygon": {
			Coin:            "USDT",
			Network:         "Polygon",
			NetworkFullName: "Polygon",
			WarningCoin:     "MATIC",
		},
		"usdt.arbitrum": {
			Coin:            "USDT",
			Network:         "Arbitrum",
			NetworkFullName: "Arbitrum One",
			WarningCoin:     "ETH",
		},
		"usdt.solana": {
			Coin:            "USDT",
			Network:         "Solana",
			NetworkFullName: "Solana",
			WarningCoin:     "SOL",
		},
		"usdt.aptos": {
			Coin:            "USDT",
			Network:         "Aptos",
			NetworkFullName: "Aptos",
			WarningCoin:     "APT",
		},
		"usdt.xlayer": {
			Coin:            "USDT",
			Network:         "X Layer",
			NetworkFullName: "OKX (X Layer)",
			WarningCoin:     "OKB",
		},
		"usdc.erc20": {
			Coin:            "USDC",
			Network:         "ERC20",
			NetworkFullName: "以太坊 (Ethereum)",
			WarningCoin:     "ETH",
		},
		"usdc.trc20": {
			Coin:            "USDC",
			Network:         "TRC20",
			NetworkFullName: "波场 (TRON)",
			WarningCoin:     "TRX",
		},
		"usdc.bep20": {
			Coin:            "USDC",
			Network:         "BEP20",
			NetworkFullName: "币安智能链 (BSC)",
			WarningCoin:     "BNB",
		},
		"usdc.polygon": {
			Coin:            "USDC",
			Network:         "Polygon",
			NetworkFullName: "Polygon",
			WarningCoin:     "MATIC",
		},
		"usdc.arbitrum": {
			Coin:            "USDC",
			Network:         "Arbitrum",
			NetworkFullName: "Arbitrum One",
			WarningCoin:     "ETH",
		},
		"usdc.solana": {
			Coin:            "USDC",
			Network:         "Solana",
			NetworkFullName: "Solana",
			WarningCoin:     "SOL",
		},
		"usdc.aptos": {
			Coin:            "USDC",
			Network:         "Aptos",
			NetworkFullName: "Aptos",
			WarningCoin:     "APT",
		},
		"usdc.xlayer": {
			Coin:            "USDC",
			Network:         "X Layer",
			NetworkFullName: "OKX (X Layer)",
			WarningCoin:     "OKB",
		},
		"tron.trx": {
			Coin:            "TRX",
			Network:         "TRON",
			NetworkFullName: "波场 (TRON)",
			WarningCoin:     "TRX",
		},
	}

	// 查找配置
	if config, exists := networkConfigs[tradeType]; exists {
		return config
	}

	// 如果没找到，尝试动态构建
	return PaymentConfig{
		Coin:            coin,
		Network:         network,
		NetworkFullName: network,
		WarningCoin:     coin,
	}
}
