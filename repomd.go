// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Repomd struct {
	XMLName  xml.Name `xml:"repomd"`
	Xmlns    string   `xml:"xmlns,attr"`
	Revision string   `xml:"revision"`
	Data     []Data   `xml:"data"`
}

type Data struct {
	XMLName      xml.Name     `xml:"data"`
	Type         string       `xml:"type,attr"`
	Checksum     Checksum     `xml:"checksum"`
	OpenChecksum OpenChecksum `xml:"open-checksum"`
	Location     Location     `xml:"location"`
	Timestamp    string       `xml:"timestamp"`
}

type Checksum struct {
	XMLName xml.Name `xml:"checksum"`
	Type    string   `xml:"type,attr"`
	Value   string   `xml:",innerxml"`
}

type OpenChecksum struct {
	XMLName xml.Name `xml:"open-checksum"`
	Type    string   `xml:"type,attr"`
	Value   string   `xml:",innerxml"`
}

type Location struct {
	XMLName xml.Name `xml:"location"`
	HRef    string   `xml:"href,attr"`
}

func downloadAndReturnFileContent(repourl string, file string, outFolder string) []byte {
	url := fmt.Sprintf("%s/%s", repourl, file)

	resp, err := http.Get(url)
	checkError(err)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("status code %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	checkError(err)

	err = os.MkdirAll(outFolder, 0700)
	checkError(err)

	outFile := normalizeFilename(file)

	err = os.WriteFile(fmt.Sprintf("%s/%s", outFolder, outFile), body, 0644)
	checkError(err)

	// Decompress file content before return
	if strings.Contains(outFile, ".gz") {
		reader := bytes.NewReader(body)

		gzipReader, err := gzip.NewReader(reader)
		checkError(err)

		defer func(gzipReader io.ReadCloser) {
			_ = gzipReader.Close()
		}(gzipReader)

		body, err = io.ReadAll(gzipReader)
		checkError(err)
	}

	return body
}

func parseRepomd(repomdxml []byte) *Repomd {
	// Unmarshal XML content into Repomd struct
	var repomd Repomd
	err := xml.Unmarshal(repomdxml, &repomd)
	checkError(err)

	return &repomd
}

func ProcessRepomd() map[string][]byte {
	var err error
	repofilesxml := make(map[string][]byte)

	repofilesxml["repomd"] = downloadAndReturnFileContent(RepoUrl, RepomdFile, *archiveFolderName)

	repomd := parseRepomd(repofilesxml["repomd"])

	for _, data := range repomd.Data {
		switch data.Type {
		case "primary", "filelists", "other":
			repofilesxml[data.Type] = downloadAndReturnFileContent(RepoUrl, data.Location.HRef, *archiveFolderName)

			filenameWithoutFolderAndChecksum := normalizeFilename(data.Location.HRef)
			err = writeChecksumToFile(*archiveFolderName, filenameWithoutFolderAndChecksum, data.Checksum.Value)
			checkError(err)

			file := fmt.Sprintf("%s/%s", *archiveFolderName, filenameWithoutFolderAndChecksum)
			err = verifyChecksum(file, data.Checksum.Value)
			checkError(err)
		}
	}

	if gitWorktreeModified(*gitRepoPath) {
		featureBranch := "feat-update-archive-repo-files"
		commitMsg := "feat: Update archive - repo files\n\n"
		featureBranchCommitMerge(*gitRepoPath, featureBranch, commitMsg)
	}

	return repofilesxml
}
