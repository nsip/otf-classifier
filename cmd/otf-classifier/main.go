package main

import (
	"github.com/namsral/flag"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	align "github.com/nsip/otf-classifier"
)

func main() {

    // TODO: Support ENV port
	port := flag.Int("port", 1576, "port to run this server on")
	configPath := flag.String("config", "./curricula", "Path to config files")
	flag.Parse()

    // TODO: Allow override with port??? And does Tracer support other ENV directly?
	os.Setenv("JAEGER_SERVICE_NAME", "OTF-CLASSIFIER")
	os.Setenv("JAEGER_SAMPLER_TYPE", "const")
	os.Setenv("JAEGER_SAMPLER_PARAM", "1")

	align.Init(configPath)
	e := echo.New()

	// Add Jaeger Tracer into Middleware
	c := jaegertracing.New(e, nil)
	defer c.Close()

	e.POST("/align", align.Align) // needs to be available as post to support json payloads
	e.GET("/align", align.Align)
	e.GET("/lookup", func(c echo.Context) error {
		query := c.QueryParam("search")
		ret, err := align.Lookup(query)
		if err != nil {
			return c.String(http.StatusNotFound, err.Error())
		}
		return c.JSON(http.StatusOK, ret)
	})
	e.GET("/index", func(c echo.Context) error {
		query := c.QueryParam("search")
		ret, err := align.Search(query)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, ret)
	})

	// log.Println("Editor: localhost:1576")
	address := fmt.Sprintf(":%d", *port)
	e.Logger.Fatal(e.Start(address))
}
