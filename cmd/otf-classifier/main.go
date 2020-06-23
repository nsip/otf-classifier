package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	align "github.com/nsip/otf-classifier"
)

func main() {

	port := flag.Int("p", 1576, "port to run this server on")
	flag.Parse()

	os.Setenv("JAEGER_SERVICE_NAME", "OTF-CLASSIFIER")
	os.Setenv("JAEGER_SAMPLER_TYPE", "const")
	os.Setenv("JAEGER_SAMPLER_PARAM", "1")

	align.Init()
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
		return c.JSONPretty(http.StatusOK, ret, "  ")
	})
	e.GET("/index", func(c echo.Context) error {
		query := c.QueryParam("search")
		ret, err := align.Search(query)
		if err != nil {
			return err
		}
		return c.JSONPretty(http.StatusOK, ret, "  ")
	})

	// log.Println("Editor: localhost:1576")
	address := fmt.Sprintf(":%d", *port)
	e.Logger.Fatal(e.Start(address))
}
