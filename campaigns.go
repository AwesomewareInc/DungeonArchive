package main

// File for functions dealing with listing or giving info about campaign (resources/*.json) files.

import (
	"strings"
	"io/fs"
	"path/filepath"
)

type CampaignInfo struct {
	Name string
	RawName string
} 

func ListCampaigns() ([]CampaignInfo) {
	// First, get all the json files in the resources directory.
	var campaigns_raw []string
	err := filepath.Walk("./resources/", func(path string, info fs.FileInfo, err error) error {
		if(err != nil) {
			return nil // ignore any generic errors we get reading files; we only care about filenames for the files we can see.
		}
		name := info.Name()
		if name[len(name)-5:len(name)] == ".json" {
			campaigns_raw = append(campaigns_raw, name)
		}
		return err
	})
	if(err != nil) {
		return []CampaignInfo{
			CampaignInfo{err.Error(),""},
		}
	}

	// Then go through all the names we just and pretty them up
	var campaigns []CampaignInfo
	for _, v := range campaigns_raw {
		// remove the file extension
		name := v[:len(v)-5]
		// remove underscores and dashes
		name = strings.Replace(name, "_"," ",9)
		name = strings.Replace(name, "-"," ",9)
		// capitalize the first letter of every word
		namesplit := strings.Split(name," ")
		var namefinal string
		for _, v := range namesplit {
			namefinal += strings.ToUpper(v[:1])
			namefinal += v[1:]+" "
		}
		campaigns = append(campaigns, CampaignInfo{
			Name: namefinal,
			RawName: v,
		})
	}
	return campaigns
}