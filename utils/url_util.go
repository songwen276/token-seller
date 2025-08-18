package utils

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// 全局随机源
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var BSC_URL_ARRAY = []string{
	// "http://192.168.100.101:8545",
	// "http://192.168.100.102:8545",
	"https://bsc-rpc.publicnode.com",
	// "http://127.0.0.1:8545/",
}

var ETH_URL_ARRAY = []string{
	"https://ethereum-rpc.publicnode.com",
	// "https://eth-mainnet.public.blastapi.io",
	// "https://gateway.tenderly.co/public/mainnet",
}

// SafeQueue 是线程安全的队列
type SafeQueue struct {
	items []string
	lock  sync.Mutex
}

func (q *SafeQueue) Put(item string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items = append(q.items, item)
}

func (q *SafeQueue) Get() string {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.items) == 0 {
		return ""
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *SafeQueue) Empty() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return len(q.items) == 0
}

// 用于存储不同类型的 URL 队列
var urlQueues sync.Map // key: string, value: *SafeQueue

// refillQueue 用于随机打乱并填充队列
func refillQueue(urlType string, urlList []string) {
	// 随机打乱 URL 顺序
	urls := append([]string(nil), urlList...) // 复制一份 URL 列表
	r.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	// 获取或创建 SafeQueue
	q, _ := urlQueues.LoadOrStore(urlType, &SafeQueue{})
	queue := q.(*SafeQueue)

	// 重新填充队列
	for _, url := range urls {
		queue.Put(url)
	}
}

// GetRpcURL 获取一个 URL，如果队列为空则重新填充
func GetRpcURL(urlType string) string {
	// 获取队列
	q, loaded := urlQueues.LoadOrStore(urlType, &SafeQueue{})
	if !loaded {
		log.Printf("Created new queue for %s", urlType)
	}
	queue := q.(*SafeQueue)

	// 如果队列为空，则重新填充
	if queue.Empty() {
		log.Printf("Queue for %s is empty, refilling...", urlType)
		switch urlType {
		case "BSC":
			refillQueue(urlType, BSC_URL_ARRAY)
		case "ETH":
			refillQueue(urlType, ETH_URL_ARRAY)
		default:
			log.Printf("Unknown URL type: %s", urlType)
			return ""
		}
	}

	// 获取并返回队列中的第一个 URL，同时从队列中剔除
	url := queue.Get()
	log.Printf("Retrieved URL %s from queue for %s", url, urlType)
	return url
}
