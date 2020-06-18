package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	align "github.com/nsip/otf-classifier"
)

func main() {

	os.Setenv("JAEGER_SERVICE_NAME", "OTF-CLASSIFIER")
	os.Setenv("JAEGER_SAMPLER_TYPE", "const")
	os.Setenv("JAEGER_SAMPLER_PARAM", "1")

	align.Init()
	e := echo.New()

	// Add Jaeger Tracer into Middleware
	c := jaegertracing.New(e, nil)
	defer c.Close()

	e.GET("/align", align.Align)
	e.GET("/index", func(c echo.Context) error {
		query := c.QueryParam("search")
		ret, err := align.Search(query)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, string(ret))
	})
	log.Println("Editor: localhost:1576")
	e.Logger.Fatal(e.Start(":1576"))
}
