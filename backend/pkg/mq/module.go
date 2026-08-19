package mq

import "go.uber.org/fx"

var Module = fx.Module("mq",
	fx.Provide(NewNats),
	fx.Provide(fx.Annotate(NewProducer, fx.As(new(Publisher)))),
	fx.Provide(fx.Annotate(NewConsumer, fx.As(new(Subscriber)))),
)
