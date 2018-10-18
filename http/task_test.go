package http_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/platform"
	"github.com/influxdata/platform/http"
	"github.com/influxdata/platform/inmem"
	_ "github.com/influxdata/platform/query/builtin"
	"github.com/influxdata/platform/task"
	"github.com/influxdata/platform/task/backend"
	"github.com/influxdata/platform/task/servicetest"
	"go.uber.org/zap/zaptest"
)

func httpTaskServiceFactory(t *testing.T) (*servicetest.System, context.CancelFunc) {
	store := backend.NewInMemStore()
	rrw := backend.NewInMemRunReaderWriter()

	ctx, cancel := context.WithCancel(context.Background())

	backingTS := task.PlatformAdapter(store, rrw)

	i := inmem.NewService()

	h := http.NewTaskHandler(zaptest.NewLogger(t))
	h.TaskService = backingTS
	h.AuthorizationService = i

	auth := platform.Authorization{UserID: 1}
	if err := i.CreateAuthorization(ctx, &auth); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(h)
	go func() {
		<-ctx.Done()
		server.Close()
	}()

	tsFunc := func() platform.TaskService {
		return http.TaskService{
			Addr:  server.URL,
			Token: auth.Token,
		}
	}

	return &servicetest.System{
		S:               store,
		LR:              rrw,
		LW:              rrw,
		Ctx:             ctx,
		TaskServiceFunc: tsFunc,
	}, cancel
}

func TestTaskService(t *testing.T) {
	t.Parallel()

	servicetest.TestTaskService(t, httpTaskServiceFactory)
}
