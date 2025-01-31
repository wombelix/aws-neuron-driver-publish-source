// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
)

const (
	REPO_URL        = "https://yum.repos.neuron.amazonaws.com"
	REPOMD_FILE     = "repodata/repomd.xml"
	GPG_PUB_FILE    = "GPG-PUB-KEY-AMAZON-AWS-NEURON.PUB"
	ARCHIVE_FOLDER  = "archive"
	SOURCE_FOLDER   = "src"
	CHECKSUM_SUFFIX = "sha256"
	CHANGELOG_URL   = "https://raw.githubusercontent.com/aws-neuron/aws-neuron-sdk/refs/heads/master/release-notes/runtime/aws-neuronx-dkms/index.rst"
	CHANGELOG_FILE  = "release-notes-runtime-aws-neuronx-dkms.rst"
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	repofilesxml := ProcessRepomd()
	//fmt.Println(string(repofilesxml["repomd"]))

	//changelog := ProcessChangelog()
	//for k, _ := range changelog {
	//	fmt.Println(k)
	//}

	primary := ProcessPrimary(repofilesxml["primary"])
	primaryJson, _ := json.MarshalIndent(primary, "", "\t")
	fmt.Println(string(primaryJson))
}
