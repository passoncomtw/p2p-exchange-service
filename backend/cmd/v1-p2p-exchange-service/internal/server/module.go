package server

import (
	"github.com/zeromicro/go-zero/rest"
	"go.uber.org/fx"
)

var Module = fx.Module("server",
	fx.Provide(NewServer),
	fx.Invoke(func(*rest.Server) {}), // 強制實例化
)
