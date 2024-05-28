package Controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"portScan/Service"
)

type ScanController struct {
	service Service.ScanService
}

func NewScanController(service Service.ScanService) *ScanController {
	return &ScanController{service: service}
}

func (sc *ScanController) HandleScanRequest(c *gin.Context) {
	ipAddress := c.Param("ipAddress")
	sc.service.ScanIPAddress(ipAddress)
	c.Status(http.StatusAccepted)
}

func (sc *ScanController) GetScanResult(c *gin.Context) {
	ipAddress := c.Param("ipAddress")
	result, err := sc.service.GetScanResult(ipAddress)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scan result not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}
