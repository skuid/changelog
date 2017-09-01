package changelog_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/skuid/changelog/src/changelog"
	"github.com/skuid/changelog/src/linkStyle"
)

func errorDiff(t *testing.T, message, expected, got string) {
	t.Helper()
	t.Errorf("%s!\nExpected\n\t%s\nGot\n\t%s", message, expected, got)
}

func TestSummary(t *testing.T) {
	repo := "https://github.com/skuid/changelog"
	style := linkStyle.Github

	cases := []struct {
		commit changelog.Commit
		want   string
	}{
		{
			changelog.Commit{
				Hash:       "029aafdc7579af19b3ce6acf0ce245a230633953",
				Subject:    "Initial Commit",
				Component:  "README",
				CommitType: "feat",
			},
			"Initial Commit ([029aafdc](https://github.com/skuid/changelog/commit/029aafdc7579af19b3ce6acf0ce245a230633953))",
		},
		{
			changelog.Commit{
				Hash:       "029aafdc7579af19b3ce6acf0ce245a230633953",
				Subject:    "Initial Commit",
				Component:  "README",
				CommitType: "feat",
				Closes:     []string{"1", "2"},
			},
			"Initial Commit ([029aafdc](https://github.com/skuid/changelog/commit/029aafdc7579af19b3ce6acf0ce245a230633953)), closes [#1](https://github.com/skuid/changelog/issues/1) [#2](https://github.com/skuid/changelog/issues/2)",
		},
		{
			changelog.Commit{
				Hash:       "029aafdc7579af19b3ce6acf0ce245a230633953",
				Subject:    "Initial Commit",
				Component:  "README",
				CommitType: "feat",
				Closes:     []string{"2"},
				Breaks:     []string{"1"},
			},
			"Initial Commit ([029aafdc](https://github.com/skuid/changelog/commit/029aafdc7579af19b3ce6acf0ce245a230633953)), closes [#2](https://github.com/skuid/changelog/issues/2), breaks [#1](https://github.com/skuid/changelog/issues/1)",
		},
	}

	for i := range cases {
		got := cases[i].commit.Summary(repo, style)
		if got != cases[i].want {
			errorDiff(t, "Commit summary failed", cases[i].want, got)
		}
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func commitsEqual(a, b changelog.Commit) bool {
	switch {
	case a.Hash != b.Hash:
		return false
	case a.Subject != b.Subject:
		return false
	case a.Component != b.Component:
		return false
	case !stringSliceEqual(a.Closes, b.Closes):
		return false
	case !stringSliceEqual(a.Breaks, b.Breaks):
		return false
	default:
		return true
	}
}

func TestNewCommit(t *testing.T) {
	cases := []struct {
		hash    string
		message string
		want    *changelog.Commit
	}{
		{
			"029aafdc7579af19b3ce6acf0ce245a230633953",
			"feat(README): Initial Commit",
			&changelog.Commit{
				Hash:      "029aafdc7579af19b3ce6acf0ce245a230633953",
				Subject:   "Initial Commit",
				Component: "README",
				Closes:    []string{},
				Breaks:    []string{},
			},
		},
	}

	for i := range cases {
		got := changelog.NewCommit(cases[i].hash, cases[i].message)
		if !commitsEqual(*got, *cases[i].want) {
			errorDiff(t, "Commits not equal!", fmt.Sprintf("%v", cases[i].want), fmt.Sprintf("%v", got))
		}
	}
}
