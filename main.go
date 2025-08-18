package main

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"token-seller/task"
)

// TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

// func main() {
// 	// 初始化日志配置
// 	for i := 0; i < 10; i++ {
// 		println(utils.GetRpcURL("BSC"))
// 		println(utils.GetRpcURL("ETH"))
// 	}
// }

func main() {
	// 初始化日志配置
	setupLogger()

	task.StartTasks()
	select {}
}

// setupLogger 配置日志系统，使用 lumberjack 处理日志轮转
func setupLogger() {
	logDir := "./logs" // 日志目录
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// 动态生成日志文件名
	logFileName := fmt.Sprintf("%s/token-seller.log", logDir)

	// 配置日志切割
	log.SetOutput(&lumberjack.Logger{
		Filename:   logFileName, // 日志文件路径
		MaxSize:    5,           // 单个日志文件的最大大小（MB）
		MaxBackups: 20,          // 最多保留的旧日志文件数量
		MaxAge:     2,           // 日志文件保留的天数
		Compress:   true,        // 是否压缩旧日志
	})

	// 创建一个多写器，同时写入文件和控制台
	mw := io.MultiWriter(os.Stdout, os.Stderr, &lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    50,
		MaxBackups: 20,
		MaxAge:     2,
		Compress:   true,
	})
	// 设置日志输出到多写器
	log.SetOutput(mw)
	// 设置日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Logger initialized with file: %s", logFileName)
}
