package api

import (
	"github.com/labstack/echo/v4"
	"subscriptions/src/utils"
)

type Impl struct{}

var Implementation *Impl = &Impl{}

//Your IDE should tell you here if you're not implementing all of the endpoints
var _ ServerInterface = (*Impl)(nil)

//Ideally we would generate & implement a second interface which provides
//MonitoringContext rather than echo.Context and returns the return type, error tuple

func (Impl) GetHealthcheck(ctx echo.Context) error {
	//TODO: Call DB
	err := ctx.JSON(200, ApplicationStateResponse{
		Up:      true,
		Details: utils.StringPtr("Successfully connected to the database"),
	})
	if err != nil {
		return err
	}

	return nil
}

func (Impl) GetLiveness(ctx echo.Context) error {
	err := ctx.JSON(200, ApplicationStateResponse{
		Up:      true,
		Details: utils.StringPtr("Application Up"),
	})
	if err != nil {
		return err
	}

	return nil
}
