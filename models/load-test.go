package models

import (
	"encoding/json"
	"strconv"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/periodic"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// LoadGenerator - represents the load generator type
type LoadGenerator string

const (
	// FortioLG - represents the Fortio load generator
	FortioLG LoadGenerator = "fortio"

	// Wrk2LG - represents the wrk2 load generator
	Wrk2LG LoadGenerator = "wrk2"
)

// Name - retrieves a string value for the generator
func (l LoadGenerator) Name() string {
	return string(l)
}

// LoadTestOptions represents the load test options
type LoadTestOptions struct {
	Name string
	URL  string

	HTTPQPS float64

	HTTPNumThreads int

	IsInsecure bool
	Duration   time.Duration

	LoadGenerator LoadGenerator

	Cert, Key, CACert string

	AllowInitialErrors bool

	IsGRPC           bool
	GRPCStreamsCount int
	GRPCDoHealth     bool
	GRPCHealthSvc    string
	GRPCDoPing       bool
	GRPCPingDelay    time.Duration
}

// LoadTestStatus - used for representing load test status
type LoadTestStatus string

const (
	// LoadTestError - respresents an error status
	LoadTestError LoadTestStatus = "error"

	// LoadTestInfo - represents a info status
	LoadTestInfo LoadTestStatus = "info"

	// LoadTestSuccess - represents a success status
	LoadTestSuccess LoadTestStatus = "success"
)

// LoadTestResponse - used to bundle the response with status to the client
type LoadTestResponse struct {
	Status  LoadTestStatus `json:"status,omitempty"`
	Message string         `json:"message,omitempty"`
	Result  *MesheryResult `json:"result,omitempty"`
}

// MesheryResult - represents the results from Meshery test run to be shipped
type MesheryResult struct {
	ID     uuid.UUID              `json:"meshery_id,omitempty"`
	Name   string                 `json:"name,omitempty"`
	Mesh   string                 `json:"mesh,omitempty"`
	Result map[string]interface{} `json:"runner_results,omitempty"`

	ServerMetrics     interface{} `json:"server_metrics,omitempty"`
	ServerBoardConfig interface{} `json:"server_board_config,omitempty"`
}

// ConvertToSpec - converts meshery result to SMP
func (m *MesheryResult) ConvertToSpec() (*BenchmarkSpec, error) {
	b := &BenchmarkSpec{
		Env:     &Environment{},
		Client:  &MeshClientConfig{},
		Metrics: &Metrics{},
		ExpUUID: m.ID.String(),
	}
	var (
		results periodic.HasRunnerResult
	)
	retcodesString, _ := m.Result["RetCodes"].(map[string]interface{})
	logrus.Debugf("retcodes: %+v, %T", m.Result["RetCodes"], m.Result["RetCodes"])
	retcodes := map[int]int64{}
	for k, v := range retcodesString {
		k1, _ := strconv.Atoi(k)
		retcodes[k1], _ = v.(int64)
	}
	// retcodes[200] = 10
	m.Result["RetCodes"] = retcodes
	logrus.Debugf("result to be converted: %+v", m)
	if m.Result["RunType"].(string) == "HTTP" {
		httpResults := &fhttp.HTTPRunnerResults{}
		resJ, err := json.Marshal(m.Result)
		if err != nil {
			err = errors.Wrap(err, "unable while converting Meshery result to Benchmark Spec")
			logrus.Error(err)
			return nil, err
		}
		err = json.Unmarshal(resJ, httpResults)
		if err != nil {
			err = errors.Wrap(err, "unable while converting Meshery result to Benchmark Spec")
			logrus.Error(err)
			return nil, err
		}

		results = httpResults
		logrus.Debugf("httpresults: %+v", httpResults)
		b.EndpointURL = httpResults.URL
	} else {
		// TODO: GRPC
	}

	result := results.Result()
	b.StartTime = result.StartTime
	b.EndTime = result.StartTime.Add(result.ActualDuration)
	b.Client.Connections = result.NumThreads
	b.Client.Rps = result.ActualQPS
	b.Client.Internal = false
	b.Client.LatenciesMs = &LatenciesMs{
		Min:     result.DurationHistogram.Min,
		Max:     result.DurationHistogram.Max,
		Average: result.DurationHistogram.Avg,
	}
	for _, p := range result.DurationHistogram.Percentiles {
		switch p.Percentile {
		case 50:
			b.Client.LatenciesMs.P50 = p.Value
		case 90:
			b.Client.LatenciesMs.P90 = p.Value
		case 99:
			b.Client.LatenciesMs.P99 = p.Value
		}
	}

	k8sI, ok := m.Result["kubernetes"]
	if ok {
		k8s, _ := k8sI.(map[string]interface{})
		b.Env.Kubernetes, _ = k8s["server_version"].(string)
		nodesI, okk := k8s["nodes"]
		if okk {
			nodes, okkk := nodesI.([]*K8SNode)
			if okkk {
				b.Env.NodeCount = len(nodes)
			}
		}
	}
	return b, nil
}
