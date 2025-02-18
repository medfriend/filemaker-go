package main

import (
	"context"
	"encoding/json"
	"filemaker-go/util"
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/medfriend/shared-commons-go/util/consul"
	"github.com/medfriend/shared-commons-go/util/env"
	"github.com/minio/minio-go/v7"
	"github.com/pebbe/zmq4"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	env.LoadEnv()

	consulClient := consul.ConnectToConsulKey("", "FILEMAKER")

	resultServiceInfo, minioClient := util.ConnectionsConsul(consulClient)

	zmqPort := resultServiceInfo["SERVICE_PORT"]
	zmpHost := resultServiceInfo["SERVICE_PATH"]

	socket, err := zmq4.NewSocket(zmq4.PULL)
	if err != nil {
		panic(err)
	}
	defer socket.Close()

	zmqConn := fmt.Sprintf("tcp://%s:%s", zmpHost, zmqPort)

	// Conectar al servidor PUSH
	err = socket.Connect(zmqConn)
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("Connected to %s", zmqConn))

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatalln(err)
	}

	// Escuchar mensajes

	for {
		msg, err := socket.Recv(0)
		if err != nil {
			fmt.Printf("Error recibiendo mensaje: %v\n", err)
			continue
		}

		// Mapa con los valores a reemplazar
		var msgObject map[string]string

		// Deserializar el JSON en el mapa
		err = json.Unmarshal([]byte(msg), &msgObject)

		var replace map[string]string
		err = json.Unmarshal([]byte(msgObject["message"]), &replace)

		if err != nil {
			return
		}

		bucketName := replace["bucketOrigen"]
		objectName := replace["plantilla"]

		// Obtener el objeto
		ctx := context.Background()
		object, err := minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
		if err != nil {
			log.Fatalln(err)
		}

		// Leer el contenido del objeto
		data, err := ioutil.ReadAll(object)
		if err != nil {
			log.Fatalln(err)
		}

		content := string(data)

		// Reemplazar las etiquetas con los valores del mapa
		for key, value := range replace {
			content = strings.ReplaceAll(content, key, value)
		}

		pdfg.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(content)))

		// Opciones de configuración para PDF, si es necesario
		pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)

		// Crear PDF
		err = pdfg.Create()
		if err != nil {
			log.Fatalln(err)
		}

		// Guardar el PDF en un archivo temporal
		tempFile, err := ioutil.TempFile(os.TempDir(), "output-*.pdf")
		if err != nil {
			log.Fatalln(err)
		}
		defer tempFile.Close()

		_, err = tempFile.Write(pdfg.Bytes())
		if err != nil {
			log.Fatalln(err)
		}

		objectName = fmt.Sprintf("%s-%s.pdf", msgObject["user"], msgObject["time"])

		_, err = minioClient.FPutObject(context.Background(), replace["bucketDestino"], objectName, tempFile.Name(), minio.PutObjectOptions{ContentType: "application/pdf"})
		if err != nil {
			log.Fatalln(err)
		}

		// Eliminar el archivo temporal
		err = os.Remove(tempFile.Name())
		if err != nil {
			log.Printf("Error eliminando el archivo temporal: %v\n", err)
		} else {
			log.Printf("Archivo temporal eliminado exitosamente\n")
		}

		fmt.Printf("PDF uploaded successfully: %s/%s\n", replace["bucketDestino"], objectName)
	}
}
