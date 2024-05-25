package Model

import "go.mongodb.org/mongo-driver/bson/primitive"

type ScanResult struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	IPAddress string             `bson:"ipaddress"`
	Result    string             `bson:"result"`
}
