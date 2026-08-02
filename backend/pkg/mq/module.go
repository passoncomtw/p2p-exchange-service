package mq

import (
	"go.uber.org/fx"
)

var Module = fx.Module("mq", fx.Provide(NewNats))
