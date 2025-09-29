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

// ToProtoMetadata - метод конвертирует информацию в метаданные
func (i *SecretInfo) ToProtoMetadata() *pb.SecretMetadata {
	return &pb.SecretMetadata{
		Id:   i.ID,
		Name: i.Name,
		Type: i.Type,
	}
}

func SecretInfoFromProtoMetadata(meta *pb.SecretMetadata) *SecretInfo {
	return &SecretInfo{
		ID:      meta.GetId(),
		Name:    meta.GetName(),
		Type:    meta.GetType(),
		Created: meta.GetCreated().AsTime(),
		Updated: meta.GetUpdated().AsTime(),
	}
}

func SecretsResponseToSecretInfo(pbSecrets *pb.GetSecretsResponse) []*SecretInfo {
	pbSecretsList := pbSecrets.GetSecrets()
	res := make([]*SecretInfo, 0, len(pbSecretsList))

	for _, s := range pbSecretsList {
		res = append(res, SecretInfoFromProtoMetadata(s))
	}
	return res
}
