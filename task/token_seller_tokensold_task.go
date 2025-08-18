package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"log/slog"
	"math/big"
	"strconv"
	"time"
	"token-seller/config"
	"token-seller/contracts"
	"token-seller/datadefine"
	"token-seller/handler"
	"token-seller/utils"
)

var tokenSellerTokenSoldTask = &datadefine.TokenSellerTask{
	Addr: "0xF3D6dde572b0Ba1380aD766BfE4495bbd35Ab793",
	// Addr:        "0x67ceaC89be67D461D405A950f5751bBe4A9113a6",
	ConfigFile:  "token_seller_tokensold_task_config.json",
	AlertType:   "BSC",
	AlertConfig: &config.AlertConfig{},
}

var Token_seller_tokensold_task_RetryTimes = 0

func init() {
	// 初始化时加载配置
	// 如果配置文件不存在，使用默认配置
	alertConfigFile := &config.AlertConfigFile{
		LastBlockNumber: "54012738",
		LogIndexs:       nil,
	}
	alertConfigFile = config.LoadConfig(tokenSellerTokenSoldTask.ConfigFile, alertConfigFile)
	if alertConfigFile == nil {
		return
	}
	tokenSellerTokenSoldTask.AlertConfig.ConfigData = alertConfigFile

	// 启动文件监控
	go config.WatchConfig(tokenSellerTokenSoldTask.ConfigFile, tokenSellerTokenSoldTask.AlertConfig)

	// 解析 ABI
	var err error
	if tokenSellerTokenSoldTask.PairAbi, err = contracts.LoadABI("./abis/TokenSellerV2V3.json"); err != nil {
		return
	}

}

func Token_seller_tokensold_task() error {
	slog.Info("start excute Token_seller_tokensold_task")
	events, toBlock, err := GetTokensSoldEvents(tokenSellerTokenSoldTask)
	if err != nil {
		slog.Error("Error fetching tokensSoldEvents", "error", err)
		time.Sleep(1 * time.Second)
		return err
	}

	if events == nil || len(events) == 0 {
		Token_seller_tokensold_task_RetryTimes++
		if Token_seller_tokensold_task_RetryTimes > 2 {
			tokenSellerTokenSoldTask.AlertConfig.SetLastBlockNumber(toBlock)
			config.SaveConfig(tokenSellerTokenSoldTask.ConfigFile, tokenSellerTokenSoldTask.AlertConfig.ConfigData)
			Token_seller_tokensold_task_RetryTimes = 0
			slog.Info("retry reached 3 times, no new tokensSoldEvents found，update next time startblocknumber")
			return nil
		}
		slog.Info("no new tokensSoldEvents found, start retry", "retryTimes", Token_seller_tokensold_task_RetryTimes)
		time.Sleep(1 * time.Second)
		return nil
	}

	for _, event := range events {
		jsonData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal TokensSoldEvent: %w", err)
		}
		slog.Info("TokensSold", "event", string(jsonData))

		err = utils.RedisSet(fmt.Sprintf("trader:%s:%s", event.Token, event.Seller), string(jsonData), 0)
		if err != nil {
			return fmt.Errorf("TokensSoldEvent failed to save redis: %w", err)
		}
	}

	tokenSellerTokenSoldTask.AlertConfig.SetLastBlockNumber(toBlock)
	config.SaveConfig(tokenSellerTokenSoldTask.ConfigFile, tokenSellerTokenSoldTask.AlertConfig.ConfigData)
	slog.Info("end excute Token_seller_tokensold_task")
	return nil
}

func GetTokensSoldEvents(tokenSellerTask *datadefine.TokenSellerTask) ([]datadefine.TokensSoldEvent, string, error) {
	// 连接到BSC节点
	rpcUrl := utils.GetRpcURL(tokenSellerTask.AlertType)
	client, err := handler.GetEthClient(rpcUrl, 5*time.Second)
	defer client.Close()
	if err != nil {
		return nil, "", err
	}

	// 查询当前最新区块, 与已处理区块，判断已处理区块是否>=最新区块，是则没有新区块待处理，直接则返回异常不处理，设置连接超时时间为 5 秒
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		return nil, "", err
	}
	handledBlock, _ := strconv.ParseUint(tokenSellerTask.AlertConfig.GetLastBlockNumber(), 10, 64)
	if handledBlock >= blockNumber {
		return nil, "", fmt.Errorf("The latest block has been reached, handledBlock: %s latestBlockNumber: %s", strconv.FormatUint(handledBlock, 10), strconv.FormatUint(blockNumber, 10))
	}

	// 构建分批查询，查询事件日志
	var fromBlock uint64
	var toBlock uint64
	if blockNumber-5 <= (handledBlock + 100) {
		// 当实时处理新增区块数据时，为避免节点区块生成但是对应的vlog数据未生成，每次查询开始区块覆盖已处理的前两个区块
		fromBlock = handledBlock + 1
		toBlock = blockNumber - 5
	} else {
		// 当分页处理旧区块数据时，vlog数据一般已生成，因此直接开始查
		fromBlock = handledBlock + 1
		toBlock = handledBlock + 100
	}

	// 解析TokensSold事件
	pAddress := common.HexToAddress(tokenSellerTask.Addr)
	eventSignature := [][]interface{}{{tokenSellerTask.PairAbi.Events["TokensSold"].ID}}
	topics, err := abi.MakeTopics(eventSignature...)
	if err != nil {
		return nil, "", err
	}
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{pAddress},
		Topics:    topics,
	}
	// 设置连接超时时间为 60 s
	ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		return nil, "", err
	}
	slog.Info("TokensSold logs", "len", len(logs))

	tokensSoldEvents := make([]datadefine.TokensSoldEvent, 0)
	var currentBlock *types.Block = nil
	for _, vLog := range logs {
		// 多个vLog可能在同一个区块，每次处理的vLog的区块发生变化则重新获取新的区块信息保存到currentBlock
		if currentBlock == nil || vLog.BlockNumber != currentBlock.NumberU64() {
			// 设置连接超时时间为 5 秒
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			block, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(vLog.BlockNumber))
			if err != nil {
				return nil, "", err
			}
			currentBlock = block
		}

		var event datadefine.TokensSoldEvent
		err = tokenSellerTask.PairAbi.UnpackIntoInterface(&event, "TokensSold", vLog.Data)
		if err != nil {
			return nil, "", fmt.Errorf("failed to unpack TokensSoldEvent: %w", err)
		}

		event.Token = common.BytesToAddress(vLog.Topics[1].Bytes()).Hex()
		event.Seller = common.BytesToAddress(vLog.Topics[2].Bytes()).Hex()
		event.TxHash = vLog.TxHash.Hex()
		event.BlockTimestamp = strconv.FormatUint(currentBlock.Time(), 10)
		tokensSoldEvents = append(tokensSoldEvents, event)
	}

	toBlockStr := strconv.FormatUint(toBlock, 10)
	slog.Info("get TokensSold events success", "lenth", len(tokensSoldEvents), "toBlock", toBlockStr)
	return tokensSoldEvents, toBlockStr, nil

}
