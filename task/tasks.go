package task

import (
	"github.com/bamzi/jobrunner"
	"time"
	"token-seller/utils"
)

func StartTasks() {
	jobrunner.Start()
	jobrunner.Every(1000*time.Millisecond, utils.WrapJob("Token_seller_tokensold_task", Token_seller_tokensold_task))
	jobrunner.Every(1000*time.Millisecond, utils.WrapJob("Token_seller_approval_task", Token_seller_approval_task))
}
