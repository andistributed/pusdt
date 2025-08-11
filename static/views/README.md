# 模板重构说明

## 重构概述

本次重构将原来的 17 个重复的 HTML 模板文件合并为一个通用的 `payment.html` 模板，大大提高了代码的可维护性和扩展性。

## 重构前后对比

### 重构前
- 17 个独立的 HTML 文件
- 每个文件内容基本相同，只有币种和网络信息不同
- 新增币种/网络需要创建新文件
- 修改样式需要修改所有文件

### 重构后
- 1 个通用模板 `payment.html`
- 使用 Go 模板语法动态渲染
- 新增币种/网络只需在配置中添加
- 样式修改只需修改一个文件

## 文件结构

```
static/views/
├── payment.html          # 通用支付模板
├── index.html           # 首页模板
└── backup/              # 旧模板备份
    ├── usdt.erc20.html
    ├── usdt.trc20.html
    ├── usdc.erc20.html
    └── ... (其他旧模板)
```

## 新增文件

### app/web/payment_config.go
定义了支付配置结构体和网络配置映射：

```go
type PaymentConfig struct {
    Coin           string // 币种名称 (USDT, USDC, TRX)
    Network        string // 网络标识 (ERC20, TRC20, BEP20, etc.)
    NetworkFullName string // 网络全名 (以太坊 (Ethereum), 波场 (TRON), etc.)
    WarningCoin    string // 警告币种 (ETH, TRX, BNB, etc.)
}
```

## 支持的币种和网络

| 币种 | 网络 | 网络全名 | 警告币种 |
|------|------|----------|----------|
| USDT | ERC20 | 以太坊 (Ethereum) | ETH |
| USDT | TRC20 | 波场 (TRON) | TRX |
| USDT | BEP20 | 币安智能链 (BSC) | BNB |
| USDT | Polygon | Polygon | MATIC |
| USDT | Arbitrum | Arbitrum One | ETH |
| USDT | Solana | Solana | SOL |
| USDT | Aptos | Aptos | APT |
| USDT | X Layer | OKX (X Layer) | OKB |
| USDC | ERC20 | 以太坊 (Ethereum) | ETH |
| USDC | TRC20 | 波场 (TRON) | TRX |
| USDC | BEP20 | 币安智能链 (BSC) | BNB |
| USDC | Polygon | Polygon | MATIC |
| USDC | Arbitrum | Arbitrum One | ETH |
| USDC | Solana | Solana | SOL |
| USDC | Aptos | Aptos | APT |
| USDC | X Layer | OKX (X Layer) | OKB |
| TRX | TRON | 波场 (TRON) | TRX |

## 如何添加新的币种/网络

1. 在 `app/web/payment_config.go` 的 `networkConfigs` 映射中添加新配置
2. 格式：`"coin.network": PaymentConfig{...}`
3. 无需创建新的 HTML 文件

## 模板变量

通用模板 `payment.html` 支持以下变量：

- `{{.Coin}}` - 币种名称
- `{{.Network}}` - 网络标识
- `{{.NetworkFullName}}` - 网络全名
- `{{.WarningCoin}}` - 警告币种
- `{{.amount}}` - 支付金额
- `{{.address}}` - 收款地址
- `{{.order_id}}` - 订单ID
- `{{.trade_id}}` - 交易ID
- `{{.expire}}` - 过期时间
- `{{.return_url}}` - 返回URL

## 好处

1. **维护性提升**：只需维护一个模板文件
2. **扩展性增强**：新增币种/网络只需配置，无需新建文件
3. **一致性保证**：所有支付页面样式完全一致
4. **代码减少**：从 17 个文件减少到 1 个文件
5. **错误减少**：避免了多文件修改导致的不一致问题

## 兼容性

- 完全向后兼容
- 所有现有的 API 接口保持不变
- 用户体验无任何变化
- 旧模板已备份到 `backup/` 目录 