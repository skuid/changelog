package changelog_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/skuid/changelog/src/changelog"
)

func TestNewSectionMap(t *testing.T) {
	commit := changelog.Commit{
		Hash:       "029aafdc7579af19b3ce6acf0ce245a230633953",
		Subject:    "Initial Commit",
		Component:  "README",
		Closes:     []string{},
		Breaks:     []string{},
		CommitType: "Features",
	}

	cases := []struct {
		commits changelog.Commits
		want    changelog.SectionMap
	}{
		{
			changelog.Commits{commit},
			changelog.SectionMap{
				Sections: map[string]changelog.ComponentMap{
					"Features": changelog.ComponentMap{
						"README": changelog.Commits{commit},
					},
				},
			},
		},
	}

	for _, c := range cases {
		got := changelog.NewSectionMap(c.commits)
		if !reflect.DeepEqual(got.Sections, c.want.Sections) {
			errorDiff(t, "SectionMaps not equal!", fmt.Sprintf("'%v'", c.want.Sections), fmt.Sprintf("'%v'", got.Sections))
		}

	}
}
