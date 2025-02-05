// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/hashicorp/go-version"
	"github.com/sassoftware/go-rpmutils"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

func validateSignature(rpmfile string, gpgpubfile string) ([]*rpmutils.Signature, error) {
	keyfile, err := os.Open(gpgpubfile)
	checkError(err)
	defer func(keyfile *os.File) {
		_ = keyfile.Close()
	}(keyfile)

	keyring, err := openpgp.ReadArmoredKeyRing(keyfile)
	checkError(err)

	file, err := os.Open(rpmfile)
	checkError(err)
	defer func(f *os.File) {
		_ = f.Close()
	}(file)

	_, signature, err := rpmutils.Verify(file, keyring)

	if err != nil {
		return nil, err
	}

	// edge case, 'rpmutils.Verify' respond with empty Signer if pub key was empty / nil, means verify is failed
	for _, sig := range signature {
		if sig.Signer == nil {
			return nil, fmt.Errorf("public GPG key missing or invalid, RPM signature validity cannot be verified")
		}
	}

	return signature, nil
}

func createTempDir(pattern string) string {
	tmpDir, err := os.MkdirTemp("", pattern)
	checkError(err)

	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	return tmpDir
}

func ProcessRpm(packages map[string]*PrimaryPackage, changelog map[string][]string) {
	var err error

	downloadFile(fmt.Sprintf("%s/%s", RepoUrl, GpgPubFile), GpgPubFile, *archiveFolderName)

	if gitWorktreeModified(*gitRepoPath) {
		featureBranch := "feat-update-archive-gpg-pub-key"
		commitMsg := "feat: Update archive - GPG Public Key\n\n"
		featureBranchCommitMerge(*gitRepoPath, featureBranch, commitMsg)
	}

	rpmFolder := fmt.Sprintf("%s/%s", *archiveFolderName, *archiveRpmFolderName)

	versionsSlice := make([]string, 0, len(packages))
	for k := range packages {
		versionsSlice = append(versionsSlice, k)
	}

	versions := make([]*version.Version, len(versionsSlice))
	for i, ver := range versionsSlice {
		v, _ := version.NewVersion(ver)
		versions[i] = v
	}
	sort.Sort(version.Collection(versions))

	gitTags := getGitTags(*gitRepoPath)

	for _, ver := range versions {
		// No need to process a version that exist as git tag already
		if slices.Contains(gitTags, ver.Original()) {
			continue
		}

		pkg := packages[ver.Original()]

		filename := filepath.Base(pkg.Location.HRef)
		rpmfilepath := filepath.Join(rpmFolder, filename)

		downloadFile(
			fmt.Sprintf("%s/%s", RepoUrl, filename),
			filename,
			rpmFolder)

		// panic if gpg validation fails
		rpmSignatures := make(map[string][]*rpmutils.Signature)
		rpmSignatures[pkg.Version.Version], err = validateSignature(
			rpmfilepath,
			fmt.Sprintf("%s/%s", *archiveFolderName, GpgPubFile))
		checkError(err)

		// panic if checksum validation fails
		err = verifyChecksum(rpmfilepath, pkg.Checksum.Value)
		checkError(err)

		err = writeChecksumToFile(rpmFolder, filename, pkg.Checksum.Value)
		checkError(err)

		// Handle RPM file
		var f *os.File
		f, err = os.Open(rpmfilepath)
		checkError(err)

		var rpm *rpmutils.Rpm
		rpm, err = rpmutils.ReadRpm(f)
		checkError(err)

		tmpDir := createTempDir("aws-neuron-driver-")

		err = rpm.ExpandPayload(tmpDir)
		checkError(err)

		err = os.RemoveAll(*sourceFolderName)
		checkError(err)
		err = os.MkdirAll(*sourceFolderName, 0755)
		checkError(err)

		var srcFolder []os.DirEntry
		srcFolder, err = os.ReadDir(filepath.Join(tmpDir, "/usr/src/"))
		checkError(err)

		var srcPathNeuron string
		for _, entry := range srcFolder {
			if entry.IsDir() && strings.Contains(entry.Name(), "aws-neuron") {
				srcPathNeuron = filepath.Join(tmpDir, "/usr/src/", entry.Name())
			}
		}

		err = os.CopyFS(*sourceFolderName, os.DirFS(srcPathNeuron))
		checkError(err)

		if gitWorktreeModified(*gitRepoPath) {
			featureBranch := "feat-neuron-driver-release"

			var releaseNotes string
			if val, ok := changelog[pkg.Version.Version]; ok {
				// Drop 'Neuron Driver release [x.y.z]' header
				val = slices.Delete(val, 0, 2)

				releaseNotes += strings.Join(val, "\n")
				releaseNotes = strings.TrimLeft(releaseNotes, "\n")
				releaseNotes = strings.TrimRight(releaseNotes, "\n")

				releaseNotes = "\nRelease Notes\n-------------\n" + releaseNotes
				releaseNotes += "\n-------------\n"
			}

			var buildTimestamp int64
			buildTimestamp, err = strconv.ParseInt(pkg.Time.Build, 10, 64)
			checkError(err)

			commitMsg := fmt.Sprintf(`feat: Neuron Driver %s

Source code extracted from file: %s
RPM Downloaded from repository: %s
%s
Metadata
--------
Package: %s
Version: %s
License: %s
Summary: %s
Description: %s
Filename: %s
Checksum: %s
Buildhost: %s
Buildtime: %s
GPG key primary uid: %s
GPG key creation time: %s
GPG key fingerprint: %s
GPG check: OK
SHA256 check: OK
--------

`,
				pkg.Version.Version,
				pkg.Location.HRef,
				RepoUrl,
				releaseNotes,
				pkg.Name,
				pkg.Version.Version,
				pkg.Format.License,
				pkg.Summary,
				pkg.Description,
				pkg.Location.HRef,
				pkg.Checksum.Value,
				pkg.Format.BuildHost,
				time.Unix(buildTimestamp, 0),
				rpmSignatures[pkg.Version.Version][0].PrimaryName,
				rpmSignatures[pkg.Version.Version][0].Signer.PrimaryKey.CreationTime.UTC(),
				strings.ToUpper(hex.EncodeToString(rpmSignatures[pkg.Version.Version][0].Signer.PrimaryKey.Fingerprint[:])),
			)

			featureBranchCommitMergeTag(*gitRepoPath, featureBranch, commitMsg, pkg.Version.Version)
		}

	}

}
