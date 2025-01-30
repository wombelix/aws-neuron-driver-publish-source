// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import "fmt"

const (
	REPO_URL        = "https://yum.repos.neuron.amazonaws.com"
	REPOMD_FILE     = "repodata/repomd.xml"
	GPG_PUB_FILE    = "GPG-PUB-KEY-AMAZON-AWS-NEURON.PUB"
	ARCHIVE_FOLDER  = "archive"
	SOURCE_FOLDER   = "src"
	CHECKSUM_SUFFIX = "sha256"
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	repofilesxml := ProcessRepomd()
	fmt.Println(string(repofilesxml["repomd"]))
}
