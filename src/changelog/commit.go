package changelog

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/skuid/changelog/src/linkStyle"
)

var (
	// CommitRegex is used to parse the first line of commits
	CommitRegex = regexp.MustCompile(`^([^:\(]+?)(?:\(([^\)]*?)?\))?:(.*)`)
	// ClosesRegex is used to find any closes links
	ClosesRegex = regexp.MustCompile(`(?:Closes|Fixes|Resolves)\s((?:#(\d+)(?:,\s)?)+)`)
	// BreaksRegex is used to find any breaks links
	BreaksRegex = regexp.MustCompile(`(?:Breaks|Broke)\s((?:#(\d+)(?:,\s)?)+)`)
	// BreakingRegex is used to find anything that is a breaking change
	BreakingRegex = regexp.MustCompile(`(?i:breaking)`)
)

// FilterCommits only keeps commits that are to be included in the changelog
func FilterCommits(commits Commits, grep string, includeAll bool) Commits {
	response := Commits{}

	regex := regexp.MustCompile(grep)
	for i := range commits {
		if includeAll || (!includeAll && regex.MatchString(commits[i].rawCommitType)) {
			response = append(response, commits[i])
		}
	}
	return response
}

// FormatCommits sets the CommitType on each commit from the given SectionAliasMap
func FormatCommits(commits Commits, sectionAliasMap SectionAliasMap) Commits {
	for i := range commits {
		commits[i].CommitType = sectionAliasMap.SectionFor(commits[i].rawCommitType)
	}
	return commits
}

// TitleCommitType sets the CommitType on each commit to the title case of itself.
//
// This is used if you wanted change log sections that are not defined in a `.clog.toml`
func TitleCommitType(commits Commits, sectionAliasMap SectionAliasMap) Commits {
	for i := range commits {
		section := sectionAliasMap.SectionFor(commits[i].rawCommitType)
		if section == "Unknown" {
			commits[i].CommitType = strings.Title(commits[i].rawCommitType)
		} else {
			commits[i].CommitType = section
		}
	}
	return commits
}

// Commit is a struct for representing a git commit
type Commit struct {
	Hash          string
	Subject       string
	Component     string
	Closes        []string
	Breaks        []string
	rawCommitType string
	CommitType    string
}

// Summary generates a summary line for the commit used in the change log
func (c *Commit) Summary(repo string, style linkStyle.Style) string {
	shortHash := c.Hash[:8]
	commitLink := style.CommitLink(c.Hash, repo)

	response := fmt.Sprintf("%s ([%s](%s))", c.Subject, shortHash, commitLink)

	closesLinks := []string{}
	for _, closer := range c.Closes {
		closesLinks = append(
			closesLinks,
			fmt.Sprintf("[#%s](%s)", closer, style.IssueLink(closer, repo)),
		)
	}

	if len(closesLinks) > 0 {
		response += fmt.Sprintf(", closes %s", strings.Join(closesLinks, " "))
	}

	breaksLinks := []string{}
	for _, breaker := range c.Breaks {
		breaksLinks = append(
			breaksLinks,
			fmt.Sprintf("[#%s](%s)", breaker, style.IssueLink(breaker, repo)),
		)
	}
	if len(breaksLinks) > 0 {
		response += fmt.Sprintf(", breaks %s", strings.Join(breaksLinks, " "))
	}
	return response
}

// Commits is a slice of Commit
type Commits []Commit

// NewCommit creates a commit
func NewCommit(hash, message string) *Commit {
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {

		return nil
	}
	match := CommitRegex.FindStringSubmatch(lines[0])

	var (
		commitType string
		component  string
		subject    string
	)
	// TODO if commitType is set but component is not, capture that
	if len(match) < 4 {
		commitType = "Unknown"
		component = "Unknown"
		subject = lines[0]
	} else {
		commitType, component, subject = match[1], match[2], match[3]
	}

	var (
		closes []string
		breaks []string
	)
	for _, line := range lines {
		if capture := ClosesRegex.FindStringSubmatch(line); len(capture) > 2 {
			closes = append(closes, capture[2])
		}
		if capture := BreaksRegex.FindStringSubmatch(line); len(capture) > 2 {
			breaks = append(breaks, capture[2])
		} else if BreakingRegex.FindString(line) != "" {
			breaks = append(breaks, "")
		}
	}

	return &Commit{
		Hash:          hash,
		Subject:       strings.TrimSpace(subject),
		Component:     component,
		rawCommitType: commitType,
		Closes:        closes,
		Breaks:        breaks,
	}
}
