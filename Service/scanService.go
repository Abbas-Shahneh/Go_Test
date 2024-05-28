package Service

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os/exec"
	"portScan/Model"
	"portScan/Repository"
	"sync"
)

type ScanService interface {
	ScanIPAddress(ipAddress string)
	GetScanResult(ipAddress string) (*Model.ScanResult, error)
	Shutdown() error
}

type scanService struct {
	repo         Repository.ScanRepository
	scanRequest  chan string
	wg           sync.WaitGroup
	shutdownCh   chan struct{}
	mu           sync.Mutex
	shuttingDown bool
}

func NewScanService(repo Repository.ScanRepository, scanRequest chan string) ScanService {
	service := &scanService{
		repo:        repo,
		scanRequest: scanRequest,
		shutdownCh:  make(chan struct{}),
	}

	const numWorkers = 5
	for i := 1; i <= numWorkers; i++ {
		service.wg.Add(1)
		go service.worker(i)
	}

	return service
}

func (s *scanService) ScanIPAddress(ipAddress string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.shuttingDown {
		s.scanRequest <- ipAddress
	}
}

func (s *scanService) GetScanResult(ipAddress string) (*Model.ScanResult, error) {
	return s.repo.FindByIP(ipAddress)
}

func (s *scanService) Shutdown() error {
	s.mu.Lock()
	s.shuttingDown = true
	close(s.scanRequest)
	s.mu.Unlock()

	s.wg.Wait()
	return nil
}

func (s *scanService) worker(id int) {
	defer s.wg.Done()

	for ipAddress := range s.scanRequest {
		s.processScan(id, ipAddress)
	}
}

func (s *scanService) processScan(id int, ipAddress string) {
	log.Printf("Worker %d: Processing %s\n", id, ipAddress)

	portRanges := createPortRanges(1, 65535, 6000)
	results := s.scanPorts(ipAddress, portRanges, id)

	combinedResult := ""
	for _, res := range results {
		combinedResult += res
	}

	scanResult := &Model.ScanResult{
		ID:        primitive.NewObjectID(),
		IPAddress: ipAddress,
		Result:    combinedResult,
	}

	if err := s.repo.Save(scanResult); err != nil {
		log.Printf("Worker %d: Failed to save scan result for %s: %v\n", id, ipAddress, err)
	} else {
		log.Printf("Worker %d: Scan result saved for %s\n", id, ipAddress)
	}
}

func (s *scanService) scanPorts(ipAddress string, portRanges []string, workerID int) []string {
	var wg sync.WaitGroup
	results := make([]string, len(portRanges))
	resultLock := sync.Mutex{}

	for i, portRange := range portRanges {
		wg.Add(1)
		go func(i int, portRange string) {
			defer wg.Done()
			log.Printf("Worker %d : %d : port range : %s", workerID, i, portRange)

			cmd := exec.Command("sh", "-c", fmt.Sprintf("nmap -p%s --open %s | grep -v 'unknown'", portRange, ipAddress))
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Worker %d: Failed to run nmap for %s: %v\n", workerID, ipAddress, err)
				return
			}
			resultLock.Lock()
			results[i] = string(output)
			resultLock.Unlock()
		}(i, portRange)
	}
	wg.Wait()

	return results
}

func createPortRanges(start, end, step int) []string {
	var ranges []string
	for i := start; i <= end; i += step {
		rangeEnd := i + step - 1
		if rangeEnd > end {
			rangeEnd = end
		}
		ranges = append(ranges, fmt.Sprintf("%d-%d", i, rangeEnd))
	}
	return ranges
}
