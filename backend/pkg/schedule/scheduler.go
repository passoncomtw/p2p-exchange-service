package schedule

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/fx"
)

// Params 透過 fx group 收集所有已注冊的 Runner。
type Params struct {
	fx.In

	Runners []Runner `group:"schedule_runner"`
}

// Scheduler 統一管理所有定時任務的生命週期。
type Scheduler struct {
	runners []Runner
}

func New(p Params, lc fx.Lifecycle) *Scheduler {
	s := &Scheduler{runners: p.Runners}

	var cancel context.CancelFunc

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			ctx, c := context.WithCancel(context.Background())
			cancel = c
			for _, r := range s.runners {
				runner := r
				logx.Infof("schedule: starting %s", runner.Name())
				go runner.Run(ctx)
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			if cancel != nil {
				cancel()
			}
			logx.Info("schedule: all runners stopped")
			return nil
		},
	})

	return s
}
