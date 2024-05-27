package api

import (
	"Magic-optic/ffmpegCmd"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"runtime"
)

func Main() {
	router := gin.Default()
	streamManager := ffmpegCmd.NewStreamManager()
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

	codecs := router.Group("/codecs")
	{
		codecs.GET("/getAudioCodecs", func(c *gin.Context) {
			audioCodecs, err := getAudioCodecs()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, audioCodecs)

		})
	}

	stream := router.Group("/stream")
	{
		stream.POST("startStream", func(c *gin.Context) {
			var configMap map[string]interface{}
			if err := c.ShouldBindJSON(&configMap); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			configJSON, err := json.Marshal(configMap)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			var streamConfig ffmpegCmd.StreamConfig
			switch streamType := configMap["streamType"].(string); streamType {
			case "rtsp":
				var rtspConfig ffmpegCmd.RtspConfig
				if err := json.Unmarshal(configJSON, &rtspConfig); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				streamConfig = rtspConfig
			case "rtmp":
				var rtmpConfig ffmpegCmd.RtmpConfig
				if err := json.Unmarshal(configJSON, &rtmpConfig); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				streamConfig = rtmpConfig
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported stream type"})
				return
			}

			id, _, err := streamManager.StartStream(streamConfig)
			if err != nil {
				log.Printf("Error starting stream: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusAccepted, gin.H{"message": "Starting stream...", "streamID": id})
		})

		stream.GET("/checkAllStream", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": streamManager.CheckAllStream(),
			})
		})

		stream.GET("/checkStream/:id", func(c *gin.Context) {
			id := c.Param("id")
			if streamManager.IsStreamRunning(id) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Stream is running",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Stream is not running",
				})
			}
		})

		stream.POST("/stopStream/:id", func(c *gin.Context) {
			id := c.Param("id")
			err := streamManager.StopStream(id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Stream stopping...",
			})
		})

		router.GET("/debug", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": runtime.NumGoroutine(),
			})
			log.Println(runtime.NumGoroutine())
		})
	}

	err := router.Run()
	if err != nil {
		log.Fatal(err)
	}
}
