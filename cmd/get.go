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
	"io"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

//add -f to force creation of dir
var nocheck bool
var globalGet bool

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a package from the internet",
	Long: `duck get <package> looks for <package> in your configured
repositories and download it from the first repository (based on
repositories array index) where its available.

Package are named following this pattern : 'author/name' (ie: 'snwfdhmp/go')
To try: duck get snwfdhmp/go`,
	Run: func(cmd *cobra.Command, args []string) {
		fs := afero.NewOsFs()

		var pkgsPath string
		var err error

		if globalGet {
			path, err := DuckGlobalConfPath()
			if err != nil {
				color.Red("Unable to get duck global configuration path. Error: " + err.Error())
				return
			}
			pkgsPath = path + "/packages"
			exists, err := afero.Exists(fs, pkgsPath)
			if err != nil {
				color.Red("Unable to check if global package storage exists. Error: " + err.Error())
				return
			}
			if !exists {
				fmt.Println("This user does not have a global package storage yet. Creating one ...")
				err := fs.MkdirAll(pkgsPath, 0755)
				if err != nil {
					color.Red("Unable to create global package storage. Error: " + err.Error())
					return
				}
			}
		} else {
			err := loadProjectConfig()
			if err != nil {
				color.Red("Unable to load project configuration.")
				color.Red("Error: " + err.Error())
				return
			}

			pkgs, err := projectCfg.GetSection("packages")
			if err != nil {
				color.Red("Unable to get 'packages' section into configuration file")
				color.Red("Error: " + err.Error())
				return
			}

			path, err := pkgs.GetKey("directory")
			if err != nil {
				color.Red("Unable to get the 'directory' key from 'packages' section into configuration file")
				color.Red("Error: " + err.Error())
				return
			}

			pkgsPath = ".duck/" + path.String()
		}

		for i := 0; i < len(args); i++ {
			// if args[i][len(args[i])-1] == "/" { //delete '/' if in last position, should be tested before use
			// 	args[i] = args[i][:len(args[i])-1]
			// }
			arr := strings.Split(args[i], "/")
			var out afero.File
			currentPath := pkgsPath
			for j := 0; j < len(arr); j++ {
				currentPath += "/" + arr[j]
				if j < len(arr)-1 {
					if !nocheck {
						exists, err := afero.Exists(fs, currentPath)
						if err != nil {
							color.Red("Could not test whether '" + currentPath + "' exists or not. (not implemented: use --no-check to force)")
							color.Red("Error: " + err.Error())
							return
						}
						if exists {
							continue
						}
					}
					err = fs.Mkdir(currentPath, 0777)
					if err != nil {
						color.Red("Could not create '" + currentPath)
						color.Red("Error: " + err.Error())
						if !nocheck {
							return
						}
					}
				} else {
					if !nocheck {
						currentPath += ".duckpkg.ini"
						exists, err := afero.Exists(fs, currentPath)
						if err != nil {
							color.Red("Could not test whether '" + currentPath + "' exists or not. (not implemented: use --no-check to force)")
							color.Red("Error: " + err.Error())
							return
						}
						if exists && !force {
							color.Red("This package seems to be already installed. Use -f to install over")
							return
						}
					}
					out, err = fs.Create(currentPath)
					if err != nil {
						color.Red("Could not create '" + currentPath)
						color.Red("Error: " + err.Error())
						if !nocheck {
							return
						}
					}
				}
			}
			defer out.Close()
			repoColor := color.New(color.FgCyan).Sprint
			pkgColor := color.New(color.FgYellow).Sprint
			installed := false

			repos, err := getRepos()
			if err != nil {
				color.Red("Could not load repos : " + err.Error())
				return
			}
			if len(repos) == 0 {
				fmt.Println("No repository configured. Installing default repository...")
				_, err = addRepo(DefaultRepoName, DefaultRepoURL)
				if err != nil {
					color.Red("Could not add default repo : " + err.Error())
					return
				}
				repos, err = getRepos()
				if err != nil {
					color.Red("Could not load repos : " + err.Error())
					return
				}
			}
			for name, url := range repos {
				pkgUrl := fmt.Sprintf("%s%s.duckpkg.ini", url, args[i])
				resp, err := http.Get(pkgUrl) //test errors
				if err != nil {
					color.Red("Could not download from '" + pkgUrl + "'")
					color.Red("Error: " + err.Error())
					continue
				}
				defer resp.Body.Close()

				_, err = io.Copy(out, resp.Body)
				if err != nil {
					color.Red("Could not write file")
					continue
				}

				color.Green("Successfully installed '" + pkgColor(args[i]) + color.New(color.FgGreen).Sprint("' from ") + repoColor(name))
				installed = true
				break
			}
			if !installed {
				color.Red("Could not install " + pkgColor(args[i]))
				continue
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolVarP(&force, "force", "f", false, "replace package if existing")
	getCmd.Flags().BoolVarP(&globalGet, "global", "g", false, "install package for user instead of project")
	getCmd.Flags().BoolVar(&nocheck, "no-check", false, "skip file/folder existance checking")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}