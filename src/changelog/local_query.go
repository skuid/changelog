package changelog

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type localQuerier struct {
	GitDir      string `toml:"git_dir"`
	GitWorkTree string `toml:"git_work_tree"`
	Format      string
}

func (l localQuerier) getWorkdir() string {
	if len(l.GitDir) > 0 {
		return filepath.Dir(l.GitDir)
	}
	if len(l.GitWorkTree) > 0 {
		return l.GitWorkTree
	}
	return "."
}

// NewLocalQuerier returns a querier that queries off of a local git repostiroy
func NewLocalQuerier(gitDir, workTree string) Querier {
	return localQuerier{
		gitDir,
		workTree,
		`%H%n%s%n%b%n==END==`,
	}
}

func (l *localQuerier) getGitWorkTree() string {
	// Check if user supplied a local git dir and working tree
	if l.GitDir != "" && l.GitWorkTree != "" {
		// user supplied both
		return fmt.Sprintf("--work-tree=%s", l.GitWorkTree)
	} else if l.GitWorkTree != "" && l.GitDir == "" {
		return fmt.Sprintf("--work-tree=%s", filepath.Dir(l.GitWorkTree))
	}
	return ""
}

func (l *localQuerier) getGitDir() string {
	if l.GitDir == "" && l.GitWorkTree == "" {
		return ""
	} else if l.GitDir != "" {
		return fmt.Sprintf("--git-dir=%s", l.GitDir)
	}
	return fmt.Sprintf("--git-dir=%s", filepath.Join(l.GitWorkTree, ".git"))
}

func (l localQuerier) gitCommandFactory(args ...string) *exec.Cmd {
	args = append([]string{l.getGitDir(), l.getGitWorkTree()}, args...)
	realArgs := []string{}
	for _, argument := range args {
		if argument != "" {
			realArgs = append(realArgs, argument)
		}
	}
	// fmt.Println(realArgs)
	return exec.Command("git", realArgs...)
}

func (l localQuerier) GetOrigin() (string, error) {
	args := []string{
		"remote",
		"get-url",
		"origin",
	}
	cmd := l.gitCommandFactory(args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", errors.WithStack(err)
	}

	origin := strings.TrimSpace(out.String())
	if strings.HasPrefix(origin, "git@") {
		origin = fmt.Sprintf("https://%s", strings.Replace(strings.TrimSuffix(origin[4:], ".git"), ":", "/", -1))
	}
	return origin, nil
}

// GetLatestCommit returns the latest commit
func (l localQuerier) GetLatestCommit() (string, error) {
	args := []string{
		"rev-list",
		"HEAD",
	}
	cmd := l.gitCommandFactory(args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return out.String(), nil
}

// GetLatestTag returns the latest tag
func (l localQuerier) GetLatestTag() (string, error) {
	args := []string{
		"rev-list",
		"--tags",
		"--max-count=1",
	}
	cmd := l.gitCommandFactory(args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return strings.TrimSpace(out.String()), nil
}

// GetLatestTagVersion returns the latest tag version
func (l localQuerier) GetLatestTagVersion() (string, error) {
	args := []string{
		"describe",
		"--tags",
		"--abbrev=0",
	}
	cmd := l.gitCommandFactory(args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return out.String(), nil
}

func (l localQuerier) parseRawCommit(repo, commitStr string) *Commit {
	lines := strings.Split(commitStr, "\n")
	if len(lines) < 2 {
		return nil
	}
	return NewCommit(lines[0], strings.Join(lines[1:], "\n"))

}

// GetCommits returns a slice of commits
func (l localQuerier) GetCommits(from, to string) (Commits, error) {
	repo, err := l.GetOrigin()
	if err != nil {
		return nil, err
	}

	if from != "" {
		from = fmt.Sprintf("%s..", from)
	}

	args := []string{
		"log",
		"-E",
		fmt.Sprintf(`--format=%s`, l.Format),
		fmt.Sprintf("%s%s", from, to),
	}
	cmd := l.gitCommandFactory(args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	commitGroups := strings.Split(out.String(), "\n==END==\n")

	var commits Commits
	for _, com := range commitGroups {
		if len(com) == 0 {
			continue
		}
		commit := l.parseRawCommit(repo, com)
		if commit == nil {
			continue
		}
		commits = append(commits, *commit)

	}
	return commits, nil
}

// GetCommits returns a slice of commits
func (l localQuerier) GetCommitRange(since, until time.Time) (Commits, error) {
	repo, err := l.GetOrigin()
	if err != nil {
		return nil, err
	}

	args := []string{
		"log",
		"-E",
		fmt.Sprintf(`--format=%s`, l.Format),
		fmt.Sprintf("--since=%s", since.Format(time.RFC3339)),
		fmt.Sprintf("--until=%s", until.Format(time.RFC3339)),
	}
	cmd := l.gitCommandFactory(args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	commitGroups := strings.Split(out.String(), "\n==END==\n")

	var commits Commits
	for _, com := range commitGroups {
		if len(com) == 0 {
			continue
		}
		commit := l.parseRawCommit(repo, com)
		if commit == nil {
			continue
		}
		commits = append(commits, *commit)

	}
	return commits, nil
}

// GetConfig returns a reader for the clog config
func (l localQuerier) GetConfig() (io.Reader, error) {
	dir := l.getWorkdir()
	content, err := ioutil.ReadFile(filepath.Join(dir, ".clog.toml"))
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(content), nil
}
