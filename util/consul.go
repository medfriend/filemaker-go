package util

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/medfriend/shared-commons-go/util/consul"
	"github.com/medfriend/shared-commons-go/util/global"
	minio2 "github.com/medfriend/shared-commons-go/util/minio"
	"github.com/minio/minio-go/v7"
)

func ConnectionsConsul(consulClient *api.Client) (map[string]string, *minio.Client, string) {
	serviceInfo, _ := consul.GetKeyValue(consulClient, "FILEMAKER")
	rabbitInfo, _ := consul.GetKeyValue(consulClient, "RABBIT")

	minioclient := minio2.CONN(consulClient)
	var resultRabbitmq map[string]string

	err := json.Unmarshal([]byte(rabbitInfo), &resultRabbitmq)

	if err != nil {
	}

	s := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		resultRabbitmq["RABBIT_USER"],
		resultRabbitmq["RABBIT_PASSWORD"],
		resultRabbitmq["RABBIT_HOST"],
		resultRabbitmq["RABBIT_PORT"])

	global.SetRabbitConn(s)

	var resultServiceInfo map[string]string
	json.Unmarshal([]byte(serviceInfo), &resultServiceInfo)

	return resultServiceInfo, minioclient, s
}
