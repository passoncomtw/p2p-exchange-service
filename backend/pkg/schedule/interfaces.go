package schedule

import "context"

// Runner 代表一個定時背景任務。
// Run 應在內部自行維護 loop，當 ctx.Done() 時結束並返回。
type Runner interface {
	Name() string
	Run(ctx context.Context)
}
