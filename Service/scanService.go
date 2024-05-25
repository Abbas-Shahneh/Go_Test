package Service

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os/exec"
	"portScan/Model"
	"portScan/Repository"
	"strconv"
	"sync"
)

type ScanService interface {
	ScanIPAddress(ipAddress string)
	GetScanResult(ipAddress string) (*Model.ScanResult, error)
}

type scanService struct {
	repo        Repository.ScanRepository
	scanRequest chan string
	wg          sync.WaitGroup
}

func NewScanService(repo Repository.ScanRepository, scanRequest chan string) ScanService {
	service := &scanService{repo: repo, scanRequest: scanRequest}

	// Worker pool
	const numWorkers = 5
	for i := 1; i <= numWorkers; i++ {
		service.wg.Add(1)
		go service.worker(i)
	}

	return service
}

func (s *scanService) ScanIPAddress(ipAddress string) {
	s.scanRequest <- ipAddress
}

func (s *scanService) GetScanResult(ipAddress string) (*Model.ScanResult, error) {
	return s.repo.FindByIP(ipAddress)
}

func (s *scanService) worker(id int) {

	defer s.wg.Done()

	for ipAddress := range s.scanRequest {
		log.Printf("Worker %d: Processing %s\n", id, ipAddress)

		var wg sync.WaitGroup
		portRanges := createPortRanges(1, 60, 10)
		//portRanges := []string{"1-170", "171-340", "341-510", "511-635"}
		results := make([]string, len(portRanges))
		resultLock := sync.Mutex{}
		for i, portRanges := range portRanges {
			wg.Add(1)
			go func(i int, portRange string) {
				defer wg.Done()
				log.Printf("Worker %d : %d : port range : %s", id, i, portRange)

				cmdCommand := fmt.Sprintf("nmap -p%s --open %s | grep -v 'unknown'", portRange, ipAddress)
				cmd := exec.Command("sh", "-c", cmdCommand)

				//cmd := exec.Command("nmap", "-sC", "-sV", ipAddress)
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("Worker %d: Failed to run nmap for %s: %v\n", id, ipAddress, err)
					return
				}
				resultLock.Lock()
				results[i] = string(output)
				resultLock.Unlock()
			}(i, portRanges)
		}
		wg.Wait()

		combinedResult := ""
		for _, res := range results {
			combinedResult += res
		}

		scanResult := &Model.ScanResult{
			ID:        primitive.NewObjectID(),
			IPAddress: ipAddress,
			Result:    combinedResult,
		}

		err := s.repo.Save(scanResult)
		if err != nil {
			log.Printf("Worker %d: Failed to save scan result for %s: %v\n", id, ipAddress, err)
			continue
		}
		log.Printf("Worker %d: Scan result saved for %s\n", id, ipAddress)
	}

	//for ipAddress := range s.scanRequest {
	//	log.Printf("Worker %d: Processing %s\n", id, ipAddress)
	//
	//	// Run NMap scan with the specified options
	//	cmd := exec.Command("nmap", "-sC", "-sV", ipAddress) // Scan with service and version detection
	//	output, err := cmd.CombinedOutput()
	//	if err != nil {
	//		log.Printf("Worker %d: Failed to run nmap for %s: %v\n", id, ipAddress, err)
	//		continue
	//	}
	//
	//	scanResult := &Model.ScanResult{
	//		ID:        primitive.NewObjectID(),
	//		IPAddress: ipAddress,
	//		Result:    string(output),
	//	}
	//
	//	err = s.repo.Save(scanResult)
	//	if err != nil {
	//		log.Printf("Worker %d: Failed to save scan result for %s: %v\n", id, ipAddress, err)
	//		continue
	//	}
	//	log.Printf("Worker %d: Scan result saved for %s\n", id, ipAddress)
	//}

}

func createPortRanges(start, end, step int) []string {
	ranges := []string{}
	for i := start; i <= end; i += step {
		rangeEnd := i + step - 1
		if rangeEnd > end {
			rangeEnd = end
		}
		ranges = append(ranges, strconv.Itoa(i)+"-"+strconv.Itoa(rangeEnd))
	}
	return ranges
}
