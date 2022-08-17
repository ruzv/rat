package main

import (
	"fmt"
	"private/rat/config"
	"private/rat/handler/graphhttp"

	"github.com/gin-gonic/gin"
)

func main() {
	err := run()
	if err != nil {
		panic(err)
	}

	// router.GET(
	// 	"/*path",
	// 	func(c *gin.Context) {
	// 		path := c.Param("path")

	// 		if len(path) > 0 && path[0] == '/' {
	// 			path = path[1:]
	// 		}

	// 		n, err := g.Get(path)
	// 		if err != nil {
	// 			c.JSON(404, gin.H{
	// 				"error": err.Error(),
	// 			})

	// 			return
	// 		}

	// 		c.JSON(200, gin.H{
	// 			"path": path,
	// 			// "body": n.Node().Body(),
	// 		})
	// 	},
	// )

	// err = router.Run(fmt.Sprintf(":%d", conf.Port))
	// if err != nil {
	// 	panic(err)
	// }

}

func run() error {
	conf, err := config.Load("./config.json")
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	graphhttp.RegisterRoutes(conf, router.RouterGroup)

	err = router.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		panic(err)
	}

	return nil
}
