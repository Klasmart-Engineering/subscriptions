package api

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"subscriptions/src/monitoring"
)

func noContentOrLog(monitoringContext *monitoring.Context, ctx echo.Context, status int) {
	if err := ctx.NoContent(status); err != nil {
		monitoringContext.Error("Could not write no body response", zap.Error(err))
	}
}

func jsonContentOrLog(monitoringContext *monitoring.Context, ctx echo.Context, status int, body interface{}) {
	if err := ctx.JSON(status, body); err != nil {
		monitoringContext.Error("Could not write JSON response", zap.Error(err))
	}
}
