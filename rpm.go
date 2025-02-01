// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/sassoftware/go-rpmutils"
	"os"
	"path/filepath"
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

func ProcessRpm(packages *[]PrimaryPackage) map[string][]*rpmutils.Signature {
	var err error

	err = downloadFile(fmt.Sprintf("%s/%s", REPO_URL, GPG_PUB_FILE), GPG_PUB_FILE, ARCHIVE_FOLDER)
	checkError(err)

	rpms := make(map[string][]*rpmutils.Signature)
	for _, pkg := range *packages {
		filename := filepath.Base(pkg.Location.HRef)
		rpmfolder := fmt.Sprintf("%s/%s", ARCHIVE_FOLDER, ARCHIVE_RPM_FOLDER)
		rpmfilepath := filepath.Join(rpmfolder, filename)

		err = downloadFile(
			fmt.Sprintf("%s/%s", REPO_URL, filename),
			filename,
			rpmfolder)
		checkError(err)

		sigs, err := validateSignature(
			rpmfilepath,
			fmt.Sprintf("%s/%s", ARCHIVE_FOLDER, GPG_PUB_FILE))
		checkError(err)

		err = verifyChecksum(rpmfilepath, pkg.Checksum.Value)
		checkError(err)

		err = writeChecksumToFile(rpmfolder, filename, pkg.Checksum.Value)
		checkError(err)

		rpms[filename] = sigs
		break
	}

	rpmsJson, _ := json.MarshalIndent(rpms, "", "\t")
	fmt.Println(string(rpmsJson))

	return rpms
}
