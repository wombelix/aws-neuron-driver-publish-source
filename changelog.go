// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func parseChangelog() map[string][]string {
	file := fmt.Sprintf("%s/%s", *archiveFolderName, CHANGELOG_FILE)

	f, err := os.Open(file)
	checkError(err)
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	scanner := bufio.NewScanner(f)

	changelog := make(map[string][]string)
	var version string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "Neuron Driver release") {
			versionRegEx := `\[(?P<version>.*)\]` // version = index 1
			re := regexp.MustCompile(versionRegEx)
			matches := re.FindStringSubmatch(line)
			if matches == nil {
				checkError(errors.New("could not parse version"))
			}
			version = matches[1]
		}

		if version != "" {
			changelog[version] = append(changelog[version], line)
		}
	}

	err = scanner.Err()
	checkError(err)

	return changelog
}

func ProcessChangelog() map[string][]string {
	err := downloadFile(CHANGELOG_URL, CHANGELOG_FILE, *archiveFolderName)
	checkError(err)

	if gitWorktreeModified(*gitRepoPath) {
		featureBranch := "feat-update-archive-release-notes"
		commitMsg := "feat: Update archive - release notes\n\n"
		featureBranchCommitMerge(*gitRepoPath, featureBranch, commitMsg)
	}

	return parseChangelog()
}
