// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func downloadAndReturnFileContent(repourl string, file string, outFolder string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", repourl, file)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("status code %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(outFolder, 0700)
	if err != nil {
		return nil, err
	}

	outFile := normalizeFilename(file)

	err = os.WriteFile(fmt.Sprintf("%s/%s", outFolder, outFile), body, 0644)
	if err != nil {
		return nil, err
	}

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

	return body, err
}

func writeStringToFile(content string, outFolder string, outFile string, suffix string) error {
	err := os.MkdirAll(outFolder, 0700)
	if err != nil {
		return err
	}

	outContent := []byte(content)

	outFile = normalizeFilename(outFile)

	if suffix != "" {
		outFile = fmt.Sprintf("%s.%s", outFile, suffix)
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s", outFolder, outFile), outContent, 0644)
	return err
}

func writeChecksumToFile(checksum string, outFolder string, outFile string) error {
	return writeStringToFile(checksum, outFolder, outFile, CHECKSUM_SUFFIX)
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

func parseRepomd(repomdxml []byte) (*Repomd, error) {
	// Unmarshal XML content into Repomd struct
	var repomd Repomd
	err := xml.Unmarshal(repomdxml, &repomd)
	if err != nil {
		return nil, err
	}
	return &repomd, nil
}

func verifyChecksum(folder string, file string, checksum string) error {
	filename := normalizeFilename(file)
	filename = fmt.Sprintf("%s/%s", folder, filename)

	f, err := os.Open(filename)
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

	return fmt.Errorf("Checksum %s does not match for file %s", fileChecksum, filename)
}

func ProcessRepomd() map[string][]byte {
	var err error
	repofilesxml := make(map[string][]byte)

	repofilesxml["repomd"], err = downloadAndReturnFileContent(REPO_URL, REPOMD_FILE, ARCHIVE_FOLDER)
	checkError(err)

	repomd, err := parseRepomd(repofilesxml["repomd"])
	checkError(err)

	for _, data := range repomd.Data {
		switch data.Type {
		case "primary", "filelists", "other":
			fmt.Println(data.Type)
			repofilesxml[data.Type], err = downloadAndReturnFileContent(REPO_URL, data.Location.HRef, ARCHIVE_FOLDER)
			checkError(err)

			err = writeChecksumToFile(data.Checksum.Value, ARCHIVE_FOLDER, data.Location.HRef)
			checkError(err)

			err = verifyChecksum(ARCHIVE_FOLDER, data.Location.HRef, data.Checksum.Value)
			checkError(err)
		}
	}

	return repofilesxml
}
