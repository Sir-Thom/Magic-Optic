package api

import (
	"Magic-optic/ffmpegCmd"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func Main() {
	router := gin.Default()
	device := router.Group("/device")
	{
		device.GET("/getDevices", func(c *gin.Context) {
			devices, err := listVideoDevices()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, devices)
		})
	}
	rtmpStream := router.Group("/rtmpStream")
	{
		rtmpStream.POST("/startStream", func(c *gin.Context) {
			// Run StartRtmpStream asynchronously in a goroutine
			go func() {
				_, err := ffmpegCmd.StartRtmpStream()
				if err != nil {
					log.Printf("Error starting RTMP stream: %v\n", err)
				}
			}()

			// Immediately respond with a 202 status code
			c.JSON(http.StatusAccepted, gin.H{
				"message": "Starting RTMP stream...",
			})
		})
		rtmpStream.GET("/checkStream", func(c *gin.Context) {
			if ffmpegCmd.IsRtmpStreamRunning() {
				log.Println(ffmpegCmd.IsRtmpStreamRunning())
				// The stream is running
				c.JSON(http.StatusOK, gin.H{
					"message": "RTMP stream is running",
				})
			} else {
				// The stream is not running
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "RTMP stream is not running",
				})
			}
		})

		rtmpStream.POST("/stopStream", func(c *gin.Context) {
			err := ffmpegCmd.StopRtmpStream()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "RTMP stream stopping...",
			})
		})
	}

	err := router.Run()
	if err != nil {
		log.Fatal(err)
	}
}
