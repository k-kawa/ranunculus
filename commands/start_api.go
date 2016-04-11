package commands

import (
	"golang.org/x/net/context"
	"github.com/gin-gonic/gin"
)


func StartApi(ctx context.Context) {
	r := gin.Default()
	r.POST("/queues", CreateQueue)
	r.GET("/queues/:id", GetQueue)
	r.POST("/queues/:id/jobs", PostJob)
	r.POST("/queues/:id/results", GetResult)
	r.Run(":3000")
}

func CreateQueue(c *gin.Context) {

}

func GetQueue(c *gin.Context) {
	jobId := c.Param("id")

	c.JSON(200, gin.H{
		"Hoge": "Fuga",
		"JobId": jobId,
	})
}

func PostJob(c *gin.Context) {

}

func GetResult(c *gin.Context) {

}