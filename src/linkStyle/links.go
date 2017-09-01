package linkStyle

import (
	"fmt"
	"sort"
	"strings"
)

// Style is type for generating issue and commit links
type Style string

const (
	// Github is for github links
	Github Style = "github"
	// Gitlab is for gitlab links
	Gitlab Style = "gitlab"
	// Stash is for stash links
	Stash Style = "stash"
	// Bitbucket is for bitbucket links
	Bitbucket Style = "bitbucket"
	// Cgit is for cgit links
	Cgit Style = "cgit"
)

// InferStyle tries to guess which style to use based on a repository URL
func InferStyle(repoURL string) Style {
	switch {
	case strings.Contains("github.com", repoURL):
		return Github
	case strings.Contains("gitlab.com", repoURL):
		return Gitlab
	case strings.Contains("bitbucket.org", repoURL):
		return Stash
	default:
		return Github
	}
}

// SupportedStyles returns a printable string of supported styles.
func SupportedStyles() string {
	styles := []string{
		string(Github),
		string(Gitlab),
		string(Stash),
		string(Bitbucket),
		string(Cgit),
	}
	sort.Strings(styles)
	return strings.Join(styles, ", ")
}

// IssueLink returns an issue link for a given Style
func (s Style) IssueLink(issue, repo string) string {
	var format string
	switch s {
	case Github:
		format = "%s/issues/%s"
	case Gitlab, Bitbucket:
		format = "%s/issues/%s"
	default:
		format = "%s"
	}
	return fmt.Sprintf(format, repo, issue)
}

// CommitLink returns an issue link for a given Style
func (s Style) CommitLink(hash, repo string) string {
	var format string
	switch s {
	case Github, Gitlab:
		format = "%s/commit/%s"
	case Stash, Bitbucket:
		format = "%s/commits/%s"
	case Cgit:
		format = "%s/commit/?id=%s"
	default:
		format = "%s"
	}
	return fmt.Sprintf(format, repo, hash)
}
