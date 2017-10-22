// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

	"github.com/fatih/color"
	"github.com/go-ini/ini"
	"github.com/spf13/cobra"
)

const (
	DefaultRepoName = "core"
	DefaultRepoURL  = "http://raw.githubusercontent.com/snwfdhmp/duck-core/master/"
)

// repoCmd represents the repo command
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Get list of your repositories",
	Run: func(cmd *cobra.Command, args []string) {
		repoColor := color.New(color.FgCyan).SprintFunc()
		urlColor := color.New(color.FgYellow).SprintFunc()

		repos, err := getRepos()
		if err != nil {
			color.Red("Could not load repos : " + err.Error())
			return
		}
		for name, url := range repos {
			fmt.Println("-", repoColor(name), "=>", urlColor(url))
		}
	},
}

func init() {
	RootCmd.AddCommand(repoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// repoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// repoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getRepos() (map[string]string, error) {
	data, err := getDuckData()
	if err != nil {
		return nil, err
	}

	reposSection, err := data.GetSection("repos")
	if err != nil {
		reposSection, err = data.NewSection("repos")
		if err != nil {
			return nil, err
		}
	}

	var repos map[string]string
	repos = make(map[string]string)

	for _, repo := range reposSection.Keys() {
		repos[repo.Name()] = repo.Value()
	}

	return repos, nil
}

func addRepo(name, url string) (*ini.Key, error) {
	data, err := getDuckData()
	if err != nil {
		return nil, err
	}

	reposSection, err := data.GetSection("repos")
	if err != nil {
		reposSection, err = data.NewSection("repos")
		if err != nil {
			return nil, err
		}
	}

	key, err := reposSection.NewKey(name, url)
	if err != nil {
		return nil, err
	}

	path, err := getDuckDataPath()
	if err != nil {
		return nil, err
	}
	err = data.SaveTo(path)

	return key, err
}