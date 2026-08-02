package notificationmodule

import (
	appnotification "p2p-exchange/pkg/notification_module/app_notification"
	wsnotification "p2p-exchange/pkg/notification_module/ws_notification"

	"go.uber.org/fx"
)

var Module = fx.Module("notificationmodule",
	fx.Provide(wsnotification.NewHub),
	fx.Provide(wsnotification.NewWebsocket),
	fx.Provide(appnotification.NewNotifier),
)
