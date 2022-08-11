package main

import (
	"fmt"
	"private/notes/config"
	"private/notes/graph"

	"github.com/gin-gonic/gin"
)

func main() {

	conf, err := config.Load("./config.json")
	if err != nil {
		panic(err)
	}

	g, err := graph.Load("notes")
	if err != nil {
		panic(err)
	}

	g.Print()

	router := gin.Default()

	router.GET(
		"/*path",
		func(c *gin.Context) {
			path := c.Param("path")

			if len(path) > 0 && path[0] == '/' {
				path = path[1:]
			}

			n, err := g.Get(path)
			if err != nil {
				c.JSON(404, gin.H{
					"error": err.Error(),
				})

				return
			}

			c.JSON(200, gin.H{
				"path": path,
				"body": n.Node().Body(),
			})
		},
	)

	err = router.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		panic(err)
	}

	// n, err := g.Get(g.Root())
	// if err != nil {
	// 	panic(err)
	// }

	// _, err = n.Add("docs")
	// if err != nil {
	// 	panic(err)
	// }

	// n.Add("test")

}
