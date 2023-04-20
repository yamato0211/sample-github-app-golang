package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v29/github"
)

func main() {
	http.HandleFunc("/github/events", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		secret := os.Getenv("GITHUB_APP_SECRET")
		payload, err := github.ValidatePayload(r, []byte(secret))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		webhookEvent, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch event := webhookEvent.(type) {
		case *github.IssuesEvent:
			if err := processIssuesEvent(ctx, event); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	})

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func processIssuesEvent(ctx context.Context, event *github.IssuesEvent) error {
	if event.GetAction() != "opened" {
		return nil
	}

	installationID := event.GetInstallation().GetID()
	client, err := newGithubClient(installationID)
	if err != nil {
		return err
	}

	repoOwner := event.Repo.GetOwner().GetLogin()
	repo := event.Repo.GetName()

	issue := event.GetIssue()
	issueNumber := issue.GetNumber()
	user := issue.GetUser().GetLogin()

	body := "hello, @" + user
	comment := &github.IssueComment{
		Body: &body,
	}
	if _, _, err := client.Issues.CreateComment(ctx, repoOwner, repo, issueNumber, comment); err != nil {
		return err
	}

	return nil
}

func newGithubClient(installationID int64) (*github.Client, error) {
	appID, err := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)
	if err != nil {
		return nil, err
	}

	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, appID, installationID, "private-key.pem")
	if err != nil {
		return nil, err
	}

	return github.NewClient(&http.Client{
		Transport: itr,
		Timeout:   5 * time.Second,
	}), nil
}
