// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

const (
	RepoUrl        = "https://yum.repos.neuron.amazonaws.com"
	RepomdFile     = "repodata/repomd.xml"
	GpgPubFile     = "GPG-PUB-KEY-AMAZON-AWS-NEURON.PUB"
	ChecksumSuffix = "sha256"
	ChangelogUrl   = "https://raw.githubusercontent.com/aws-neuron/aws-neuron-sdk/refs/heads/master/release-notes/runtime/aws-neuronx-dkms/index.rst"
	ChangelogFile  = "release-notes-runtime-aws-neuronx-dkms.rst"
)

var (
	gitRepoPath          *string
	archiveFolderName    *string
	archiveRpmFolderName *string
	sourceFolderName     *string
	logger               *slog.Logger
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	gitRepoPath = flag.String("repopath", "./", "Path to the local git repository with the aws-neuron-driver source code.")
	archiveFolderName = flag.String("archive-folder", "archive", "Sub-folder in the git repo where to store and archive processed files.")
	archiveRpmFolderName = flag.String("archive-rpm-folder", "rpm", "Sub-folder within the archive directory to store processed RPM files.")
	sourceFolderName = flag.String("source-folder", "src", "Sub-folder in the git repo where to store the aws-neuron-driver source code.")
	logLevel := flag.String("loglevel", "INFO", "Log level, available options: INFO (default), DEBUG")
	flag.Parse()

	*archiveFolderName = fmt.Sprintf("%s/%s", *gitRepoPath, *archiveFolderName)
	*sourceFolderName = fmt.Sprintf("%s/%s", *gitRepoPath, *sourceFolderName)

	level := slog.LevelInfo
	if *logLevel == "DEBUG" {
		level = slog.LevelDebug
	}
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}))

	repofilesxml := ProcessRepomd()

	changelog := ProcessChangelog()

	primary := ProcessPrimary(repofilesxml["primary"])

	ProcessRpm(primary, changelog)

}
