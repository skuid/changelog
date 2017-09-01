package changelog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type githubQuerier struct {
	repo   string
	client *github.Client
}

// NewGithubQuerier queries Github for commits
func NewGithubQuerier(repo, token string) Querier {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	return githubQuerier{repo, client}
}

func (g githubQuerier) getOwnerRepo() (owner string, repo string) {
	regex := regexp.MustCompile(`(?:.*)github.com[\/\:]([\w-]*)\/([\w-]*)(?:\.git)?`)
	if capture := regex.FindStringSubmatch(g.repo); len(capture) > 2 {
		return capture[1], capture[2]
	}
	return "", ""
}

func (g githubQuerier) GetOrigin() (string, error) {
	return g.repo, nil
}

func (g githubQuerier) listCommits(opts *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error) {
	owner, repo := g.getOwnerRepo()

	return g.client.Repositories.ListCommits(
		context.Background(),
		owner,
		repo,
		opts,
	)
}

func (g githubQuerier) compareCommits(from, to string) (*github.CommitsComparison, *github.Response, error) {
	owner, repo := g.getOwnerRepo()

	return g.client.Repositories.CompareCommits(
		context.Background(),
		owner,
		repo,
		from,
		to,
	)
}

func (g githubQuerier) GetCommitRange(since, until time.Time) (Commits, error) {
	allGhCommits := []*github.RepositoryCommit{}

	opt := &github.CommitsListOptions{Since: since, Until: until}
	for {
		ghCommits, resp, err := g.listCommits(opt)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		allGhCommits = append(allGhCommits, ghCommits...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	commits := Commits{}
	for _, c := range allGhCommits {
		commit := NewCommit(c.GetSHA(), c.Commit.GetMessage())
		if commit == nil {
			continue
		}
		commits = append(commits, *commit)
	}

	return commits, nil
}

func (g githubQuerier) GetCommits(from, to string) (Commits, error) {
	allGhCommits := []*github.RepositoryCommit{}

	if to != "HEAD" {
		comparison, _, err := g.compareCommits(from, to)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if *comparison.TotalCommits >= 250 {
			// We've hit GH's comparison limit
			fmt.Fprint(os.Stderr, "Github limits commit comparison to 250 commits! Result may be truncated")
		}
		for _, commit := range comparison.Commits {
			allGhCommits = append(allGhCommits, &commit)
		}
	} else {
		opt := &github.CommitsListOptions{SHA: from}
		for {
			ghCommits, resp, err := g.listCommits(opt)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			allGhCommits = append(allGhCommits, ghCommits...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	}

	commits := Commits{}
	for _, c := range allGhCommits {
		commit := NewCommit(c.GetSHA(), c.Commit.GetMessage())
		if commit == nil {
			continue
		}
		commits = append(commits, *commit)
	}

	return commits, nil
}

func (g githubQuerier) GetLatestCommit() (string, error) {
	ghCommits, _, err := g.listCommits(&github.CommitsListOptions{})
	if err != nil {
		return "", errors.WithStack(err)
	}
	if len(ghCommits) == 0 {
		return "", errors.New("No commits in response")
	}
	return ghCommits[0].GetSHA(), nil
}

func (g githubQuerier) listTags() ([]*github.RepositoryTag, error) {
	owner, repo := g.getOwnerRepo()

	tags, _, err := g.client.Repositories.ListTags(
		context.Background(),
		owner,
		repo,
		&github.ListOptions{},
	)
	return tags, err
}

func (g githubQuerier) GetLatestTag() (string, error) {
	tags, err := g.listTags()
	if err != nil {
		return "", errors.WithStack(err)
	}
	if len(tags) == 0 {
		return "", errors.New("No tags in response")
	}
	return tags[0].Commit.GetSHA(), nil
}

func (g githubQuerier) GetLatestTagVersion() (string, error) {
	tags, err := g.listTags()
	if err != nil {
		return "", errors.WithStack(err)
	}
	if len(tags) == 0 {
		return "", errors.New("No tags in response")
	}
	return tags[0].Commit.GetSHA(), nil
}

func (g githubQuerier) GetConfig() (io.Reader, error) {
	owner, repo := g.getOwnerRepo()
	fileContent, _, _, err := g.client.Repositories.GetContents(
		context.Background(),
		owner,
		repo,
		"/.clog.toml",
		nil,
	)
	if err != nil {
		return nil, err
	}
	content, err := fileContent.GetContent()
	if err != nil {
		return nil, err
	}
	return bytes.NewBufferString(content), nil
}
