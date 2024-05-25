package Routes

import (
	"github.com/gin-gonic/gin"
	"portScan/Controller"
)

func SetupRouter(controller *Controller.ScanController) *gin.Engine {
	router := gin.Default()
	router.POST("/scan/:ipAddress", controller.HandleScanRequest)
	router.GET("/scan/result/:ipAddress", controller.GetScanResult)
	return router
}
