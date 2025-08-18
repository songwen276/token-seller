package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

var tokenSellerApprovalTask = &datadefine.TokenSellerTask{
	Addr: "0xF3D6dde572b0Ba1380aD766BfE4495bbd35Ab793",
	// Addr:        "0x67ceaC89be67D461D405A950f5751bBe4A9113a6",
	ConfigFile:  "token_seller_approval_task_config.json",
	AlertType:   "BSC",
	AlertConfig: &config.AlertConfig{},
}

var Token_seller_approval_task_RetryTimes = 0

func init() {
	// 初始化时加载配置
	// 如果配置文件不存在，使用默认配置
	alertConfigFile := &config.AlertConfigFile{
		LastBlockNumber: "54012738",
		LogIndexs:       nil,
	}
	alertConfigFile = config.LoadConfig(tokenSellerApprovalTask.ConfigFile, alertConfigFile)
	if alertConfigFile == nil {
		return
	}
	tokenSellerApprovalTask.AlertConfig.ConfigData = alertConfigFile

	// 启动文件监控
	go config.WatchConfig(tokenSellerApprovalTask.ConfigFile, tokenSellerApprovalTask.AlertConfig)

	// 解析 ABI
	var err error
	if tokenSellerApprovalTask.PairAbi, err = contracts.LoadABI("./abis/TokenSellerV2V3.json"); err != nil {
		return
	}

}

func Token_seller_approval_task() error {
	slog.Info("start excute Token_seller_approval_task")
	events, toBlock, err := GetApprovalEvents(tokenSellerApprovalTask)
	if err != nil {
		slog.Error("Error fetching approvalEvents", "error", err)
		time.Sleep(1 * time.Second)
		return err
	}

	if events == nil || len(events) == 0 {
		Token_seller_approval_task_RetryTimes++
		if Token_seller_approval_task_RetryTimes > 2 {
			tokenSellerApprovalTask.AlertConfig.SetLastBlockNumber(toBlock)
			config.SaveConfig(tokenSellerApprovalTask.ConfigFile, tokenSellerApprovalTask.AlertConfig.ConfigData)
			Token_seller_approval_task_RetryTimes = 0
			slog.Info("retry reached 3 times, no new approvalEvents found，update next time startblocknumber")
			return nil
		}
		slog.Info("no new approvalEvents found, start retry", "retryTimes", Token_seller_approval_task_RetryTimes)
		time.Sleep(1 * time.Second)
		return nil
	}

	for _, event := range events {
		jsonData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal approvalEvent: %w", err)
		}
		slog.Info("approvalEvent", "event", string(jsonData))

		err = utils.RedisSet(fmt.Sprintf("approval:%s:%s", event.TokenAddr, event.Owner), string(jsonData), 0)
		if err != nil {
			return fmt.Errorf("approvalEvent failed to save redis: %w", err)
		}
	}

	tokenSellerApprovalTask.AlertConfig.SetLastBlockNumber(toBlock)
	config.SaveConfig(tokenSellerApprovalTask.ConfigFile, tokenSellerApprovalTask.AlertConfig.ConfigData)
	slog.Info("end excute Token_seller_approval_task")
	return nil
}

type TokenMap map[string]TokenInfo

// TokenInfo 表示每个代币的详细信息
type TokenInfo struct {
	ID            int    `json:"id"`
	TokenName     string `json:"token_name"`
	TokenLogo     string `json:"token_logo"`
	TokenSymbol   string `json:"token_symbol"`
	TGEStartTime  string `json:"tge_start_time"`
	TGEEndTime    string `json:"tge_end_time"`
	AutoStartTime string `json:"auto_start_time"`
	AutoEndTime   string `json:"auto_end_time"`
	IsActive      bool   `json:"is_active"`
}

func GetApprovalEvents(tokenSellerTask *datadefine.TokenSellerTask) ([]datadefine.WBTCApproval, string, error) {
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

	// 解析approval事件
	exists, err := utils.RedisExists("token_configs")
	if err != nil {
		return nil, "", err
	}

	approvalEvents := make([]datadefine.WBTCApproval, 0)
	if exists {
		var tokenConfigs TokenMap
		err := utils.RedisGetObject("token_configs", &tokenConfigs)
		if err != nil {
			return nil, "", err
		}

		pAddress := common.HexToAddress(tokenSellerTask.Addr)
		for tokenAddr, tokenInfo := range tokenConfigs {
			if tokenInfo.IsActive {
				// 创建approval过滤器选项
				ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				filterOpts := &bind.FilterOpts{
					Start:   fromBlock,
					End:     &toBlock,
					Context: ctx,
				}
				tokenAddress := common.HexToAddress(tokenAddr)
				filterer, err := contracts.NewWBTCFilterer(tokenAddress, client)
				if err != nil {
					return nil, "", err
				}
				approval, err := filterer.FilterApproval(filterOpts, nil, []common.Address{pAddress})
				if err != nil {
					return nil, "", err
				}

				var currentBlock *types.Block = nil
				for approval.Next() {
					event := approval.Event

					// 多个vLog可能在同一个区块，每次处理的vLog的区块发生变化则重新获取新的区块信息保存到currentBlock
					if currentBlock == nil || event.Raw.BlockNumber != currentBlock.NumberU64() {
						// 设置连接超时时间为 5 秒
						ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						block, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(event.Raw.BlockNumber))
						if err != nil {
							return nil, "", err
						}
						currentBlock = block
					}

					var approvalEvent = datadefine.WBTCApproval{
						TokenAddr:      tokenAddress.Hex(),
						Owner:          event.Owner.Hex(),
						Spender:        event.Spender.Hex(),
						Value:          event.Value,
						TxHash:         event.Raw.TxHash.Hex(),
						BlockTimestamp: strconv.FormatUint(currentBlock.Time(), 10),
					}
					approvalEvents = append(approvalEvents, approvalEvent)
				}
			}
		}
	}

	toBlockStr := strconv.FormatUint(toBlock, 10)
	slog.Info("get approvalEvents success", "lenth", len(approvalEvents), "toBlock", toBlockStr)
	return approvalEvents, toBlockStr, nil

}
