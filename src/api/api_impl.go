package api

import (
	"github.com/labstack/echo/v4"
	"subscriptions/src/monitoring"
)

type Impl struct{}

var Implementation = &Impl{}

//Your IDE should tell you here if you're not implementing all the endpoints
var _ ServerInterface = (*Impl)(nil)

func (Impl) GetHealthcheck(ctx echo.Context, monitoringContext *monitoring.Context) error {
	//TODO: Call DB
	err := ctx.JSON(200, ApplicationStateResponse{
		Up:      true,
		Details: "Successfully connected to the database",
	})
	if err != nil {
		return err
	}

	return nil
}

func (Impl) GetLiveness(ctx echo.Context, monitoringContext *monitoring.Context) error {
	err := ctx.JSON(200, ApplicationStateResponse{
		Up:      true,
		Details: "Application Up",
	})
	if err != nil {
		return err
	}

	return nil
}

func (i Impl) PostTestId(ctx echo.Context, monitoringContext *monitoring.Context, id int) error {
	//TODO implement me
	panic("implement me")
}
