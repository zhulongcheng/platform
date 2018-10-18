package main

import (
	"context"
	"fmt"
	"os"

	"github.com/influxdata/flux/repl"
	"github.com/influxdata/platform"
	"github.com/influxdata/platform/cmd/influx/internal"
	"github.com/influxdata/platform/http"
	"github.com/spf13/cobra"
)

// task Command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "task related commands",
	Run:   taskF,
}

func taskF(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "log related commands",
	Run:   logF,
}

func logF(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run related commands",
	Run:   runF,
}

func runF(cmd *cobra.Command, args []string) {
	cmd.Usage()
}

func init() {
	taskCmd.AddCommand(runCmd)
	taskCmd.AddCommand(logCmd)
}

// TaskCreateFlags define the Create Command
type TaskCreateFlags struct {
	org   string
	orgID string
	flux  string
}

var taskCreateFlags TaskCreateFlags

func init() {
	taskCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create task",
		Run:   taskCreateF,
	}

	taskCreateCmd.Flags().StringVarP(&taskCreateFlags.orgID, "org-id", "", "", "id of the organization that owns the task")
	taskCreateCmd.Flags().StringVarP(&taskCreateFlags.flux, "flux", "", "", "flux script. Can be pulled from a file using @/path/to/file.")
	taskCreateCmd.MarkFlagRequired("flux")

	taskCmd.AddCommand(taskCreateCmd)
}

func taskCreateF(cmd *cobra.Command, args []string) {
	if taskCreateFlags.org != "" && taskCreateFlags.orgID != "" {
		fmt.Println("must specify exactly one of org or org-id")
		cmd.Usage()
		os.Exit(1)
	}

	s := &http.TaskService{
		Addr:  flags.host,
		Token: flags.token,
	}

	flux, err := repl.LoadQuery(taskCreateFlags.flux)
	if err != nil {
		fmt.Printf("error parsing flux script: %s\n", err)
		os.Exit(1)
	}

	t := &platform.Task{
		Flux: flux,
	}

	if taskCreateFlags.orgID != "" {
		id, err := platform.IDFromString(taskCreateFlags.orgID)
		if err != nil {
			fmt.Printf("error parsing organization id: %v\n", err)
			os.Exit(1)
		}
		t.Organization = *id
	}

	if err := s.CreateTask(context.Background(), t); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	w := internal.NewTabWriter(os.Stdout)
	w.WriteHeaders(
		"ID",
		"Name",
		"Organization",
		"Status",
		"Every",
		"Cron",
	)
	w.Write(map[string]interface{}{
		"ID":           t.ID.String(),
		"Name":         t.Name,
		"Organization": t.Organization.String,
		"Status":       t.Status,
		"Every":        t.Every,
		"Cron":         t.Cron,
	})
	w.Flush()
}

// taskFindFlags define the Find Command
type TaskFindFlags struct {
	user  string
	id    string
	orgID string
}

var taskFindFlags TaskFindFlags

func init() {
	taskFindCmd := &cobra.Command{
		Use:   "find",
		Short: "Find tasks",
		Run:   taskFindF,
	}

	taskFindCmd.Flags().StringVarP(&taskFindFlags.id, "id", "i", "", "task ID")
	taskFindCmd.Flags().StringVarP(&taskFindFlags.user, "user-id", "n", "", "task owner ID")
	taskFindCmd.Flags().StringVarP(&taskFindFlags.orgID, "org-id", "", "", "task organization ID")

	taskCmd.AddCommand(taskFindCmd)
}

func taskFindF(cmd *cobra.Command, args []string) {
	s := &http.TaskService{
		Addr:  flags.host,
		Token: flags.token,
	}

	filter := platform.TaskFilter{}
	if taskFindFlags.user != "" {
		id, err := platform.IDFromString(taskFindFlags.user)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		filter.User = id
	}

	if taskFindFlags.orgID != "" {
		id, err := platform.IDFromString(taskFindFlags.orgID)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		filter.Organization = id
	}

	var tasks []*platform.Task
	var err error

	if taskFindFlags.id != "" {
		id, err := platform.IDFromString(taskFindFlags.id)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		task, err := s.FindTaskByID(context.Background(), *id)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		tasks = append(tasks, task)
	} else {
		tasks, _, err = s.FindTasks(context.Background(), filter)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	w := internal.NewTabWriter(os.Stdout)
	w.WriteHeaders(
		"ID",
		"Name",
		"Organization",
		"Status",
		"Every",
		"Cron",
	)
	for _, t := range tasks {
		w.Write(map[string]interface{}{
			"ID":           t.ID.String(),
			"Name":         t.Name,
			"Organization": t.Organization.String,
			"Status":       t.Status,
			"Every":        t.Every,
			"Cron":         t.Cron,
		})
	}
	w.Flush()
}

// taskUpdateFlags define the Update Command
type TaskUpdateFlags struct {
	id     string
	status string
	flux   string
}

var taskUpdateFlags TaskUpdateFlags

func init() {
	taskUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update task",
		Run:   taskUpdateF,
	}

	taskUpdateCmd.Flags().StringVarP(&taskUpdateFlags.id, "id", "i", "", "task ID (required)")
	taskUpdateCmd.Flags().StringVarP(&taskUpdateFlags.status, "status", "", "", "update task status")
	taskUpdateCmd.Flags().StringVarP(&taskUpdateFlags.flux, "flux", "", "", "new flux script")
	taskUpdateCmd.MarkFlagRequired("id")

	taskCmd.AddCommand(taskUpdateCmd)
}

func taskUpdateF(cmd *cobra.Command, args []string) {
	s := &http.TaskService{
		Addr:  flags.host,
		Token: flags.token,
	}

	var id platform.ID
	if err := id.DecodeFromString(taskUpdateFlags.id); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	update := platform.TaskUpdate{}
	if taskUpdateFlags.status != "" {
		update.Status = &taskUpdateFlags.status
	}

	if taskUpdateFlags.flux != "" {
		flux, err := repl.LoadQuery(taskCreateFlags.flux)
		if err != nil {
			fmt.Printf("error parsing flux script: %s\n", err)
			os.Exit(1)
		}

		update.Flux = &flux
	}

	t, err := s.UpdateTask(context.Background(), id, update)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	w := internal.NewTabWriter(os.Stdout)
	w.WriteHeaders(
		"ID",
		"Name",
		"Organization",
		"Status",
		"Every",
		"Cron",
	)
	w.Write(map[string]interface{}{
		"ID":           t.ID.String(),
		"Name":         t.Name,
		"Organization": t.Organization.String,
		"Status":       t.Status,
		"Every":        t.Every,
		"Cron":         t.Cron,
	})
	w.Flush()
}

// taskDeleteFlags define the Delete command
type TaskDeleteFlags struct {
	id string
}

var taskDeleteFlags TaskDeleteFlags

func init() {
	taskDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete task",
		Run:   taskDeleteF,
	}

	taskDeleteCmd.Flags().StringVarP(&taskDeleteFlags.id, "id", "i", "", "task id (required)")
	taskDeleteCmd.MarkFlagRequired("id")

	taskCmd.AddCommand(taskDeleteCmd)
}

func taskDeleteF(cmd *cobra.Command, args []string) {
	s := &http.TaskService{
		Addr:  flags.host,
		Token: flags.token,
	}

	var id platform.ID
	err := id.DecodeFromString(taskDeleteFlags.id)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx := context.TODO()
	t, err := s.FindTaskByID(ctx, id)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = s.DeleteTask(ctx, id); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	w := internal.NewTabWriter(os.Stdout)
	w.WriteHeaders(
		"ID",
		"Name",
		"Organization",
		"Status",
		"Every",
		"Cron",
	)
	w.Write(map[string]interface{}{
		"ID":           t.ID.String(),
		"Name":         t.Name,
		"Organization": t.Organization.String,
		"Status":       t.Status,
		"Every":        t.Every,
		"Cron":         t.Cron,
	})
	w.Flush()
}

// taskLogFindFlags define the Delete command
type TaskLogFindFlags struct {
	taskID string
	runID  string
	orgID  string
}

var taskLogFindFlags TaskLogFindFlags

func init() {
	taskLogFindCmd := &cobra.Command{
		Use:   "",
		Short: "Delete task",
		Run:   taskLogFindF,
	}

	taskLogFindCmd.Flags().StringVarP(&taskLogFindFlags.taskID, "task-id", "", "", "task id (required)")
	taskLogFindCmd.Flags().StringVarP(&taskLogFindFlags.runID, "run-id", "", "", "run id")
	taskLogFindCmd.Flags().StringVarP(&taskLogFindFlags.orgID, "org-id", "", "", "organization id")
	taskLogFindCmd.MarkFlagRequired("task-id")

	taskCmd.AddCommand(taskLogFindCmd)
}

func taskLogFindF(cmd *cobra.Command, args []string) {
	s := &http.TaskService{
		Addr:  flags.host,
		Token: flags.token,
	}

	var filter platform.LogFilter
	id, err := platform.IDFromString(taskLogFindFlags.taskID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	filter.Task = id

	if taskLogFindFlags.runID != "" {
		id, err := platform.IDFromString(taskLogFindFlags.runID)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		filter.Run = id
	}

	if taskLogFindFlags.orgID != "" {
		id, err := platform.IDFromString(taskLogFindFlags.orgID)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		filter.Org = id
	}

	ctx := context.TODO()
	logs, _, err := s.FindLogs(ctx, filter)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	w := internal.NewTabWriter(os.Stdout)
	w.WriteHeaders(
		"Log",
	)
	for _, log := range logs {
		w.Write(map[string]interface{}{
			"Log": log,
		})
	}
	w.Flush()
}
