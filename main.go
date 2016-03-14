package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rschmukler/drone-slack-enhanced/slack"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin"
)

type Slack struct {
	Webhook   string `json:"webhook_url"`
	Channel   string `json:"channel"`
	Recipient string `json:"recipient"`
	Username  string `json:"username"`
	VCS       string `json:"vcs"`
}

func main() {
	var (
		repo  = new(drone.Repo)
		build = new(drone.Build)
		sys   = new(drone.System)
		vargs = new(Slack)
	)

	plugin.Param("build", build)
	plugin.Param("repo", repo)
	plugin.Param("system", sys)
	plugin.Param("vargs", vargs)

	err := plugin.Parse()
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

	attach := msg.NewAttachment()
	attach.Title = fmt.Sprintf("Build #%d %s in %s", build.Number, status(build), time.Duration((build.Finished-build.Started)*int64(time.Second)).String())
	attach.TitleLink = fmt.Sprintf("%s/%s/%d", sys.Link, repo.FullName, build.Number)
	attach.Fallback = fallback(repo, build)
	attach.Color = color(build)
	attach.MrkdwnIn = []string{"text", "fallback"}
	attach.Fields = []*slack.Field{
		{Title: "Commit", Value: fmt.Sprintf("<%s|%s>", commitURL(build, repo, vargs), build.Message)},
		{Title: "Repo", Value: fmt.Sprintf("<%s|%s>", repoURL(repo, vargs), repo.FullName), Short: true},
		{Title: "Branch", Value: fmt.Sprintf("<%s|%s>", branchURL(build, repo, vargs), build.Branch), Short: true},
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

func commitURL(build *drone.Build, repo *drone.Repo, args *Slack) string {
	return fmt.Sprintf("%s/commit/%s", repoURL(repo, args), build.Commit)
}

func branchURL(build *drone.Build, repo *drone.Repo, args *Slack) string {
	return fmt.Sprintf("%s/src/%s", repoURL(repo, args), build.Branch)
}

func repoURL(repo *drone.Repo, args *Slack) string {
	return fmt.Sprintf("https://%s/%s", args.VCS, repo.FullName)
}

func fallback(repo *drone.Repo, build *drone.Build) string {
	return fmt.Sprintf("%s %s/%s#%s (%s) by %s",
		build.Status,
		repo.Owner,
		repo.Name,
		build.Commit[:8],
		build.Branch,
		build.Author,
	)
}

func color(build *drone.Build) string {
	switch build.Status {
	case drone.StatusSuccess:
		return "good"
	case drone.StatusFailure, drone.StatusError, drone.StatusKilled:
		return "danger"
	default:
		return "warning"
	}
}

func status(build *drone.Build) string {
	switch build.Status {
	case drone.StatusSuccess:
		return "Passed"
	case drone.StatusFailure:
		return "Failed"
	case drone.StatusKilled:
		return "Aborted"
	case drone.StatusError:
		return "Errored"
	default:
		return "Failed"
	}
}
