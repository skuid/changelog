package github

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/skuid/changelog/src/changelog"
	"github.com/skuid/changelog/webhooks"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const StatusFailure = "failure"
const StatusPending = "pending"
const StatusSuccess = "success"
const StatusError = "error"

type githubWebhookHelper struct {
	*github.Client
}

func newGithubWebhookHelper(apiToken string) *githubWebhookHelper {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apiToken},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	return &githubWebhookHelper{client}
}

func (client *githubWebhookHelper) getPrCommits(event *github.PullRequestEvent, apiToken string) (changelog.Commits, error) {
	// list the commits on the pull request
	prCommits, _, err := client.PullRequests.ListCommits(
		context.Background(),
		event.Repo.Owner.GetLogin(),
		event.Repo.GetName(),
		event.PullRequest.GetNumber(),
		&github.ListOptions{},
	)
	if err != nil {
		return nil, err
	}
	// format them properly
	commits := changelog.Commits{}
	for _, c := range prCommits {
		commit := changelog.NewCommit(c.GetSHA(), c.Commit.GetMessage())
		if commit == nil {
			continue
		}
		commits = append(commits, *commit)
	}
	return commits, nil
}

func (client *githubWebhookHelper) updateRepoStatus(repo *github.Repository, sha, state string) error {

	var description string
	switch state {
	case StatusFailure:
		description = "commit was improperly formatted"
	case StatusSuccess:
		description = "commit looks good"
	case StatusPending:
		description = "beginning commit format validation"
	case StatusError:
		description = "there was a problem validating commit format"
	default:
		return fmt.Errorf("repo status state %s is invalid", state)
	}
	creating := &github.RepoStatus{
		State:       github.String(state),
		Description: github.String(description),
		Context:     github.String(webhooks.WebhookContextPullRequest),
	}
	_, _, err := client.Repositories.CreateStatus(
		context.Background(),
		repo.Owner.GetLogin(),
		repo.GetName(),
		sha,
		creating,
	)
	if err != nil {
		return err
	}
	return nil
}

func handlePullRequestEvent(event *github.PullRequestEvent, apiToken string) {

	eventAction := event.GetAction()
	// only handle these specific actions
	if eventAction != "opened" && eventAction != "repoened" && eventAction != "synchronize" {
		return
	}
	client := newGithubWebhookHelper(apiToken)
	commits, err := client.getPrCommits(event, apiToken)

	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	pullRequstNumber := event.PullRequest.GetNumber()
	lastCommit := commits[len(commits)-1]
	buildStatusSha := lastCommit.Hash

	zap.L().Info("validating commit format for pull request", zap.Int("pull_request", pullRequstNumber))
	err = client.updateRepoStatus(event.Repo, buildStatusSha, StatusPending)
	if err != nil {
		zap.L().Error(err.Error())
		return
	}

	iviper := viper.New()
	iviper.SetConfigType("toml")
	querier := changelog.NewGithubQuerier(event.Repo.GetHTMLURL(), apiToken)

	if config, err := querier.GetConfig(); err == nil {
		iviper.ReadConfig(config)
	} else {
		zap.L().Warn(err.Error())
	}

	sectionAliasMap := changelog.MergeSectionAliasMaps(
		changelog.NewSectionAliasMap(),
		iviper.GetStringMapStringSlice("sections"),
	)

	commits = changelog.FilterCommits(
		commits,
		sectionAliasMap.Grep(),
		false,
	)
	commits = changelog.FormatCommits(commits, sectionAliasMap)

	if len(commits) < event.PullRequest.GetCommits() {
		zap.L().Info("failed to validate commit format for pull request", zap.Int("pull_request", pullRequstNumber))
		err := client.updateRepoStatus(event.Repo, buildStatusSha, StatusFailure)
		if err != nil {
			zap.L().Error(err.Error())
		}
		return
	}

	// everything looks good
	err = client.updateRepoStatus(event.Repo, buildStatusSha, StatusSuccess)
	if err != nil {
		zap.L().Error(err.Error())
		return
	}
	zap.L().Info("validated commit for pull request", zap.Int("pull_request", pullRequstNumber))
	return
}

func GithubWebhook(secret, apiToken string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := github.ValidatePayload(r, []byte(secret))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		event, err := github.ParseWebHook(github.WebHookType(r), payload)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}

		if evt, ok := event.(*github.PullRequestEvent); ok {
			go handlePullRequestEvent(evt, apiToken)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "unable to process event type")
			return
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "{}")
		return
	}
}
