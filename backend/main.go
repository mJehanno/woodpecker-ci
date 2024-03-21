package main

import (
	"flag"
	"net"
	"os"
	"woodpecker-ci/db"
	"woodpecker-ci/handlers"
	"woodpecker-ci/logger"

	"github.com/labstack/echo/middleware"

	"github.com/labstack/echo"
)

func main() {
	var socketPath string
	flag.StringVar(&socketPath, "socket", "/run/guest-services/backend.sock", "Unix domain socket to listen on")
	flag.Parse()

	_ = os.RemoveAll(socketPath)

	logger := logger.GetLogger()

	err := db.CreateDB()
	if err != nil {
		logger.WithError(err).Fatal("failed to create database")
	}

	logMiddleware := middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}",` +
			`"method":"${method}","uri":"${uri}",` +
			`"status":${status},"error":"${error}"` +
			`}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
		Output:           logger.Writer(),
	})

	logger.Infof("Starting listening on %s\n", socketPath)
	router := echo.New()
	router.HideBanner = true
	router.Use(logMiddleware)
	startURL := ""

	ln, err := listen(socketPath)
	if err != nil {
		logger.Fatal(err)
	}
	router.Listener = ln

	router.POST("/api/upload", handlers.Upload)
	router.GET("/api/pipeline", handlers.GetPipeline)
	router.POST("/api/pipeline/lint", handlers.LintPipeline)

	logger.Fatal(router.Start(startURL))
}

func listen(path string) (net.Listener, error) {
	return net.Listen("unix", path)
}
