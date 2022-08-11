package main

import (
	"private/rat/graph"

	"private/rat/errors"
)

func run() error {
	g, err := graph.Init("notes", "./content")
	if err != nil {
		return errors.Wrap(err, "failed to init graph")
	}

	g.Print()

	n, err := g.Get("notes/tools")
	if err != nil {
		return errors.Wrap(err, "failed to get node")
	}

	_, err = n.Add("commit-script")
	if err != nil {
		return errors.Wrap(err, "failed to add node")
	}

	g.Print()

	return nil
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}

	// conf, err := config.Load("./config.json")
	// if err != nil {
	// 	panic(err)
	// }

	// router := gin.Default()

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
