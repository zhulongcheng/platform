package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/influxdata/platform"
	"github.com/influxdata/platform/query"
	_ "github.com/influxdata/platform/query/builtin"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/execute/executetest"
	"github.com/influxdata/platform/query/plan"
)

func main() {
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	spec, err := query.Compile(context.Background(), string(data), time.Now())
	if err != nil {
		panic(err)
	}

	lplanner := plan.NewLogicalPlanner()
	lp, err := lplanner.Plan(spec)
	if err != nil {
		panic(err)
	}
	planner := plan.NewPlanner()
	p, err := planner.Plan(lp, nil)
	if err != nil {
		panic(err)
	}

	exec := execute.NewExecutor(nil)

	results, err := exec.Execute(context.Background(), platform.ID([]byte("foo")), p, executetest.UnlimitedAllocator)
	if err != nil {
		panic(err)
	}

	names := make([]string, 0, len(results))
	for name := range results {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		r := results[name]
		tables := r.Tables()
		fmt.Println("Result:", name)
		err := tables.Do(func(tbl query.Table) error {
			_, err := execute.NewFormatter(tbl, nil).WriteTo(os.Stdout)
			return err
		})
		if err != nil {
			panic(err)
		}
	}
}
