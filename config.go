package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// File for working with the config.toml; functions here don't return errors because they can be used in a template page.

var Config struct {
	Title       string
	Description string
}

func init() {
	// First, get the config file name. There could be two; config.example.toml or config.toml.
	var filename string
	if _, err := os.Open("./config.toml"); err == nil {
		filename = "./config.toml"
	} else if _, err := os.Open("./config.example.toml"); err == nil {
		filename = "./config.example.toml"
	} else {
		fmt.Println(err)
		return
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = toml.Decode(string(file), &Config)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Function for getting config values, for the template files.
func ConfigValue(name string) string {
	switch name {
	case "Title":
		return Config.Title
	case "Description":
		return Config.Description
	}
	return ""
}
