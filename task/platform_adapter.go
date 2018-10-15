package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/platform"
	"github.com/influxdata/platform/task/backend"
	"github.com/influxdata/platform/task/options"
)

type StoreRunController interface {
	backend.Store
	RunController
}
type RunController interface {
	CancelRun(ctx context.Context, taskID, runID platform.ID) error
	//TODO: add retry run to this.
}

// PlatformAdapter wraps a task.Store into the platform.TaskService interface.
func PlatformAdapter(s StoreRunController, r backend.LogReader) platform.TaskService {
	return pAdapter{s: s, r: r}
}

type pAdapter struct {
	s StoreRunController
	r backend.LogReader
}

var _ platform.TaskService = pAdapter{}

func (p pAdapter) FindTaskByID(ctx context.Context, id platform.ID) (*platform.Task, error) {
	t, m, err := p.s.FindTaskByIDWithMeta(ctx, id)
	if err != nil {
		return nil, err
	}

	// The store interface specifies that a returned task is nil if the operation succeeded without a match.
	if t == nil {
		return nil, nil
	}

	return toPlatformTask(*t, m)
}

func (p pAdapter) FindTasks(ctx context.Context, filter platform.TaskFilter) ([]*platform.Task, int, error) {
	const pageSize = 100 // According to the platform.TaskService.FindTasks API.

	params := backend.TaskSearchParams{PageSize: pageSize}
	if filter.Organization != nil {
		params.Org = *filter.Organization
	}
	if filter.User != nil {
		params.User = *filter.User
	}
	if filter.After != nil {
		params.After = *filter.After
	}
	ts, err := p.s.ListTasks(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	pts := make([]*platform.Task, len(ts))
	for i, t := range ts {
		pts[i], err = toPlatformTask(t, nil)
		if err != nil {
			return nil, 0, err
		}
	}

	totalResults := len(pts) // TODO(mr): don't lie about the total results. Update ListTasks signature?
	return pts, totalResults, nil
}

func (p pAdapter) CreateTask(ctx context.Context, t *platform.Task) error {
	opts, err := options.FromScript(t.Flux)
	if err != nil {
		return err
	}

	// TODO(mr): decide whether we allow user to configure scheduleAfter. https://github.com/influxdata/platform/issues/595
	scheduleAfter := time.Now().Unix()
	req := backend.CreateTaskRequest{Org: t.Organization, User: t.Owner.ID, Script: t.Flux, ScheduleAfter: scheduleAfter}
	id, err := p.s.CreateTask(ctx, req)
	if err != nil {
		return err
	}
	t.ID = id
	t.Every = opts.Every.String()
	t.Cron = opts.Cron

	return nil
}

func (p pAdapter) UpdateTask(ctx context.Context, id platform.ID, upd platform.TaskUpdate) (*platform.Task, error) {
	if upd.Flux == nil && upd.Status == nil {
		return nil, errors.New("cannot update task without content")
	}

	task := &platform.Task{
		ID:     id,
		Name:   "TODO",
		Status: "TODO",
		Owner:  platform.User{}, // TODO(mr): populate from context?
	}
	if upd.Flux != nil {
		task.Flux = *upd.Flux

		opts, err := options.FromScript(task.Flux)
		if err != nil {
			return nil, err
		}
		task.Every = opts.Every.String()
		task.Cron = opts.Cron

		if err := p.s.ModifyTask(ctx, id, task.Flux); err != nil {
			return nil, err
		}
	}

	if upd.Status != nil {
		var err error
		switch *upd.Status {
		case string(backend.TaskActive):
			err = p.s.EnableTask(ctx, id)
		case string(backend.TaskInactive):
			err = p.s.DisableTask(ctx, id)
		default:
			err = fmt.Errorf("invalid status: %s", *upd.Status)
		}

		if err != nil {
			return nil, err
		}
		task.Status = *upd.Status
	}

	return task, nil
}

func (p pAdapter) DeleteTask(ctx context.Context, id platform.ID) error {
	_, err := p.s.DeleteTask(ctx, id)
	// TODO(mr): Store.DeleteTask returns false, nil if ID didn't match; do we want to handle that case?
	return err
}

func (p pAdapter) FindLogs(ctx context.Context, filter platform.LogFilter) ([]*platform.Log, int, error) {
	logs, err := p.r.ListLogs(ctx, filter)
	logPointers := make([]*platform.Log, len(logs))
	for i := range logs {
		logPointers[i] = &logs[i]
	}
	return logPointers, len(logs), err
}

func (p pAdapter) FindRuns(ctx context.Context, filter platform.RunFilter) ([]*platform.Run, int, error) {
	runs, err := p.r.ListRuns(ctx, filter)
	return runs, len(runs), err
}

func (p pAdapter) FindRunByID(ctx context.Context, orgID, id platform.ID) (*platform.Run, error) {
	return p.r.FindRunByID(ctx, orgID, id)
}

func (p pAdapter) RetryRun(ctx context.Context, id platform.ID) (*platform.Run, error) {
	return nil, errors.New("not yet implemented")
}

func (p pAdapter) CancelRun(ctx context.Context, taskID, runID platform.ID) error {
	return p.s.CancelRun(ctx, taskID, runID)
}

func toPlatformTask(t backend.StoreTask, m *backend.StoreTaskMeta) (*platform.Task, error) {
	opts, err := options.FromScript(t.Script)
	if err != nil {
		return nil, err
	}

	pt := &platform.Task{
		ID:           t.ID,
		Organization: t.Org,
		Name:         t.Name,
		Owner: platform.User{
			ID:   t.User,
			Name: "", // TODO(mr): how to get owner name?
		},
		Flux: t.Script,
		Cron: opts.Cron,
	}
	if opts.Every != 0 {
		pt.Every = opts.Every.String()
	}
	if m != nil {
		pt.Status = string(m.Status)
	}
	return pt, nil
}
