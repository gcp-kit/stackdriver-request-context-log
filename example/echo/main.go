package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gcp-kit/stalog"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// Set request handler
	e.GET("/", func(c echo.Context) error {
		// Get request context logger
		logger := stalog.RequestContextLogger(c.Request())

		// These logs are grouped with the request log
		logger.Debugf("Hi")
		logger.Infof("Hello")
		logger.Warnf("World")

		return c.String(http.StatusOK, "OK")
	})

	projectId := "my-gcp-project"

	// Make config for this library
	config := stalog.NewConfig(projectId)
	config.RequestLogOut = os.Stderr               // request log to stderr
	config.ContextLogOut = os.Stdout               // context log to stdout
	config.Severity = stalog.SeverityInfo          // only over INFO logs are logged
	config.AdditionalData = stalog.AdditionalData{ // set additional fields for all logs
		"service": "foo",
		"version": 1.0,
	}

	// Set middleware for the request log to be automatically logged
	e.Use(stalog.RequestLoggingWithEcho(config))

	// Run server
	fmt.Println("Waiting requests on port 8080...")
	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}
