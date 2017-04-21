package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/rschmukler/drone-slack-enhanced/slack"
)

type Args struct {
	Webhook   string `envconfig:"webhook_url"`
	Channel   string
	Recipient string
	Username  string
}

type DroneVars struct {
	BuildNumber   int    `envconfig:"build_number"`
	BuildFinished string `envconfig:"build_finished"`
	BuildStatus   string `envconfig:"build_status"`
	BuildLink     string `envconfig:"build_link"`
	CommitSha     string `envconfig:"commit_sha"`
	CommitBranch  string `envconfig:"commit_branch"`
	CommitAuthor  string `envconfig:"commit_author"`
	CommitLink    string `envconfig:"commit_link"`
	CommitMessage string `envconfig:"commit_message"`
	JobStarted    int64  `envconfig:"job_started"`
	Repo          string `envconfig:"build_link"`
	RepoLink      string `envconfig:"repo_link"`
	System        string
}

func main() {
	var (
		err   error
		vargs Args
		drone DroneVars
	)

	err = envconfig.Process("plugin", &vargs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = envconfig.Process("drone", &drone)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create the Slack client
	client := slack.NewClient(vargs.Webhook)

	// generate the Slack message
	msg := slack.Message{
		Username: vargs.Username,
		Channel:  vargs.Recipient,
	}

	if msg.Username == "" {
		msg.Username = "Drone CI"
	}

	// prepend the @ or # symbol if the user forgot to include
	// in their configuration string.
	if len(vargs.Recipient) != 0 {
		msg.Channel = prepend("@", vargs.Recipient)
	} else {
		if vargs.Channel == "" {
			vargs.Channel = "dev"
		}
		msg.Channel = prepend("#", vargs.Channel)
	}

	runTime := time.Duration((time.Now().Unix() - drone.JobStarted) * int64(time.Second))

	attach := msg.NewAttachment()
	attach.Title = fmt.Sprintf("Build #%d %s in %s", drone.BuildNumber, fmtStatus(drone.BuildStatus), runTime.String())
	attach.TitleLink = fmt.Sprintf(drone.BuildLink)
	attach.Fallback = fallback(&drone)
	attach.Color = color(drone.BuildStatus)
	attach.MrkdwnIn = []string{"text", "fallback", "fields"}
	attach.Fields = []*slack.Field{
		{Title: "Commit", Value: fmt.Sprintf("<%s|%s>", drone.CommitLink, strings.TrimSpace(drone.CommitMessage))},
		{Title: "Repo", Value: fmt.Sprintf("<%s|%s>", drone.RepoLink, drone.Repo), Short: true},
		{Title: "Branch", Value: fmt.Sprintf("<%s|%s>", branchURL(&drone), drone.CommitBranch), Short: true},
	}

	// sends the message
	if err := client.SendMessage(&msg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func prepend(prefix, s string) string {
	if !strings.HasPrefix(s, prefix) {
		return prefix + s
	}
	return s
}

func branchURL(drone *DroneVars) string {
	return fmt.Sprintf("%s/src/%s", drone.RepoLink, drone.CommitBranch)
}

func fallback(drone *DroneVars) string {
	return fmt.Sprintf("%s %s#%s (%s) by %s",
		fmtStatus(drone.BuildStatus),
		drone.Repo,
		drone.CommitSha[:8],
		drone.CommitBranch,
		drone.CommitAuthor,
	)
}

func color(buildStatus string) string {
	switch buildStatus {
	case "success":
		return "good"
	case "failure", "error", "killed":
		return "danger"
	default:
		return "warning"
	}
}

func fmtStatus(status string) string {
	return strings.Title(status)
}
