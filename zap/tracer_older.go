// +build !go1.11

package zap

import "context"

type task struct{}

func newTask(pctx context.Context, name string) (context.Context, *task) {
	return pctx, new(task)
}

func (t *task) End() {}
