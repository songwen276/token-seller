package datadefine

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"math/big"
	"token-seller/config"
)

type TokenSellerTask struct {
	Addr        string
	AlertConfig *config.AlertConfig
	AlertType   string
	ConfigFile  string
	PairAbi     abi.ABI
}

type TokensSoldEvent struct {
	Token          string   `json:"Token"`
	Seller         string   `json:"Seller"`
	AmountIn       *big.Int `json:"AmountIn"`
	AmountOut      *big.Int `json:"AmountOut"`
	UsedV3         bool     `json:"UsedV3"`
	Fee            *big.Int `json:"Fee"`
	TxHash         string   `json:"TxHash"`
	BlockTimestamp string   `json:"BlockTimestamp"`
}

type WBTCApproval struct {
	TokenAddr      string   `json:"TokenAddr"`
	Owner          string   `json:"Owner"`
	Spender        string   `json:"Spender"`
	Value          *big.Int `json:"Value"`
	TxHash         string   `json:"TxHash"`
	BlockTimestamp string   `json:"BlockTimestamp"`
}
