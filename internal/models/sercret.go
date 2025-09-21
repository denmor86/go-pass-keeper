package models

import (
	pb "go-pass-keeper/pkg/proto"
	"time"
)

// SecretInfo - модель информации о секрете
type SecretInfo struct {
	ID      string
	Name    string
	Type    string
	Created time.Time
	Updated time.Time
}

func ToModel(pbSecrets *pb.GetSecretsResponse) []*SecretInfo {
	res := make([]*SecretInfo, len(pbSecrets.GetSecrets()))
	for _, s := range pbSecrets.GetSecrets() {
		res = append(res, &SecretInfo{
			ID:      s.GetId(),
			Name:    s.GetName(),
			Type:    s.GetType(),
			Created: s.GetCreated().AsTime(),
			Updated: s.GetUpdated().AsTime(),
		})
	}
	return res
}
