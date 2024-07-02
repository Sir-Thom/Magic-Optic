package api

import (
	"Magic-optic/api/utils"
	"Magic-optic/stream"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"runtime"
)

func Main() {
	router := gin.Default()
	streamManager := stream.NewStreamManager()
	device := router.Group("/device")
	{
		device.GET("", func(c *gin.Context) {
			devices, err := utils.ListVideoDevices()
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
		codecs.GET("/audio", func(c *gin.Context) {
			audioCodecs, err := utils.GetAudioCodecs()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, audioCodecs)
		})

		codecs.GET("/video", func(c *gin.Context) {
			videoCodecs, err := utils.GeVideoCodecs()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, videoCodecs)
		})
	}

	streaming := router.Group("/stream")
	{
		streaming.POST("startStream", func(c *gin.Context) {
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

			var streamConfig stream.ConfigStream
			switch streamType := configMap["streamType"].(string); streamType {
			case "rtsp":
				var rtspConfig stream.RtspConfig
				if err := json.Unmarshal(configJSON, &rtspConfig); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				streamConfig = rtspConfig
			case "rtmp":
				var rtmpConfig stream.RtmpConfig
				if err := json.Unmarshal(configJSON, &rtmpConfig); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				streamConfig = rtmpConfig
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported stream type"})
				return
			}
			// Check if the stream is for Raspberry Pi
			raspberrypi, _ := configMap["raspberrypi"].(bool)
			id, _, err := streamManager.StartStream(streamConfig, raspberrypi)
			if err != nil {
				log.Printf("Error starting stream: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusAccepted, gin.H{"message": "Starting stream...", "streamID": id})
		})

		streaming.GET("/checkAllStream", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": streamManager.CheckAllStream(),
			})
		})

		streaming.GET("/checkStream/:id", func(c *gin.Context) {
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

		streaming.POST("/stopStream/:id", func(c *gin.Context) {
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

		streaming.GET("/debug", func(c *gin.Context) {
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
