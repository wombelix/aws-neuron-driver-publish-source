// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func downloadChangelog() error {
	resp, err := http.Get(CHANGELOG_URL)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("status code %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = os.MkdirAll(ARCHIVE_FOLDER, 0700)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s", ARCHIVE_FOLDER, CHANGELOG_FILE), body, 0644)
	if err != nil {
		return err
	}

	return err
}

func parseChangelog() map[string][]string {
	file := fmt.Sprintf("%s/%s", ARCHIVE_FOLDER, CHANGELOG_FILE)

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
	err := downloadChangelog()
	checkError(err)

	return parseChangelog()
}
