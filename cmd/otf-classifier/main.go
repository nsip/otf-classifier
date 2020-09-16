package main

import (
	"github.com/jnovack/flag"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	align "github.com/nsip/otf-classifier"
)

func main() {
    var port int
    var curriculaPath string
	flag.IntVar(&port, "port", 1576, "port to run this server on")
    // flag.StringVar(&name, "name", "", "help message")
	flag.StringVar(&curriculaPath, "curriculapath", "./curricula", "Path to curricula files")
	flag.Parse()

    // TODO: Allow override with port??? And does Tracer support other ENV directly?
	os.Setenv("JAEGER_SERVICE_NAME", "OTF-CLASSIFIER")
	os.Setenv("JAEGER_SAMPLER_TYPE", "const")
	os.Setenv("JAEGER_SAMPLER_PARAM", "1")

	align.Init(curriculaPath)
	e := echo.New()

	// Add Jaeger Tracer into Middleware
	c := jaegertracing.New(e, nil)
	defer c.Close()

    // Enable metrics middleware
    p := prometheus.NewPrometheus("echo", nil)
    p.Use(e)

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
	address := fmt.Sprintf(":%d", port)
	e.Logger.Fatal(e.Start(address))
}
