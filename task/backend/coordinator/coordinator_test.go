package coordinator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	_ "github.com/influxdata/platform/query/builtin"
	"github.com/influxdata/platform/task/backend"
	"github.com/influxdata/platform/task/backend/coordinator"
	"github.com/influxdata/platform/task/mock"
	platformtesting "github.com/influxdata/platform/testing"
)

func timeoutSelector(ch <-chan *mock.Task) (*mock.Task, error) {
	select {
	case task := <-ch:
		return task, nil
	case <-time.After(time.Second):
		return nil, errors.New("timeout on select")
	}
}

const script = `option task = {name: "a task",cron: "* * * * *"} from(bucket:"test") |> range(start:-1h)`

func TestCoordinator(t *testing.T) {
	st := backend.NewInMemStore()
	sched := mock.NewScheduler()

	coord := coordinator.New(sched, st)
	createChan := sched.TaskCreateChan()
	releaseChan := sched.TaskReleaseChan()
	updateChan := sched.TaskUpdateChan()

	orgID := platformtesting.MustIDBase16("69746f7175650d0a")
	usrID := platformtesting.MustIDBase16("6c61757320657420")
	id, err := coord.CreateTask(context.Background(), backend.CreateTaskRequest{Org: orgID, User: usrID, Script: script})
	if err != nil {
		t.Fatal(err)
	}

	task, err := timeoutSelector(createChan)
	if err != nil {
		t.Fatal(err)
	}

	if task.Script != script {
		t.Fatal("task sent to scheduler doesnt match task created")
	}

	deleted, err := coord.DeleteTask(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}

	if !deleted {
		t.Fatal("no error and not deleted")
	}

	task, err = timeoutSelector(releaseChan)
	if err != nil {
		t.Fatal(err)
	}

	if task.Script != script {
		t.Fatal("task sent to scheduler doesnt match task created")
	}

	id, err = coord.CreateTask(context.Background(), backend.CreateTaskRequest{Org: orgID, User: usrID, Script: script})
	if err != nil {
		t.Fatal(err)
	}

	task, err = timeoutSelector(createChan)
	if err != nil {
		t.Fatal(err)
	}

	res, err := coord.UpdateTask(context.Background(), backend.UpdateTaskRequest{ID: id, Status: backend.TaskInactive})
	if err != nil {
		t.Fatal(err)
	}
	// Only validating res on the first update.
	if res.NewTask.ID != id {
		t.Fatalf("unexpected ID on update result: got %s, want %s", res.NewTask.ID.String(), id.String())
	}
	if res.NewTask.Script != script {
		t.Fatalf("unexpected script on update result: got %q, want %q", res.NewTask.Script, script)
	}
	if res.NewMeta.Status != string(backend.TaskInactive) {
		t.Fatalf("unexpected meta status on update result: got %q, want %q", res.NewMeta.Status, backend.TaskInactive)
	}
	if res.OldStatus != backend.TaskActive {
		t.Fatalf("unexpected old status on update result: got %q, want %q", res.OldStatus, backend.TaskActive)
	}

	task, err = timeoutSelector(releaseChan)
	if err != nil {
		t.Fatal(err)
	}

	if task.Script != script {
		t.Fatal("task sent to scheduler doesnt match task created")
	}

	if _, err := coord.UpdateTask(context.Background(), backend.UpdateTaskRequest{ID: id, Status: backend.TaskActive}); err != nil {
		t.Fatal(err)
	}

	task, err = timeoutSelector(createChan)
	if err != nil {
		t.Fatal(err)
	}

	if task.Script != script {
		t.Fatal("task sent to scheduler doesnt match task created")
	}

	newScript := `option task = {name: "a task",cron: "1 * * * *"} from(bucket:"test") |> range(start:-2h)`
	if _, err := coord.UpdateTask(context.Background(), backend.UpdateTaskRequest{ID: id, Script: newScript}); err != nil {
		t.Fatal(err)
	}

	task, err = timeoutSelector(updateChan)
	if err != nil {
		t.Fatal(err)
	}

	if task.Script != newScript {
		t.Fatal("task sent to scheduler doesnt match task created")
	}
}

func TestCoordinator_DeleteUnclaimedTask(t *testing.T) {
	st := backend.NewInMemStore()
	sched := mock.NewScheduler()

	coord := coordinator.New(sched, st)

	// Create an isolated task directly through the store so the coordinator doesn't know about it.
	id, err := st.CreateTask(context.Background(), backend.CreateTaskRequest{Org: 1, User: 2, Script: script})
	if err != nil {
		t.Fatal(err)
	}

	// Deleting the task through the coordinator should succeed.
	if _, err := coord.DeleteTask(context.Background(), id); err != nil {
		t.Fatal(err)
	}

	if _, err := st.FindTaskByID(context.Background(), id); err != backend.ErrTaskNotFound {
		t.Fatalf("expected deleted task not to be found; got %v", err)
	}
}
