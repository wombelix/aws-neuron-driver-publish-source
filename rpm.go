// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/hashicorp/go-version"
	"github.com/sassoftware/go-rpmutils"
	"os"
	"path/filepath"
	"slices"
	"sort"
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

func getFilenamesInDirectory(directory string) []string {
	folder, err := os.Open(directory)
	checkError(err)
	defer func(folder *os.File) {
		_ = folder.Close()
	}(folder)

	filenames, _ := folder.Readdirnames(0) // 0 to read all files and folders

	return filenames
}

func ProcessRpm(packages map[string]*PrimaryPackage, changelog map[string][]string) {
	var err error

	err = downloadFile(fmt.Sprintf("%s/%s", REPO_URL, GPG_PUB_FILE), GPG_PUB_FILE, *archiveFolderName)
	checkError(err)

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
		if slices.Contains(gitTags, ver.Original()) {
			continue
		}

		pkg := packages[ver.Original()]

		filename := filepath.Base(pkg.Location.HRef)
		rpmfilepath := filepath.Join(rpmFolder, filename)

		err = downloadFile(
			fmt.Sprintf("%s/%s", REPO_URL, filename),
			filename,
			rpmFolder)
		checkError(err)

		_, err := validateSignature(
			rpmfilepath,
			fmt.Sprintf("%s/%s", *archiveFolderName, GPG_PUB_FILE))
		checkError(err)

		err = verifyChecksum(rpmfilepath, pkg.Checksum.Value)
		checkError(err)

		err = writeChecksumToFile(rpmFolder, filename, pkg.Checksum.Value)
		checkError(err)

		break
	}

}
