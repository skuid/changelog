package changelog

import (
	"io"
	"time"
)

// Querier is an interface for the functions needed to generate a changelog
// from a git repository
type Querier interface {
	GetCommits(from, to string) (Commits, error)
	GetCommitRange(from, to time.Time) (Commits, error)
	GetOrigin() (string, error)
	GetLatestCommit() (string, error)
	GetLatestTag() (string, error)
	GetLatestTagVersion() (string, error)
	GetConfig() (io.Reader, error)
}
