package Repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"portScan/Model"
)

type ScanRepository interface {
	Save(scanResult *Model.ScanResult) error
	FindByIP(ipAddress string) (*Model.ScanResult, error)
}

type scanRepository struct {
	collection *mongo.Collection
}

func NewScanRepository(client *mongo.Client) ScanRepository {
	collection := client.Database("nmap").Collection("scans")
	return &scanRepository{collection}
}

func (r *scanRepository) Save(scanResult *Model.ScanResult) error {
	_, err := r.collection.InsertOne(context.TODO(), scanResult)
	return err
}

func (r *scanRepository) FindByIP(ipAddress string) (*Model.ScanResult, error) {
	var result Model.ScanResult
	filter := bson.M{"ipaddress": ipAddress}
	err := r.collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
