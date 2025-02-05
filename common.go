// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func downloadFile(url string, filename string, folder string) {
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

	err = os.MkdirAll(folder, 0700)
	checkError(err)

	err = os.WriteFile(fmt.Sprintf("%s/%s", folder, filename), body, 0644)
	checkError(err)
}

func normalizeFilename(filename string) string {
	if strings.Contains(filename, "-") {
		filename = dropChecksumFromFilename(filename)
	}

	// Drop folder from filename string
	if strings.Contains(filename, "/") {
		filename = filepath.Base(filename)
	}

	return filename
}

func dropChecksumFromFilename(filename string) string {
	// Drop Checksum prefix from filename
	// Example:
	//		057288a8dfecacaf588228e429c1511a3f1f3801b1d2fb4a068d5c14e3d1fb27-filelists.xml.gz
	// to
	//		filelists.xml.gz
	return strings.Split(filename, "-")[1]
}

func verifyChecksum(file string, checksum string) error {
	f, err := os.Open(file)
	checkError(err)

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	h := sha256.New()

	_, err = io.Copy(h, f)
	checkError(err)

	fileChecksum := hex.EncodeToString(h.Sum(nil))
	if fileChecksum == checksum {
		return nil
	}

	return fmt.Errorf("checksum %s does not match for file %s", fileChecksum, file)
}

func writeChecksumToFile(folder string, file string, checksum string) error {
	err := os.MkdirAll(folder, 0700)
	if err != nil {
		return err
	}

	file = fmt.Sprintf("%s.%s", file, ChecksumSuffix)

	err = os.WriteFile(fmt.Sprintf("%s/%s", folder, file), []byte(checksum), 0644)
	return err
}
