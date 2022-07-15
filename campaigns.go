package main

// File for functions dealing with listing or giving info about campaign (resources/*.json) files.

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type CampaignInfo struct {
	Name    string
	RawName string
}

type Campaign struct {
	Name  	string
	Areas 	map[string]*Area
	Authors map[string]*Author
	Valid 	bool
	Error 	string
}

var Campaigns map[string]*Campaign

func ListCampaigns() []CampaignInfo {
	// First, get all the json files in the resources directory.
	var campaigns_raw []string
	err := filepath.Walk("./resources/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil // ignore any generic errors we get reading files; we only care about filenames for the files we can see.
		}
		name := info.Name()
		if name[len(name)-5:] == ".json" {
			campaigns_raw = append(campaigns_raw, name)
		}
		return err
	})
	if err != nil {
		return []CampaignInfo{
			{err.Error(), ""},
		}
	}

	// Then go through all the names we just and pretty them up
	var campaigns []CampaignInfo
	for _, v := range campaigns_raw {
		campaigns = append(campaigns, CampaignInfo{
			Name:    PrettyString(v),
			RawName: v,
		})
	}
	return campaigns
}

// Function for getting a extensionless version of a file name
func StringNoExtension(value string) string {
	return strings.Split(value, ".")[0]
}

// Function for getting a pretty version of snake case formatted string
func PrettyString(value string) string {
	// remove the file extension, if any.
	name := StringNoExtension(value)
	// remove underscores and dashes
	name = strings.Replace(name, "_", " ", 9)
	name = strings.Replace(name, "-", " ", 9)
	// capitalize the first letter of every word
	return Capitalize(name)
}
