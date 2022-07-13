package main

// Functions relating to message objects, and sorting them.

import (
	"encoding/json"
	"path/filepath"
	"io/fs"
	"strconv"
	"os"
	"fmt"
	"strings"
	"regexp"
	"time"
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
			// Check the area tag of the new message to see if
			// the corresponding area exists, and if not create it.
			var area *Area
			if(newCampaign.Areas[message.Area] == nil) {
				area = &Area{}
				area.Name = message.Area
				if(message.Area == "") {
					area.Name = "?"
				}
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

// Function for stripping a certain regex from a message
func StripRegex(message Message, regex_ string) (string) {
	content := message.Content
	author := message.Author
	regex := strings.Split(regex_, " ")
	for _, v := range regex {
		switch(v) {
			case "italics":
				content = Regexps["italics"].ReplaceAllString(content,"$2")
			case "bold":
				content = Regexps["bold"].ReplaceAllString(content,"$2")
			case "name":
				// this isn't a regex but we might as well stick it here;
				// if a sentence starts with the author's name or starts with a part of 
				// the author's name, remove it.
				author_slice := strings.Split(strings.ToLower(author)," ")
				content_slice := strings.Split(strings.ToLower(content)," ")
				var content_ string
				for i := 0; i < len(author_slice); i+=2 {
					firstName := author_slice[i]
					var lastName string
					if(i < len(author_slice)-1) {
						lastName = author_slice[i+1]
					}
					firstNameLength := len(firstName)
					lastNameLength := len(lastName)
					if(content_slice[0] == firstName && content_slice[1] == lastName) {
						content_ = content[firstNameLength+lastNameLength+2:len(content)]
						break
					}
					if(content_slice[0] == firstName || content_slice[0] == lastName) {
						content_ = content[firstNameLength+1:len(content)]
						break
					}
				}
				if(content_ != "") {content = content_}
		}
	}
	return content

}

// Function for formatting the unix timestamp into a time object
func DateFormatted(date_ string) (time.Time) {
	date, err := strconv.ParseInt(date_, 10, 64)
	if(err != nil) {
		return time.Now() // too bad. templates can't handle errors.
	}
	unix := time.Unix(date,0)
	return unix
}

// Function for getting the combined day/month/year from a formatted time; 
// useful (and hence here in this file) for seperating messages based on what day
// they were posted
func CombinedDate(date_ string) (int) {
	date := DateFormatted(date_)
	return int(date.Month())+date.Day()+date.Year()
}

// Return the date as as string.
func DateString(date_ string) (string) {
	date := DateFormatted(date_)
	return fmt.Sprintf("%v",date)
}
