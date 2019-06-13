# hubble 接口文档

## /marketData/*
连行情服务器

### GET /marketData/STATUS
对应 hgetall STATUS

返回所有交易所的状态。这里返回的key是所有有效的交易所名
```json
{
 	"BINANCE":{"Time": "2019-05-20 02:27:10.783452", "Connected":true},
 	"GDAX":{"Time": "2019-05-20 02:27:10.783479", "Connected": true},
 	...
}
```

### GET /marketData/TRADEx
对应 psubscribe TRADEx\*

返回所有币对的最新成交价。这里的返回中，所有有效的币对各出现一次。如果自服务启动到该接口被调用这段时间内，某币对没有成交过，那么字符型字段为空，数字型字段为-1或-1.0。
```json
[
	{"Exchange": "GDAX", "Symbol": "BTC_USD", "Time": "2019-05-21 02:40:41.417000", "TimeArrive": "2019-05-21 02:40:41.542170", "Direction": "Buy", "LastPx": 7905.1, "Qty": 0.00351662},
	...
]
```

### POST /marketData/KLINE
对应 hget KLINE ETH_BTC.GDAX

调用参数为 币对 和 交易所
```
{
	"Symbol":"ETH_BTC",
	"Exchange":"BINANCE"
}
```

返回值为1440分钟k线数据
list<list<Timestamp, Open, High, Low, Close, Volume, Quotes_cum>,>
```json
[
	[1558248420.0, 0.03171, 0.03171, 0.03171, 0.03171, 0.26688195000000003, 2.0], 
	[1558248480.0, 0.03171, 0.03171, 0.03168, 0.03168, 0.31260202000000004, 12.0], 
	[1558248540.0, 0.03171, 0.03173, 0.03171, 0.03172, 2.4976941599999996, 4.0], 
	[1558248600.0, 0.03171, 0.03171, 0.03169, 0.0317, 0.9206529300000001, 5.0],
	... 
]
```

### POST /marketData/DEPTHx
对应 get DEPTHx|ETH_BTC.GDAX

调用参数为 币对
```
{
	"Symbol":"ETH_BTC"
}
```

统一推送所有交易所该币对的DEPTHx
```
[
	{
        "Msg_seq": 4809082, 
        "Exchange": "GDAX",
        "Symbol": "ETH_BTC",
        "Time": "2019-05-20 07:06:11.144273", 
        "TimeArrive": "2019-05-20 07:06:11.144290", 
    	"AskDepth": [[0.03152, 0.72885946], [0.03154,75.01944855], ... ], // [price,volume], price ascending
    	"BidDepth": [[0.03151, 54.88], [0.0315, 0.19], ... ], // [price,volume], price descending
    },
    ...	
]
```

## /trading/*
//连交易服务器
### GET /trading/STATUS
统一推送所有交易所所有柜台
返回所有柜台的状态。这里返回的key是所有有效的柜台名。
```json
{
	"XDAEXK1": {
		"INIT_AT":"2019-05-22 08:15:22.269939", // 启动时间
		"UPDATED_AT":"2019-05-22 08:16:44.988385", // 状态更新时间(每秒更新一次)
		"LOGIN_AT":"", // 系统注册时间
		"TIMESPAN":0, // 交易所校时差值
		"ALIAS":"", // 实例别名(与消息主题一致)
		"TARGET":"https://test1.pro.hashkey.com:5566/APITrade", // REST接口目标地址
		"PENDING_ORDERS":0, // 待交易所返回单号订单数
		"ORDERS":0, // 当前挂出订单数
		"MISSING_MATCH":[0, []], // 成交结果无法匹配数
		"BUSY_WORKERS":0, // 当前REST并发数
		"FREE_WORKERS":1, // 当前可用请求实例数
		"WS_STATUS":-1 // Websocket当前连线状态(默认为0, 如果Websocket不可用)
	},
	...
}
```
### POST /trading/MYORDERS
和 DEPTHx 一样

### GET /trading/GALAXY
对应 psubscribe GalaxyDetail
今后可能在active这一级加其他字段

```json
{
	"ACTIVE": 1 // 1代表galaxy处于激活状态，0代表处于非激活状态
}
```

### GET /trading/SINGULARITY
对应 psubscribe StrategyDetail
今后可能在active这一级加其他字段

```json
{
	"strategy1": {
		"Active": 1
	},
	"用户自由命名": {
		"Active": 0
	},
	...
}
```

### GET /trading/PORTFOLIOS
投资组合状态
```json
[
	{
		"PortfolioName":"HUB",
		"Balance":{
			"BTC":{"free":1,"locked":2,"total":3},
			"ETH":{"free":10,"locked":0,"total":10},
			...
		},
		"PnL":100
	},
	...
]
```

## /sentinel
连接redis哨兵
### GET /sentinel/STATUS
显示活着的实例
```json
{
	"sentinel":["127.0.0.1:26379","127.0.0.1:26380","127.0.0.1:26381"],
	"trading":{
		"master":"127.0.0.1:6379",
		"slave":["127.0.0.1:6380","127.0.0.1:6381"]
	},
	"marketData":{
		"master":"127.0.0.1:6379",
		"slave":["127.0.0.1:6380","127.0.0.1:6381"]
	},
	"OMS":{
		"master":"127.0.0.1:6379",
		"slave":["127.0.0.1:6380","127.0.0.1:6381"]
	},
	"galaxy":{
		"master":"127.0.0.1:6379",
		"slave":["127.0.0.1:6380","127.0.0.1:6381"]
	}
}
```