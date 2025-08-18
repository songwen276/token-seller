# token-seller

## 功能介绍

### 1. 解析tokenseller合约的TokensSold事件
解析tokenseller合约的TokensSold事件，并存储到redis中。

### 2. 解析IEC20合约的Approval事件
解析IEC20合约的Approval事件，并存储到redis中。

## 配置文件说明
### 定时任务配置
- `task/task.go`: 该文件用于配置定时任务，控制任务执行间隔：
- `task/token_seller_tokensold_task.go`: 该文件为监控tokenseller合约的tokensold事件主任务，其中全局变量中的tokenSellerTask.Addr为tokenseller合约地址：
- `task/token_seller_approval_task.go`: 该文件为监控授权tokenseller合约代币事件主任务，其中全局变量中的tokenSellerTask.Addr为tokenseller合约地址：
- `utils/url_util.go`: 该文件为获取节点url工具类，需要调整则修改对应链的url固定数组，如BSC链的BSC_URL_ARRAY：

### token_seller_task_config.json
该文件用于配置项目的相关参数，具体配置项说明如下：
- `lastBlockNumber`: 上次处理的区块号,程序重启后会从该区块高度处理。
- `logIndexs`: 当前处理的交易日志索引列表。

示例配置：
```json
{
  "lastBlockNumber": "54012738",
  "logIndexs": null
}

```
## 本地开发启动
```shell
# 安装go环境后，在项目根目录下执行，编译二进制文件
go build -o token-seller main.go
```

## 启动服务
```shell
pm2 start /root/token-seller/start-token-seller.sh --name token-seller
```
## 查看日志,根目录下执行
```shell
cd logs
tail -100f token-seller_output.log
```

