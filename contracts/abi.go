package contracts

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"io/ioutil"
	"path/filepath"
)

func LoadABI(filePath string) (abi.ABI, error) {
	// 规范化路径，确保 Windows/Linux 都能正确解析
	normalizedPath := filepath.FromSlash(filePath)

	// 读取文件
	abiJson, err := ioutil.ReadFile(normalizedPath)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to read ABI file: %s, error: %v", normalizedPath, err)
	}

	// 解析 ABI
	parsedABI, err := abi.JSON(bytes.NewReader(abiJson))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse ABI file: %s, error: %v", normalizedPath, err)
	}

	return parsedABI, nil
}
