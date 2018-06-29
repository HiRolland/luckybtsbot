# Telegram 红包机器人
botcasino 是基于 [Telegram](https://telegram.org/) 机器人的红包机器人。自从中国禁止加密货币与法币的交易后，中国越来越多的加密货币玩家开始使用Telegram聊天。但是Telegram一直没有发红包功能，对于习惯使用QQ和微信的朋友来说，确实是一个遗憾。botcasino 正是为了解决这个问题而被创造出来。由于加密货币的价格不够稳定，在支付或交易的过程中会有很多不方便的地方。所有 botcasino 选择了[比特股](https://bitshares.org/)系统内的智能货币 [BitCNY](https://coinmarketcap.com/currencies/bitcny/)/[BitUSD](https://coinmarketcap.com/currencies/bitusd/) 作为红包的基础货币。

请在 Telegram 中 @luck_money_bot，或者打开 [http://telegram.me/luck_money_bot](http://telegram.me/luck_money_bot) 体验吧。

![](http://i796.photobucket.com/albums/yy247/zhangpanyi/1_zpsuxxjuzgp.png)

# 获取代码
```
git clone https://github.com/zhangpanyi/botcasino.git
glide install
```

# 解析域名
botcasino 服务使用了 [Webhook](https://core.telegram.org/bots/api#setwebhook) 的方式接收机器人消息更新。所有必须准备一个域名，并解析到运行 botcasino 的服务器上。

# 生成密钥
```shell
cd certs

openssl genrsa -out server.key 2048

openssl ecparam -genkey -name secp384r1 -out server.key

openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650 -subj "/C=US/ST=New York/L=Brooklyn/O=Brooklyn Company/CN=YOURDOMAIN"

```
> 注意： 请将 YOURDOMAIN 修改为你自己的域名。

# 账户监控
btsmonitor 用于监控机器人的比特股托管账户。当有人转账到机器人的托管账户时它就会通知 botcasino 服务，用户提现申请也是通过 btsmonitor 去处理，它们之间通过 HTTP 协议进行交流。
```
git clone https://github.com/zhangpanyi/btsmonitor.git
```
请参考 README.md 文档启动 btsmonitor 服务。

# 配置文件
1. `dynamic.yml` 是动态配置文件，可在服务运行期间修改生效，使用默认配置就可以了。
2. `master.yml` 是服务的基本配置文件，启动服务前必须将 `domain`、`token`和`monitor_url`字段改为自己的配置。`domain` 字段请使用 `www.google.com` 格式，不要使用 `https://www.google.com/` 格式。

# 启动服务
```
go build
./botcasino
```

# Docker容器
> 构建容器前请先编译生成 botcasino 可执行文件，并且配置`master.yml`、`dynamic.yml`文件，以及并生成密钥。

```
sudo docker build -t="botcasino" -f docker/Dockerfile .
sudo docker run --name botcasino -d -p 8443:8443 botcasino
```
