package livelog

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hanoch-jfrog/forest/client/livelog/constants"
	"github.com/hanoch-jfrog/forest/client/livelog/model"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

const (
	mockHttpStrategyNodesEndpoint = "api/mock/nodes"
)

func Test_client_SetNodeId(t *testing.T) {
	s := &client{}
	s.SetNodeId("node-1")
	require.Equal(t, "node-1", s.nodeId)
}

func Test_client_SetLogFileName(t *testing.T) {
	s := &client{}
	s.SetLogFileName("one.log")
	require.Equal(t, "one.log", s.logFileName)
}

func Test_client_SetLogsRefreshRate(t *testing.T) {
	s := &client{}
	s.SetLogsRefreshRate(time.Minute)
	require.Equal(t, time.Minute, s.logsRefreshRate)
}

func Test_client_GetServiceNodeIds(t *testing.T) {
	tests := []struct {
		name            string
		mockGetResponse []byte
		mockGetErr      error
		want            []string
		wantErr         bool
	}{
		{
			name:       "error response",
			mockGetErr: fmt.Errorf("some-error"),
			wantErr:    true,
		},
		{
			name:            "no node ids response",
			mockGetResponse: []byte("{\"nodes\": []}"),
			wantErr:         true,
		},
		{
			name:            "nodes response",
			mockGetResponse: []byte("{\"nodes\": [{\"node_id\": \"node-1\"}, {\"node_id\": \"node-2\"}]}"),
			want:            []string{"node-1", "node-2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &client{
				httpStrategy: &mockHttpStrategy{
					t:              t,
					expectEndpoint: mockHttpStrategyNodesEndpoint,
					getResponse:    tt.mockGetResponse,
					getErr:         tt.mockGetErr,
				},
			}
			got, err := s.GetServiceNodeIds(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServiceNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServiceNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_GetConfig(t *testing.T) {
	tests := []struct {
		name            string
		nodeId          string
		mockGetResponse []byte
		mockGetErr      error
		want            *model.Config
		wantErr         bool
	}{
		{
			name:    "missing node id",
			wantErr: true,
		},
		{
			name:       "error response",
			nodeId:     "node-1",
			mockGetErr: fmt.Errorf("some-error"),
			wantErr:    true,
		},
		{
			name:            "no log files found response",
			nodeId:          "node-1",
			mockGetResponse: []byte("{\"logs\": [], \"refresh_rate_millis\": 123}"),
			wantErr:         true,
		},
		{
			name:            "config response",
			nodeId:          "node-1",
			mockGetResponse: []byte("{\"logs\": [\"one.log\", \"two.log\"], \"refresh_rate_millis\": 123}"),
			want: &model.Config{
				LogFileNames:      []string{"one.log", "two.log"},
				RefreshRateMillis: int64(123),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &client{
				httpStrategy: &mockHttpStrategy{
					t:              t,
					expectEndpoint: constants.ConfigEndpoint,
					expectNodeId:   tt.nodeId,
					getResponse:    tt.mockGetResponse,
					getErr:         tt.mockGetErr,
				},
				nodeId: tt.nodeId,
			}
			got, err := s.GetConfig(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_CatLog(t *testing.T) {
	tests := []struct {
		name            string
		nodeId          string
		logFileName     string
		mockGetResponse []byte
		mockGetErr      error
		want            string
		wantErr         bool
	}{
		{
			name:    "missing node id",
			wantErr: true,
		},
		{
			name:    "missing log file name",
			nodeId:  "node-1",
			wantErr: true,
		},
		{
			name:        "error response",
			nodeId:      "node-1",
			logFileName: "one.log",
			mockGetErr:  fmt.Errorf("some-error"),
			wantErr:     true,
		},
		{
			name:            "cat log response",
			nodeId:          "node-1",
			logFileName:     "one.log",
			mockGetResponse: []byte("{\"log_content\": \"some log content\", \"file_size\": 123}"),
			want:            "some log content",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &client{
				httpStrategy: &mockHttpStrategy{
					t:              t,
					expectEndpoint: fmt.Sprintf("%s?$file_size=0&id=%s", constants.DataEndpoint, tt.logFileName),
					expectNodeId:   tt.nodeId,
					getResponse:    tt.mockGetResponse,
					getErr:         tt.mockGetErr,
				},
				nodeId:      tt.nodeId,
				logFileName: tt.logFileName,
			}
			out := &bytes.Buffer{}
			err := s.CatLog(context.Background(), out)
			if (err != nil) != tt.wantErr {
				t.Errorf("CatLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(out.String(), tt.want) {
				t.Errorf("CatLog() got = %v, want %v", out.String(), tt.want)
			}
		})
	}
}

func Test_client_TailLog_errors(t *testing.T) {
	tests := []struct {
		name        string
		nodeId      string
		logFileName string
		mockGetErr  error
		want        string
		wantErr     bool
	}{
		{
			name:    "missing node id",
			wantErr: true,
		},
		{
			name:    "missing log file name",
			nodeId:  "node-1",
			wantErr: true,
		},
		{
			name:        "error response",
			nodeId:      "node-1",
			logFileName: "one.log",
			mockGetErr:  fmt.Errorf("some-error"),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &client{
				httpStrategy: &mockHttpStrategy{
					t:              t,
					expectEndpoint: fmt.Sprintf("%s?$file_size=0&id=%s", constants.DataEndpoint, tt.logFileName),
					expectNodeId:   tt.nodeId,
					getErr:         tt.mockGetErr,
				},
				nodeId:      tt.nodeId,
				logFileName: tt.logFileName,
			}

			out := &bytes.Buffer{}
			err := s.TailLog(context.Background(), out)
			if (err != nil) != tt.wantErr {
				t.Errorf("TailLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(out.String(), tt.want) {
				t.Errorf("TailLog() got = %v, want %v", out.String(), tt.want)
			}
		})
	}
}

func Test_client_TailLog(t *testing.T) {
	s := &client{
		httpStrategy: &mockTailLogHttpStrategy{
			t:                 t,
			expectNodeId:      "node-1",
			expectLogFileName: "one.log",
			getResponses: [][]byte{
				[]byte("{\"log_content\": \"one \", \"file_size\": 1}"),
				[]byte("{\"log_content\": \"two \", \"file_size\": 2}"),
				[]byte("{\"log_content\": \"three \", \"file_size\": 3}"),
				[]byte("{\"log_content\": \"four \", \"file_size\": 4}"),
				[]byte("{\"log_content\": \"five \", \"file_size\": 5}"),
			},
		},
		nodeId:          "node-1",
		logFileName:     "one.log",
		logsRefreshRate: 100 * time.Millisecond,
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	out := &bytes.Buffer{}
	err := s.TailLog(timeoutCtx, out)
	require.NoError(t, err)
	// expected to get 3 first log chunks before context times out
	if !reflect.DeepEqual(out.String(), "one two three ") {
		t.Errorf("TailLog() got = %v, want %v", out.String(), "one two three ")
	}
}

type mockHttpStrategy struct {
	t              *testing.T
	expectEndpoint string
	expectNodeId   string
	getResponse    []byte
	getErr         error
}

func (s *mockHttpStrategy) NodesEndpoint() string {
	return mockHttpStrategyNodesEndpoint
}

func (s *mockHttpStrategy) SendGet(_ context.Context, endpoint, nodeId string) ([]byte, error) {
	require.Equal(s.t, s.expectEndpoint, endpoint)
	require.Equal(s.t, s.expectNodeId, nodeId)
	return s.getResponse, s.getErr
}

type mockTailLogHttpStrategy struct {
	t                 *testing.T
	expectNodeId      string
	expectLogFileName string
	getResponses      [][]byte
	getCount          int
}

func (s *mockTailLogHttpStrategy) NodesEndpoint() string {
	return mockHttpStrategyNodesEndpoint
}

func (s *mockTailLogHttpStrategy) SendGet(_ context.Context, endpoint, nodeId string) ([]byte, error) {
	expectEndpoint := fmt.Sprintf("%s?$file_size=%d&id=%s", constants.DataEndpoint, s.getCount, s.expectLogFileName)
	require.Equal(s.t, expectEndpoint, endpoint)
	require.Equal(s.t, s.expectNodeId, nodeId)
	res := s.getResponses[s.getCount]
	s.getCount++
	return res, nil
}
