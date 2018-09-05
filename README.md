# 项目介绍
luckybtsbot 是基于 [Telegram](https://telegram.org/) 的比特股(BTS)红包机器人服务。基于 [luckybot](https://github.com/zhangpanyi/luckybot) 项目开发，只简单修改配置文件以及编写部分 Lua 脚本，即实现了此项目。而对接比特股(BTS)部分则搭建了一个 [btsmonitor](https://github.com/zhangpanyi/btsmonitor) 服务。现在可以在 Telegram 中搜索用户 `@luckybts_bot`，或者点击 [http://telegram.me/luckybts_bot](http://telegram.me/luckybts_bot) 进行使用体验。

# 运行环境
* Python
* Docker 18.06

# 部署指南
luckybtsbot 基于 Docker 部署，需要安装 Docker 18.06，安装教程参考：[Get Docker CE](https://docs.docker.com/install/)。另外需要创建一个 Telegram 机器人，并开启 [Inline mode](https://core.telegram.org/bots/inline) 功能。

### 1. Overlay Network
btsmonitor 服务用于比特股转账以及发送收款通知，它需要和 luckybtsbot 服务进行通信。目前的做法是将它们放在不同的 Docker 容器之中，这两个容器都加入同一个Overlay Network，这样它们就可以互相访问同时不能被外部访问了。执行以下命令：
```bash
sudo docker swarm init
sudo docker network create --driver=overlay --subnet=10.0.1.0/24 --attachable luckybot
```

### 2. 部署 btsmonitor

首先拉取 btsmonitor 项目代码：
```bash
git clone https://github.com/zhangpanyi/btsmonitor
cd btsmonitor
```

然后运行初始化配置脚本，如需要修改比特股网络和账户配置请编辑 `server.yml` 文件：
```bash
python init_config.py
```

然后构建 Docker 镜像：
```bash
sudo docker build -t="btsmonitor" -f docker/Dockerfile .
```

最后运行容器：
```bash
sudo docker run --name btsmonitor --network luckybot --ip 10.0.1.101 -d btsmonitor
```

如需查看容器日志执行命令：
```bash
sudo docker logs -f btsmonitor
```

### 3. 部署 luckybtsbot

运行初始化配置脚本，如需要修改配置请编辑 `server.yml` 文件：
```bash
python init_config.py
```

然后构建 Docker 镜像：
```bash
sudo docker build -t="luckybtsbot" -f docker/Dockerfile .
```

最后运行容器：
```bash
sudo docker run --name luckybtsbot --network luckybot --ip 10.0.1.102 -d -p 80:80 luckybtsbot
```

如需查看容器日志执行命令：
```bash
sudo docker logs -f luckybtsbot
```
