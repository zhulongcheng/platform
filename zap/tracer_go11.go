// +build go1.11

package zap

import (
	"context"
	"runtime/trace"
)

type task trace.Task

func newTask(pctx context.Context, name string) (context.Context, *task) {
	ctx, stask := trace.NewTask(pctx, name)
	return ctx, (*task)(stask)
}

func (t *task) End() { (*trace.Task)(t).End() }
