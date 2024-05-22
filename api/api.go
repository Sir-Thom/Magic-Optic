package api

import (
	"Magic-optic/ffmpegCmd" // Import the ffmpegCmd package
	"encoding/json"
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
	stream := router.Group("/stream")
	{
		stream.POST("/startStream", func(c *gin.Context) {
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

			var streamConfig interface{}
			switch streamType := configMap["streamType"].(string); streamType {
			case "hls":
				var hlsCfg ffmpegCmd.HlsConfig
				if err := json.Unmarshal(configJSON, &hlsCfg); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				streamConfig = hlsCfg
			case "rtmp":
				var rtmpCfg ffmpegCmd.RtmpConfig
				if err := json.Unmarshal(configJSON, &rtmpCfg); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				streamConfig = rtmpCfg
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported stream type"})
				return
			}

			go func() {
				switch cfg := streamConfig.(type) {
				case ffmpegCmd.HlsConfig:
					_, err := ffmpegCmd.StartHlsStream(cfg)
					if err != nil {
						log.Printf("Error starting HLS stream: %v\n", err)
					}
				case ffmpegCmd.RtmpConfig:
					_, err := ffmpegCmd.StartRtmpStream(cfg)
					if err != nil {
						log.Printf("Error starting RTMP stream: %v\n", err)
					}
				}
			}()

			c.JSON(http.StatusAccepted, gin.H{"message": "Starting stream..."})
		})

		stream.GET("/checkStream", func(c *gin.Context) {
			if ffmpegCmd.IsStreamRunning() {
				c.JSON(http.StatusOK, gin.H{
					"message": "Stream is running",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Stream is not running",
				})
			}
		})

		stream.POST("/stopStream", func(c *gin.Context) {
			err := ffmpegCmd.StopStream()
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
	}
	err := router.Run()
	if err != nil {
		log.Fatal(err)
	}
}
