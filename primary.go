// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/xml"
)

type PrimaryMetadata struct {
	XMLName  xml.Name         `xml:"metadata"`
	Packages int              `xml:"packages,attr"`
	Package  []PrimaryPackage `xml:"package"`
}

type PrimaryPackage struct {
	XMLName     xml.Name        `xml:"package"`
	Type        string          `xml:"type,attr"`
	Name        string          `xml:"name"`
	Arch        string          `xml:"arch"`
	Version     PrimaryVersion  `xml:"version"`
	Checksum    PrimaryChecksum `xml:"checksum"`
	Summary     string          `xml:"summary"`
	Description string          `xml:"description"`
	Time        PrimaryTime     `xml:"time"`
	Location    PrimaryLocation `xml:"location"`
	Format      PrimaryFormat   `xml:"format"`
}

type PrimaryVersion struct {
	XMLName xml.Name `xml:"version"`
	Version string   `xml:"ver,attr"`
}

type PrimaryChecksum struct {
	XMLName xml.Name `xml:"checksum"`
	Type    string   `xml:"type,attr"`
	Value   string   `xml:",innerxml"`
}

type PrimaryTime struct {
	XMLName xml.Name `xml:"time"`
	File    string   `xml:"file,attr"`
	Build   string   `xml:"build,attr"`
}

type PrimaryLocation struct {
	XMLName xml.Name `xml:"location"`
	HRef    string   `xml:"href,attr"`
}

type PrimaryFormat struct {
	XMLName   xml.Name `xml:"format"`
	License   string   `xml:"license"`
	BuildHost string   `xml:"buildhost"`
}

func parsePrimary(primaryxml []byte) map[string]*PrimaryPackage {
	// Unmarshal XML content into PrimaryPackage struct
	var primary PrimaryMetadata
	err := xml.Unmarshal(primaryxml, &primary)
	checkError(err)

	var packages = make(map[string]*PrimaryPackage)
	for _, pkg := range primary.Package {
		if (pkg.Name == "aws-neuron-dkms" || pkg.Name == "aws-neuronx-dkms") && pkg.Arch == "noarch" {
			packages[pkg.Version.Version] = &pkg

		}
	}

	return packages
}

func ProcessPrimary(primaryxml []byte) map[string]*PrimaryPackage {
	primary := parsePrimary(primaryxml)

	return primary
}
