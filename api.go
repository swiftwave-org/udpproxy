package main

import "github.com/labstack/echo/v4"

func newAPIServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.POST("/v1/proxy/add", addProxyAPI)
	e.POST("/v1/proxy/remove", removeProxyAPI)
	e.POST("/v1/proxy/exist", isExistProxyAPI)
	e.GET("/v1/proxy/list", listProxyAPI)
	return e
}

func addProxyAPI(c echo.Context) error {
	var pr ProxyRecord
	if err := c.Bind(&pr); err != nil {
		return &echo.HTTPError{Code: 400, Message: "Invalid request"}
	}

	// add the proxy
	success, err := proxyManager.Add(pr)
	if err != nil {
		c.JSON(200, AddProxyResponse{Success: false, Error: err.Error()})
	} else {
		c.JSON(200, AddProxyResponse{Success: success, Error: ""})
	}

	return nil
}

func removeProxyAPI(c echo.Context) error {
	var pr ProxyRecord
	if err := c.Bind(&pr); err != nil {
		return &echo.HTTPError{Code: 400, Message: "Invalid request"}
	}
	if success, err := proxyManager.Remove(pr); err != nil {
		c.JSON(200, RemoveProxyResponse{Success: false, Error: err.Error()})
	} else {
		c.JSON(200, RemoveProxyResponse{Success: success, Error: ""})
	}
	return nil
}

func isExistProxyAPI(c echo.Context) error {
	var pr ProxyRecord
	if err := c.Bind(&pr); err != nil {
		return &echo.HTTPError{Code: 400, Message: "Invalid request"}
	}
	exist := proxyManager.Exist(pr)
	c.JSON(200, IsExistProxyResponse{Exist: exist})
	return nil
}

func listProxyAPI(c echo.Context) error {
	proxies := proxyManager.List()
	c.JSON(200, proxies)
	return nil
}
