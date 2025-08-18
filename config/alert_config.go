package config

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"log/slog"
	"os"
	"sync"
)

// 配置文件结构
type AlertConfigFile struct {
	LastBlockNumber string `json:"lastBlockNumber"` // 上次处理的区块号
	LogIndexs       []uint `json:"logIndexs"`       // 当前已处理的交易哈希列表
}

// 加载配置文件
func LoadConfig(filePath string, initConfigFile *AlertConfigFile) *AlertConfigFile {
	file, err := os.Open(filePath)
	if err != nil && initConfigFile != nil {
		slog.Error("Error opening config file, using default config", "error", err)
		SaveConfig(filePath, initConfigFile)
		return nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var config AlertConfigFile
	err = decoder.Decode(&config)
	if err != nil {
		slog.Error("Error decoding config data", "error", err)
		return nil
	}

	return &config
}

var configMutex sync.Mutex

// 保存配置文件
func SaveConfig(filePath string, config *AlertConfigFile) {
	configMutex.Lock()
	defer configMutex.Unlock()
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("Error creating config file", "error", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 格式化输出
	err = encoder.Encode(config)
	if err != nil {
		slog.Error("Error encoding config data", "error", err)
	}
}

// 监控配置文件变化
func WatchConfig(filePath string, config *AlertConfig) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Failed to create watcher", "error", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(filePath)
	if err != nil {
		slog.Error("Failed to add config file to watcher", "error", err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				configMutex.Lock()
				slog.Info("Config file modified, reloading...")
				config.ConfigData = LoadConfig(filePath, nil) // 配置文件修改时重新加载
				configMutex.Unlock()

			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("Watcher error", "error", err)
		}
	}
}

type AlertConfig struct {
	ConfigData  *AlertConfigFile // 全局配置数据
	ConfigMutex sync.RWMutex     // 配置读写锁      // 限制 BTC 价格
}

// 获取上次处理的区块号
func (cfg *AlertConfig) GetLastBlockNumber() string {
	cfg.ConfigMutex.RLock()
	defer cfg.ConfigMutex.RUnlock()
	return cfg.ConfigData.LastBlockNumber
}

// 获取上次区块中已处理的日志索引列表
func (cfg *AlertConfig) GetLogIndexs() []uint {
	cfg.ConfigMutex.RLock()
	defer cfg.ConfigMutex.RUnlock()
	return cfg.ConfigData.LogIndexs
}

// 更新上次处理的区块号
func (cfg *AlertConfig) SetLastBlockNumber(blockNumber string) {
	cfg.ConfigMutex.Lock()
	defer cfg.ConfigMutex.Unlock()
	cfg.ConfigData.LastBlockNumber = blockNumber
}

// 更新当前已处理的交易哈希列表
func (cfg *AlertConfig) SetLogIndexs(logIndexs []uint) {
	cfg.ConfigMutex.Lock()
	defer cfg.ConfigMutex.Unlock()
	cfg.ConfigData.LogIndexs = logIndexs
}
