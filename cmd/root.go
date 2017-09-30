// Copyright © 2017 Skuid <devops@skuid.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/skuid/changelog/src/changelog"
	"github.com/skuid/changelog/src/linkStyle"
	"github.com/skuid/changelog/src/writer"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var providers = []string{
	"local",
	"github",
}

var (
	subtitle          = flag.String("subtitle", "", "The release subtitle")
	changelogFile     = flag.String("changelog", "", "The file to write. Defaults to STDOUT if not set.")
	fromLatestTag     = flag.Bool("from-latest-tag", false, "If you use tags, set to true to get changes from latest tag.")
	fromCommit        = flag.StringP("from", "f", "", "The beginning commit. Defaults to beginning of the repository history")
	toCommit          = flag.StringP("to", "t", "HEAD", "The last commit.")
	sinceTime         = flag.String("since", "", "Show commits more recent than a specific date. Use RFC3339 time '2017-08-01T00:00:00Z'. Takes precedence over to/from.")
	untilTime         = flag.String("until", "", "Show commits older than a specific date. Defaults to current time if not set, but --since is. Takes precedence over to/from.")
	version           = flag.StringP("version", "v", "", "The version you are creating")
	repoLink          = flag.StringP("repo", "r", "", "The repository URL. Defaults to `$(git remote get-url origin)` if using a local provider")
	includeAllCommits = flag.Bool("include-all", false, "Set to true to include all commits in the changelog. Commit messages that cannot be parsed will be placed in a section titled \"Unknown\".")

	provider = flag.StringP("provider", "p", "local", fmt.Sprintf(`The provider to use. Must be one of %s`, strings.Join(providers, ", ")))

	token = flag.String("token", "", "API token for remote provider. Only applies to github provider")

	gitDir   = flag.String("git-dir", "", "The path to the git directory. If no '--repo' is set, defaults to `$(pwd)/.git`. Only applies to local provider")
	workTree = flag.String("work-tree", "", "The path to the directory containing the .git directory. Only applies to local provider.")

	// outfile = flagSet.String("outfile", "", "") // TODO to maintain clog compatibility
	// infile = flagSet.String("infile", "", "") // TODO to maintain clog compatibility
	// outputFormat = flagSet.String("output-format", "markdown", "markdown or json") // TODO to maintain clog compatibility

	// debug = flagSet.Bool("debug", false, "Set to output debug logging")
)

func validateProvider(provider string) error {
	for i := range providers {
		if provider == providers[i] {
			return nil
		}
	}
	return fmt.Errorf("Provider %s not found! Must be one of %s", provider, strings.Join(providers, ", "))
}

func exitOnError(err error) {
	fmt.Printf("Fatal Error: %s", err.Error())
	os.Exit(1)
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate a Clog changelog",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateProvider(viper.GetString("provider")); err != nil {
			exitOnError(err)
		}

		var style linkStyle.Style
		var querier changelog.Querier

		switch viper.GetString("provider") {
		case "github":
			style = linkStyle.Github
			querier = changelog.NewGithubQuerier(viper.GetString("repo"), viper.GetString("token"))
		default:
			querier = changelog.NewLocalQuerier(viper.GetString("git-dir"), viper.GetString("work-tree"))
			repo := viper.GetString("repo")
			if len(repo) == 0 {
				repo, err := querier.GetOrigin()
				if err != nil {
					exitOnError(err)
				}
				viper.Set("repo", repo)
			}
			style = linkStyle.InferStyle(repo)
		}

		// Read in the `.clog.toml` file of the repo we're using
		if config, err := querier.GetConfig(); err == nil {
			err := viper.ReadConfig(config)
			if err != nil {
				exitOnError(err)
			}
		}

		c := changelog.ChangeLog{
			Repo:     viper.GetString("repo"),
			Version:  viper.GetString("version"),
			Subtitle: viper.GetString("subtitle"),
		}

		sectionAliasMap := changelog.MergeSectionAliasMaps(
			changelog.NewSectionAliasMap(),
			viper.GetStringMapStringSlice("sections"),
		)
		var commits changelog.Commits

		if len(viper.GetString("since")) > 0 || len(viper.GetString("until")) > 0 {
			if viper.GetString("until") == "" {
				viper.Set("until", time.Now().Format(time.RFC3339))
			}
			if viper.GetString("since") == "" {
				viper.Set("since", time.Unix(1, 0).Format(time.RFC3339))
			}

			until, err := time.Parse(time.RFC3339, viper.GetString("until"))
			if err != nil {
				exitOnError(err)
			}

			since, err := time.Parse(time.RFC3339, viper.GetString("since"))
			if err != nil {
				exitOnError(err)
			}

			commits, err = querier.GetCommitRange(since, until)
			if err != nil {
				exitOnError(errors.Wrap(err, "Could not get list of commits"))
			}
		} else {
			if viper.GetBool("from-latest-tag") {
				version, err := querier.GetLatestTag()
				if err != nil {
					exitOnError(errors.Wrap(err, "Could not get latest tag revision"))
				}
				viper.Set("from", version)
			}
			var err error
			commits, err = querier.GetCommits(viper.GetString("from"), viper.GetString("to"))
			if err != nil {
				exitOnError(errors.Wrap(err, "Could not get list of commits"))
			}
		}

		// Filter out commits if we're not including all
		if !viper.GetBool("include-all") {
			commits = changelog.FilterCommits(
				commits,
				sectionAliasMap.Grep(),
				viper.GetBool("include-all"),
			)
			// Set the proper CommitType on each commit from the sectionAliasMap
			commits = changelog.FormatCommits(commits, sectionAliasMap)
		} else {
			// Use all commits
			commits = changelog.TitleCommitType(commits, sectionAliasMap)
		}

		sectionMap := changelog.NewSectionMap(commits)

		// If we have an `order` key, use it to set the order on the sectionMap
		if order := viper.GetStringSlice("order"); len(order) > 0 {
			sectionMap.SetOrder(order)
		}

		w := writer.MarkdownWriter{Writer: os.Stdout}
		err := w.Generate(c, style, sectionMap)
		if err != nil {
			panic(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix("changelog")
	viper.AutomaticEnv()
	viper.SetConfigName(".clog.toml")
	viper.SetConfigType("toml")
	viper.BindPFlags(flag.CommandLine)
	viper.ReadInConfig()
}
