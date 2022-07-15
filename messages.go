package main

// Functions relating to message objects, and sorting them.

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Regexp strings for matching certain typing behaviors

var Regexps map[string]*regexp.Regexp

type Area struct {
	Name     	string
	Messages 	[]*Message
}

type Author struct {
	Name 		string
	Messages 	[]*Message
}

type Message struct {
	Author    	string `json:"author"`
	Content   	string `json:"content"`
	Timestamp 	string `json:"timestamp"`
	Fictional 	string `json:"fictional"`
	Area      	string `json:"channel"`
	ID 			string `json:"id"`

	PreceededBy *Message
	FollowedBy 	*Message
}

// Template friendly functions for getting PreceededBy and FollowedBy
// Something is fucked up, so the names are actually reversed. I should
// look into this later.
func (message *Message) Last() (*Message) {
	if(message.FollowedBy == nil) {
		return &Message{}
	} else {
		return message.FollowedBy
	}
}
func (message *Message) Next() (*Message) {
	if(message.PreceededBy == nil) {
		return &Message{}
	} else {
		return message.PreceededBy
	}
}

func init() {
	InitRegexes()
	InitCampaigns()
}

func InitRegexes() {
	Regexps = make(map[string]*regexp.Regexp)
	Regexps["italics"] = regexp.MustCompile(`^(\*)([^\*]*)(\*)$`)
	Regexps["bold"] = regexp.MustCompile(`^(\*){2}([^\*]*)(\*){2}$`)
	Regexps["lettersonly"] = regexp.MustCompile(`[^A-z]*`)
	// experimental, may be removed: regexp for determining if an author was a narrator
	Regexps["narrator"] = regexp.MustCompile(`^((\?)*)$`)
}

// initialize the campaigns that we'll be showing.
func InitCampaigns() {
	Campaigns = make(map[string]*Campaign)
	// First, get all the json files in the resources directory.
	var campaign_files []string
	_ = filepath.Walk("./resources/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			// ignore any generic errors we get reading files;
			// we only care about filenames for the files we can see.
			return nil
		}
		name := info.Name()
		if name[len(name)-5:] == ".json" {
			campaign_files = append(campaign_files, "./resources/"+name)
		}
		return nil
	})
	// Then, for all of the files we just got...
	for _, v := range campaign_files {
		// Make a new campaign
		newCampaign := &Campaign{}
		name := v[12 : len(v)-5]
		Campaigns[name] = newCampaign
		newCampaign.Name = name

		newCampaign.Areas = make(map[string]*Area)
		newCampaign.Authors = make(map[string]*Author)

		// Read the file.
		file, err := os.ReadFile(v)
		if err != nil {
			newCampaign.Valid = false
			newCampaign.Error = err.Error()
			fmt.Println("Error reading " + v + "; " + err.Error())
		}
		// Split the file.
		lines := strings.Split(string(file), "\n")
		// And unmarshal each line into a new message object.
		messageObjectFromLine := func(line string) (*Message) {
			message := &Message{}
			json.Unmarshal([]byte(line), message)
			// Some messages have blank area names, even though the json file doesn't have them?
			// And then the areas end up having no messages anyways so...?
			// yeah just ignore them lol
			if message.Area == "" {
				return nil
			}
			return message
		}

		registerMessageObject := func(message *Message) {
			// Check the area tag of the new message to see if
			// the corresponding area exists, and if not create it.
			var area *Area
			if newCampaign.Areas[message.Area] == nil {
				area = &Area{}
				area.Name = message.Area
				newCampaign.Areas[message.Area] = area
			} else {
				area = newCampaign.Areas[message.Area]
			}
			area.Messages = append(area.Messages, message)
			// Check the author tag of the new message to see if
			// the corresponding author exists, and if not create it.
			authorName := strings.ToLower(Regexps["lettersonly"].ReplaceAllString(message.Author,""))
			var author *Author
			if newCampaign.Authors[authorName] == nil {
				author = &Author{}
				author.Name = authorName
				newCampaign.Authors[authorName] = author
			} else {
				author = newCampaign.Authors[authorName]
			}
			author.Messages = append(author.Messages, message)
		}

		var lastMessage *Message

		for i := range lines {
			// get the current message
			message := messageObjectFromLine(lines[i])
			// if we're not at the beginning of the lines
			if(i > 0) {
				// and last message isn't already set
				if(lastMessage == nil) {
					// set it to the last line.
					lastMessage = messageObjectFromLine(lines[i-1])
					// ...as long as that last line isn't a null message.
					if(lastMessage != nil) {
						// and it's not from another channel.
						if(lastMessage.Area == message.Area) {
							lastMessage.FollowedBy = message
						}
					}
				} else {
					// Otherwise, link what we currently have
					lastMessage.FollowedBy = message
				}
			}
			// and if this message isn't null
			if(message != nil) {
				// register it to the authors/areas thing
				registerMessageObject(message)
				message.PreceededBy = lastMessage
				// set it as the previous message
				lastMessage = message
			}
		}
	}
}

// Shorthand for sanitizing a string based on Regexps["lettersonly"]
func Sanitize(value string) (string) {
	return strings.ToLower(Regexps["lettersonly"].ReplaceAllString(value,""))
}

// Function for listing the areas in a campaign
func ListAreas(value string) []string {
	var areas []string
	if Campaigns[value] == nil {
		return []string{}
	}
	campaign := Campaigns[value].Areas
	for _, v := range campaign {
		areas = append(areas, v.Name)
	}
	return areas
}

// Function for listing messages in an area.
func ListMessages(value string, area string) []Message {
	if Campaigns[value] == nil {
		// No messages is an indication of an error and is handled accordingly.
		return []Message{}
	}

	// Area "all" is a keyword for every area in a campaign.
	// If we're given it...
	if area == "all" {
		// First, just get every message in the campaign.
		var messages []Message
		for _, a := range Campaigns[value].Areas {
			for _, m := range a.Messages {
				messages = append(messages, *m)
			}
		}
		// But then, we want to sort it by the time posted.
		sort.Slice(messages, func(a, b int) bool {
			return DateFormatted(messages[a].Timestamp).Before(DateFormatted(messages[b].Timestamp))
		})
		return messages
	}

	if Campaigns[value].Areas[area] == nil {
		return []Message{}
	}
	messages := Campaigns[value].Areas[area].Messages
	messagelen := len(messages)
	// We actually want to create a new array for them
	// so that we can return them in a reverse order.
	messagesnew := make([]Message, messagelen)
	for i := messagelen - 1; i > 0; i-- {
		messagesnew = append(messagesnew, *messages[i])
	}
	return messagesnew
}

// Function for getting a "message type" that determines how it looks.
func GetMessageType(message Message) string {
	content := message.Content
	author := message.Author
	if Regexps["narrator"].Match([]byte(author)) {
		return "narration"
	}
	if Regexps["italics"].Match([]byte(content)) {
		return "action"
	}
	if Regexps["bold"].Match([]byte(content)) {
		return "loud"
	}
	return "normal"
}

// Function for editing a message marked as an action; it takes the author's name out if it can and
// uncapitalizes the first letter
func ParseActionMessage(message Message) string {
	content, noSpace := StripName(message)
	content = HTMLEscape(content)
	contentSlice := strings.Split(content, " ")

	lowercase := strings.ToLower(contentSlice[0])
	contentNew := lowercase + " "
	for _, v := range contentSlice[1:] {
		contentNew += v + " "
	}
	if !noSpace {
		contentNew = "­ " + contentNew
	}
	return contentNew
}

// Function for matching author objects with a name in them.
func MatchNames(campaign, name string) ([]string) {
	var matches []string
	for _, a := range Campaigns[campaign].Authors {
		namelen := len(name)
		if(len(a.Name) < namelen) {continue}
		if(	a.Name[:namelen] == name || // starts with
			a.Name[len(a.Name)-namelen:] == name || // ends with
			a.Name == name) { // equals 
				matches = append(matches,a.Name)
		}
	}
	return matches
}

// Function for searching through a campaign's messages for character interactions
func SearchMessages(campaign, area string, query_ []string) []*Message {
	query := strings.Split(query_[0],",")
	var authors []string
	for _, v := range query {
		authors = append(authors, Sanitize(v))
	}
	// Get the author objects that match the first author name
	firstAuthorMatches := MatchNames(campaign, authors[0])

	// Then, get all messages from the people found.
	var allmessages []*Message
	for _, a := range firstAuthorMatches {
		messages := Campaigns[campaign].Authors[a].Messages
		for _, v := range messages {
			allmessages = append(allmessages, v)
		}
	}
	// And sort it by the time posted.
	sort.Slice(allmessages, func(a, b int) bool {
		return DateFormatted(allmessages[a].Timestamp).Before(DateFormatted(allmessages[b].Timestamp))
	})

	// If only one author is specified, we can just return those.
	if(len(query) <= 1) {
		return allmessages
	// Otherwise, we'll have to search each message a little more throughly, and specifically, search the messages after it, to see if they're
	// relevant to the search
	} else {
		// for now, only two characters are supported.
		nextCharacterMatches := MatchNames(campaign, authors[1])
		var allmessages_ []*Message
		// for each of the messages we got earlier
		for _, v := range allmessages {
			// check if the message following it is from any of the characters we're looking for.
			for _, a := range nextCharacterMatches {
				author := Sanitize(v.Next().Author)
				if(author == a) {
					allmessages_ = append(allmessages_,v)
					allmessages_ = append(allmessages_,v.Next())
				}
			}
		}
		return allmessages_
	}
}


// Template-ready function for checking if a name is in an query array.
func NameInSearch(value_ string, query []string) (bool) {
	// The query is actually a weird array that has what we want as a fake array in the first value
	names_ := strings.Split(query[0],",")
	// We want to sanitize what we get before comparing them.
	var names []string
	for _, v := range names_ {
		names = append(names, Sanitize(v))
	}
	// Same for the other value we're given.
	value := Sanitize(value_)
	// Now, compare them.
	for _, v := range names {
		if(v == value) {
			return true
		}
	}
	return false
}