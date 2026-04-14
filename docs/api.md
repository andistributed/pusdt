## 📚 接口文档

<details>
<summary>创建订单</summary>  

### 请求地址

```http
POST /api/v1/order/create-transaction
```

- 使用相同订单号创建订单时，不会产生两个交易；T1时间创建完成，T2时间重复提交会根据实际参数重建订单，超时暂时不重置。  
- 因为支持订单重建，所以对于商户端来讲，可以独立实现收银台，针对同一个订单号，随意变更交易类型、地址和金额。  

### 请求数据

```json
{
  "address": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",  // 可根据实际情况传入收款地址，亦可留空
  "trade_type": "usdt.trc20",  // usdt.trc20(默认) 可选完整列表 https://github.com/v03413/BEpusdt/blob/main/docs/trade-type.md
  "order_id": "787240927112940881",   // 商户订单编号
  "amount": 28.88,   // 请求支付金额，CNY
  "signature":"123456abcd", // 签名
  "notify_url": "https://example.com/callback",   // 回调地址
  "redirect_url": "https://example.com/callback", // 支付成功跳转地址
  "timeout": 1200, // 超时时间(秒) 最低60；留空则取配置文件 expire_time，还是没有取默认600
  "rate": 7.4 // 强制指定汇率，留空则取配置汇率；支持多种写法，如：7.4表示固定7.4、～1.02表示最新汇率上浮2%、～0.97表示最新汇率下浮3%、+0.3表示最新加0.3、-0.2表示最新减0.2
}
```

### 响应内容

```json
{
  "status_code": 200,
  "message": "success",
  "data": {
    "trade_id": "b3d2477c-d945-41da-96b7-f925bbd1b415", // 本地交易ID
    "order_id": "787240927112940881", // 商户订单编号
    "amount": "28.88", // 请求支付金额，CNY
    "token_amount": "10", // 实际支付数额 usdt or trx
    "token": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", // 收款地址
    "expiration_time": 1200, // 订单有效期，秒
    "payment_url": "https://example.com//pay/checkout-counter/b3d2477c-d945-41da-96b7-f925bbd1b415"  // 收银台地址
  },
  "request_id": ""
}

```

</details>

<details>
<summary>取消订单</summary>  

商户端系统可以通过此接口取消订单，取消后，系统将不再监控此订单，同时释放对应金额占用。

### 请求地址

```http
POST /api/v1/order/cancel-transaction
```

### 请求数据

```json
{
  "trade_id": "0TJV0br98YbNTQe7nQ",   // 交易ID
  "signature":"123456abcd" // 签名内容
}
```

### 响应内容

```json
{
  "data": {
    "trade_id": "0TJV0br98YbNTQe7nQ"
  },
  "message": "success",
  "request_id": "",
  "status_code": 200
}
```

</details>

<details>
<summary>回调通知</summary>

```json
{
  "trade_id": "b3d2477c-d945-41da-96b7-f925bbd1b415",
  "order_id": "787240927112940881",
  "amount": 28.88,
  "token_amount": 10,
  "token": "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
  "block_transaction_id": "12ef6267b42e43959795cf31808d0cc72b3d0a48953ed19c61d4b6665a341d10",
  "signature": "123456abcd",
  "status": 2   //  1:等待支付  2:支付成功  3:支付超时
}
```

</details>  

## 签名算法

**这里给出一个PHP参考签名函数 [点击查看](https://github.com/v03413/Epay-BEpusdt/blob/b7fa8fd608d71ce50e0f8eabb1717783c96761ac/bepusdt_plugin.php#L108:L127)，其它语言大家统一参考，避免各种奇怪问题。**  

签名生成的通用步骤如下：

第一步，将所有非空参数值的参数按照参数名ASCII码从小到大排序（字典序），使用URL键值对的格式（即key1=value1&key2=value2…）拼接成
`待加密参数`。

重要规则：   
◆ 参数名ASCII码从小到大排序（字典序）；         
◆ 如果参数的值为空不参与签名；        
◆ 参数名区分大小写；
第二步，`待加密参数`最后拼接上`api接口认证token`得到`待签名字符串`，并对`待签名字符串`进行MD5运算，再将得到的`MD5字符串`
所有字符转换为`小写`，得到签名`signature`。 注意：`signature`的长度为32个字节。

举例：

假设传送的参数如下：

```
order_id : 20220201030210321
amount : 42
notify_url : http://example.com/notify
redirect_url : http://example.com/redirect
```

假设api接口认证token为：`epusdt_password_xasddawqe`(api接口认证token可以在`conf.toml`文件设置)

第一步：对参数按照key=value的格式，并按照参数名ASCII字典序排序如下：

```
amount=42&notify_url=http://example.com/notify&order_id=20220201030210321&redirect_url=http://example.com/redirect
```

第二步：拼接API密钥并加密：

```
MD5(amount=42&notify_url=http://example.com/notify&order_id=20220201030210321&redirect_url=http://example.com/redirectepusdt_password_xasddawqe)
```

最终得到最终发送的数据：

```
order_id : 20220201030210321
amount : 42
notify_url : http://example.com/notify
redirect_url : http://example.com/redirect
signature : 1cd4b52df5587cfb1968b0c0c6e156cd
```

## 参考引用

- https://github.com/assimon/epusdt/blob/master/wiki/API.md