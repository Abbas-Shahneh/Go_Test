package Controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"portScan/Controller"
	"portScan/Model"
	"testing"
)

type MockScanService struct{}

func (m *MockScanService) ScanIPAddress(ipAddress string) {}

func (m *MockScanService) GetScanResult(ipAddress string) (*Model.ScanResult, error) {
	if ipAddress == "127.0.0.1" {
		return &Model.ScanResult{
			IPAddress: "127.0.0.1",
			Result:    "Scan result for 127.0.0.1",
		}, nil
	}
	return nil, fmt.Errorf("scan result not found")
}

type MockScanServiceWithError struct{}

func (m *MockScanServiceWithError) ScanIPAddress(ipAddress string) {}

func (m *MockScanServiceWithError) GetScanResult(ipAddress string) (*Model.ScanResult, error) {
	return nil, fmt.Errorf("error occurred while fetching scan result")
}

func TestScanController_HandleScanRequest(t *testing.T) {

	mockService := &MockScanService{}
	controller := Controller.NewScanController(mockService)

	req, _ := http.NewRequest("POST", "/scan/127.0.0.1", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "ipAddress", Value: "127.0.0.1"}}

	controller.HandleScanRequest(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanController_GetScanResult(t *testing.T) {

	mockService := &MockScanService{}
	controller := Controller.NewScanController(mockService)

	req, _ := http.NewRequest("GET", "/scan/result/127.0.0.1", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "ipAddress", Value: "127.0.0.1"}}

	controller.GetScanResult(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanController_GetScanResult_Error(t *testing.T) {

	mockService := &MockScanServiceWithError{}
	controller := Controller.NewScanController(mockService)

	req, _ := http.NewRequest("GET", "/scan/result/127.0.0.1", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "ipAddress", Value: "127.0.0.1"}}

	controller.GetScanResult(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestScanController_HandleScanRequest_EmptyIP(t *testing.T) {

	mockService := &MockScanService{}
	controller := Controller.NewScanController(mockService)

	req, _ := http.NewRequest("POST", "/scan/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	controller.HandleScanRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanController_HandleScanRequest_InvalidIP(t *testing.T) {

	mockService := &MockScanService{}
	controller := Controller.NewScanController(mockService)

	req, _ := http.NewRequest("POST", "/scan/invalid-ip", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "ipAddress", Value: "invalid-ip"}}

	controller.HandleScanRequest(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanController_GetScanResult_NotFound(t *testing.T) {

	mockService := &MockScanService{}
	controller := Controller.NewScanController(mockService)

	req, _ := http.NewRequest("GET", "/scan/result/192.168.0.1", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "ipAddress", Value: "192.168.0.1"}}

	controller.GetScanResult(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
