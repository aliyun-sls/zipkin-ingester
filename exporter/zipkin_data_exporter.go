package exporter

import (
	"bytes"
	"fmt"
	"github.com/aliyun-sls/zipkin-ingester/configure"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type ZipkinDataExporter interface {
	SendData(data []byte, sugar *zap.SugaredLogger) error
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

func (z zipkinDataExporterImpl) SendData(data []byte, sugar *zap.SugaredLogger) error {
	if data == nil || len(data) == 0 {
		return nil
	}

	req, err := http.NewRequest("POST", z.requestURL, bytes.NewBuffer(data))
	if err != nil {
		sugar.Warn("Failed to create request post")
		return err
	}

	req.Header.Set("x-sls-otel-project", z.configure.Project)
	req.Header.Set("x-sls-otel-instance-id", z.configure.Instance)
	req.Header.Set("x-sls-otel-ak-id", z.configure.AccessKey)
	req.Header.Set("x-sls-otel-ak-secret", z.configure.AccessSecret)

	client := &http.Client{}
	if resp, e := client.Do(req); e == nil {
		if resp.StatusCode != 200 {
			d, _ := io.ReadAll(resp.Body)
			sugar.Warn("Failed to send data", "StatusCode", resp.StatusCode, "requestURL", req.URL, "response body", string(d))
		} else {
			sugar.Info("Send data successfully")
		}
		return nil
	} else {
		sugar.Warn("Failed to send data", "StatusCode", resp.StatusCode, " ErrorMessage", resp.Header.Get("errorMessage"))
		return e
	}
}
