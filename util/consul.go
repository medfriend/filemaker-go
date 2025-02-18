package util

import (
	"encoding/json"
	"github.com/hashicorp/consul/api"
	"github.com/medfriend/shared-commons-go/util/consul"
	minio2 "github.com/medfriend/shared-commons-go/util/minio"
	"github.com/minio/minio-go/v7"
)

func ConnectionsConsul(consulClient *api.Client) (map[string]string, *minio.Client) {
	serviceInfo, _ := consul.GetKeyValue(consulClient, "FILEMAKER")
	minioclient := minio2.CONN(consulClient)

	var resultServiceInfo map[string]string
	json.Unmarshal([]byte(serviceInfo), &resultServiceInfo)

	return resultServiceInfo, minioclient
}
