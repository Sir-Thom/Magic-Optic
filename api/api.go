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
		rtmpStream.GET("/startStream", func(c *gin.Context) {
			// Create a channel to communicate the result
			resultCh := make(chan error)

			// Respond immediately indicating that the stream is starting
			c.JSON(http.StatusOK, gin.H{
				"message": "Starting RTMP stream...",
			})

			// Run StartRtmpStream asynchronously in a goroutine
			go func() {
				// This function runs in the background
				err := ffmpegCmd.StartRtmpStream()
				resultCh <- err
			}()

			// Wait for the result from the channel
			err := <-resultCh
			if err != nil {
				log.Printf("Error starting RTMP stream: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to start RTMP stream",
				})
				return
			}
		})

		rtmpStream.GET("/stopStream", func(c *gin.Context) {
			if err := ffmpegCmd.StopRtmpStream(); err != nil {
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
