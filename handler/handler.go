package handler

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"sync"
	"time"
)

func GetEthClient(rpcUrl string, timeout time.Duration) (*ethclient.Client, error) {
	// 创建用于接收连接结果的通道
	clientCh := make(chan *ethclient.Client, 1)
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 尝试连接到 RPC 节点
		client, err := ethclient.Dial(rpcUrl)
		if err != nil {
			errCh <- err
			return
		}
		clientCh <- client
	}()

	// 创建一个带有超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 等待连接结果或超时
	select {
	case <-ctx.Done():
		// 超时处理
		wg.Wait() // 等待 goroutine 结束
		return nil, fmt.Errorf("连接超时，在 %v 内未能连接到节点", timeout)
	case err := <-errCh:
		// 连接出错
		return nil, err
	case client := <-clientCh:
		// 连接成功
		return client, nil
	}
}
