package main

import "github.com/labstack/echo/v4"

func newAPIServer() *echo.Echo {
	e := echo.New()
	e.POST("/v1//proxy/add", addProxyAPI)
	e.POST("/v1/proxy/remove", removeProxyAPI)
	e.POST("/v1/proxy/exist", isExistProxyAPI)
	return e
}

func addProxyAPI(c echo.Context) error {
	var pr ProxyRequest
	if err := c.Bind(&pr); err != nil {
		return &echo.HTTPError{Code: 400, Message: "Invalid request"}
	}

	return nil
}

func removeProxyAPI(c echo.Context) error {
	var pr ProxyRequest
	if err := c.Bind(&pr); err != nil {
		return &echo.HTTPError{Code: 400, Message: "Invalid request"}
	}
	return nil
}

func isExistProxyAPI(c echo.Context) error {
	var pr ProxyRequest
	if err := c.Bind(&pr); err != nil {
		return &echo.HTTPError{Code: 400, Message: "Invalid request"}
	}
	return nil
}
