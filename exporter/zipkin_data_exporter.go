package exporter

import (
	"bytes"
	"fmt"
	"github.com/aliyun-sls/zipkin-ingester/configure"
	"log"
	"net/http"
)

type ZipkinDataExporter interface {
	SendData(data []byte) error
}

func NewZipkinExporter(c *configure.Configuration) ZipkinDataExporter {
	return &zipkinDataExporterImpl{
		requestURL: fmt.Sprintf("https://%s.%s/zipkin/api/v2/spans", c.Project, c.Endpoint),
		configure:  c,
	}
}

type zipkinDataExporterImpl struct {
	requestURL string
	configure  *configure.Configuration
}

func (z zipkinDataExporterImpl) SendData(data []byte) error {
	if data == nil || len(data) == 0 {
		return nil
	}

	req, err := http.NewRequest("POST", z.requestURL, bytes.NewBuffer(data))
	if err != nil {
		log.Println("Failed to create request post")
		return err
	}

	req.Header.Set("x-sls-otel-project", z.configure.Project)
	req.Header.Set("x-sls-otel-instance-id", z.configure.Instance)
	req.Header.Set("x-sls-otel-ak-id", z.configure.AccessKey)
	req.Header.Set("x-sls-otel-ak-secret", z.configure.AccessSecret)

	client := &http.Client{}
	if resp, err := client.Do(req); err == nil {
		return nil
	} else {
		log.Println(fmt.Sprintf("Failed to send data. StatuCode:%s, ErrorMessage: %s", resp.StatusCode, resp.Header.Get("errorMessage")))
		return err
	}
}
