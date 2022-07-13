package main

// Functions relating to message objects, and sorting them.

import (
	"encoding/json"
	"path/filepath"
	"io/fs"
	"os"
	"fmt"
	"strings"
	"regexp"
	"sort"
)

// Regexp strings for matching certain typing behaviors

var Regexps map[string]*regexp.Regexp

type Area struct {
	Name  		string
	Messages 	[]Message
}

type Message struct {
	Author 		string  		`json:"author"`
	Content 	string 			`json:"content"`
	Timestamp 	string 			`json:"timestamp"`
	Fictional 	string 			`json:"fictional"`
	Area  		string 			`json:"channel"`
}

func init() {
	InitRegexes()
	InitCampaigns()
}

func InitRegexes() {
	Regexps = make(map[string]*regexp.Regexp)
	Regexps["italics"] = regexp.MustCompile(`^(\*)([^\*]*)(\*)$`)
	Regexps["bold"] = regexp.MustCompile(`^(\*){2}([^\*]*)(\*){2}$`)
	// experimental, may be removed: regexp for determining if an author was a narrator
	Regexps["narrator"] = regexp.MustCompile(`^((\?)*)$`)
}

// initialize the campaigns that we'll be showing.
func InitCampaigns() {
	Campaigns = make(map[string]*Campaign)
	// First, get all the json files in the resources directory.
	var campaign_files []string
	_ = filepath.Walk("./resources/", func(path string, info fs.FileInfo, err error) error {
		if(err != nil) {
			// ignore any generic errors we get reading files; 
			// we only care about filenames for the files we can see.
			return nil 
		}
		name := info.Name()
		if name[len(name)-5:len(name)] == ".json" {
			campaign_files = append(campaign_files, "./resources/"+name)
		}
		return nil
	})
	// Then, for all of the files we just got...
	for _, v := range campaign_files {
		// Make a new campaign
		newCampaign := &Campaign{}
		name := v[12:len(v)-5]
		Campaigns[name] = newCampaign
		newCampaign.Name = name

		newCampaign.Areas = make(map[string]*Area)

		// Read the file.
		file, err := os.ReadFile(v)
		if(err != nil) {
			newCampaign.Valid = false
			newCampaign.Error = err.Error()
			fmt.Println("Error reading "+v+"; "+err.Error())
		}
		// Split the file.
		lines := strings.Split(string(file),"\n")
		// And unmarshal each line into a new message object.
		for _, n := range lines {
			message := Message{}
			json.Unmarshal([]byte(n),&message)
			// Some messages have blank area names, even though the json file doesn't have them?
			// And then the areas end up having no messages anyways so...?
			// yeah just ignore them lol
			if(message.Area == "") {
				continue
			}
			// Check the area tag of the new message to see if
			// the corresponding area exists, and if not create it.
			var area *Area
			if(newCampaign.Areas[message.Area] == nil) {
				area = &Area{}
				area.Name = message.Area
				newCampaign.Areas[message.Area] = area
			} else {
				area = newCampaign.Areas[message.Area]
			}
			area.Messages = append(area.Messages, message)
		}
	}
}

// Function for listing the areas in a campaign
func ListAreas(value string) ([]string) {
	var areas []string
	if(Campaigns[value] == nil) {
		return []string{}
	}
	campaign := Campaigns[value].Areas
	for _, v := range campaign {
		areas = append(areas, v.Name)
	}
	return areas
}

// Function for listing messages in an area.
func ListMessages(value string, area string) ([]Message) {
	if(Campaigns[value] == nil) {
		// No messages is an indication of an error and is handled accordingly.
		return []Message{} 
	}

	// Area "all" is a keyword for every area in a campaign.
	// If we're given it...
	if(area == "all") {
		// First, just get every message in the campaign.
		var messages []Message
		for _, a := range Campaigns[value].Areas {
			for _, m := range a.Messages {
				messages = append(messages, m)
			}
		}
		// But then, we want to sort it by the time posted.
		sort.Slice(messages, func(a, b int) bool {
			return DateFormatted(messages[a].Timestamp).Before(DateFormatted(messages[b].Timestamp))
		})
		return messages
	}

	if(Campaigns[value].Areas[area] == nil) {
		return []Message{}
	}
	messages := Campaigns[value].Areas[area].Messages
	messagelen := len(messages)
	// We actually want to create a new array for them
	// so that we can return them in a reverse order.
	messagesnew := make([]Message, messagelen)
	for i := messagelen-1; i > 0; i-- {
		messagesnew = append(messagesnew, messages[i])
	}
	return messagesnew
}

// Function for getting a "message type" that determines how it looks.
func GetMessageType(message Message) (string) {
	content := message.Content
	author := message.Author
	if(Regexps["narrator"].Match([]byte(author))) {
		return "narration"
	}
	if(Regexps["italics"].Match([]byte(content))) {
		return "action"
	}
	if(Regexps["bold"].Match([]byte(content))) {
		return "loud"
	}
	return "normal"
}

// Function for editing a message marked as an action; it takes the author's name out if it can and 
// uncapitalizes the first letter
func ParseActionMessage(message Message) string {
	content, noSpace := StripName(message)
	content = HTMLEscape(content)
	contentSlice := strings.Split(content," ")

	lowercase := strings.ToLower(contentSlice[0])
	contentNew := lowercase+" "
	for _, v := range contentSlice[1:] {
		contentNew += v+" "
	}
	if(!noSpace) {
		contentNew = "­ "+contentNew
	}
	return contentNew
}

