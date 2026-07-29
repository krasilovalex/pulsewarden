package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/wayzzoo/pulsewarden/internal/domain/checkresult"
	repositorypostgres "github.com/wayzzoo/pulsewarden/internal/repository/postgres"
)

type monitorResultsListerFunc func(
	context.Context,
	uuid.UUID,
	int,
) ([]checkresult.CheckResult, error)

func (f monitorResultsListerFunc) Execute(
	ctx context.Context,
	monitorID uuid.UUID,
	limit int,
) ([]checkresult.CheckResult, error) {
	return f(ctx, monitorID, limit)
}

func TestListMonitorResultsHandler(t *testing.T) {
	monitorID := uuid.New()
	resultID := uuid.New()
	statusCode := 200

	checkedAt := time.Date(
		2026,
		time.July,
		30,
		1,
		30,
		0,
		0,
		time.UTC,
	)

	server := NewServer(ServerConfig{
		Logger: testLogger(),
		MonitorResultsLister: monitorResultsListerFunc(func(
			_ context.Context,
			gotMonitorID uuid.UUID,
			limit int,
		) ([]checkresult.CheckResult, error) {
			if gotMonitorID != monitorID {
				t.Fatalf(
					"monitor ID = %s, want %s",
					gotMonitorID,
					monitorID,
				)
			}

			if limit != 25 {
				t.Fatalf(
					"limit = %d, want 25",
					limit,
				)
			}

			return []checkresult.CheckResult{
				{
					ID:         resultID,
					MonitorID:  monitorID,
					Status:     checkresult.StatusUp,
					StatusCode: &statusCode,
					Latency:    84 * time.Millisecond,
					CheckedAt:  checkedAt,
				},
			}, nil
		}),
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/monitors/"+
			monitorID.String()+
			"/results?limit=25",
		nil,
	)

	responseRecorder := httptest.NewRecorder()

	server.Handler.ServeHTTP(
		responseRecorder,
		request,
	)

	result := responseRecorder.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			result.StatusCode,
			http.StatusOK,
		)
	}

	var payload listMonitorResultsResponse

	if err := json.NewDecoder(
		result.Body,
	).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload.Items) != 1 {
		t.Fatalf(
			"items length = %d, want 1",
			len(payload.Items),
		)
	}

	item := payload.Items[0]

	if item.ID != resultID.String() {
		t.Fatalf(
			"result ID = %q, want %q",
			item.ID,
			resultID.String(),
		)
	}

	if item.Status != "up" {
		t.Fatalf(
			"status = %q, want up",
			item.Status,
		)
	}

	if item.LatencyMilliseconds != 84 {
		t.Fatalf(
			"latency = %d, want 84",
			item.LatencyMilliseconds,
		)
	}

	if item.StatusCode == nil ||
		*item.StatusCode != 200 {
		t.Fatalf(
			"status code = %v, want 200",
			item.StatusCode,
		)
	}

	if item.CheckedAt != checkedAt.Format(
		time.RFC3339Nano,
	) {
		t.Fatalf(
			"checked at = %q, want %q",
			item.CheckedAt,
			checkedAt.Format(time.RFC3339Nano),
		)
	}
}

func TestListMonitorResultsHandlerInvalidLimit(
	t *testing.T,
) {
	server := NewServer(ServerConfig{
		Logger: testLogger(),
		MonitorResultsLister: monitorResultsListerFunc(func(
			context.Context,
			uuid.UUID,
			int,
		) ([]checkresult.CheckResult, error) {
			t.Fatal("lister must not be called")
			return nil, nil
		}),
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/monitors/"+
			uuid.NewString()+
			"/results?limit=0",
		nil,
	)

	responseRecorder := httptest.NewRecorder()

	server.Handler.ServeHTTP(
		responseRecorder,
		request,
	)

	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status code = %d, want %d",
			responseRecorder.Code,
			http.StatusBadRequest,
		)
	}
}

func TestListMonitorResultsHandlerMonitorNotFound(
	t *testing.T,
) {
	server := NewServer(ServerConfig{
		Logger: testLogger(),
		MonitorResultsLister: monitorResultsListerFunc(func(
			context.Context,
			uuid.UUID,
			int,
		) ([]checkresult.CheckResult, error) {
			return nil,
				repositorypostgres.ErrMonitorNotFound
		}),
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/monitors/"+
			uuid.NewString()+
			"/results",
		nil,
	)

	responseRecorder := httptest.NewRecorder()

	server.Handler.ServeHTTP(
		responseRecorder,
		request,
	)

	if responseRecorder.Code != http.StatusNotFound {
		t.Fatalf(
			"status code = %d, want %d",
			responseRecorder.Code,
			http.StatusNotFound,
		)
	}
}
