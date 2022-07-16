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
	// the bare minimum for a character's name; useful for the search page
	Regexps["sanitization"] = regexp.MustCompile(`[^A-z0-9\-]*`)
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
			authorName := Sanitize(message.Author)
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
				// register it to the characters/areas thing
				registerMessageObject(message)
				message.PreceededBy = lastMessage
				// set it as the previous message
				lastMessage = message
			}
		}
	}
}

// Shorthand for sanitizing a string based on Regexps["sanitization"]
func Sanitize(value string) (string) {
	return strings.ToLower(Regexps["sanitization"].ReplaceAllString(value,""))
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
func SearchMessages(campaign string, query_ []string) []*Message {
	// Split the query up based on characters
	query := strings.Split(query_[0],",")

	// Split THOSE queries up based on actions
	var values []string
	var lastAction string
	for _, v := range query {
		part := strings.Split(v,"::")
		values = append(values,part[0])
		if(len(part) <= 1) {
			// If the last action before termination was an and-or, we want to do a special
			// termination that is described later
			if(lastAction == "and-or") {
				values = append(values,"add")
			} else {
				values = append(values,"terminate")
			}
		} else {
			values = append(values,part[1])
			lastAction = part[1]
		}
		
	}

	// Get all the characters to look for 
	var characters []string
	for i := 0; i < len(values); i+=2 {
		characters = append(characters, Sanitize(values[i]))
	}

	var allmessages []*Message

	// Get the messages from the first character and add them (...if the second argument isn't "and-or")
	// (because and-or will handle adding the messages for us)
	if(len(values) >= 1) {
		if(values[1] != "and-or") {
			allmessages = GetMessagesFrom(campaign, values[0])
		}
	}

	// Then, go through the query and filter the messages list more and more based on the search values.
	for i := 0; i < len(values); i+=2 {
		// Name the values we want from the values array.
		character := Sanitize(values[i])
		action := Sanitize(values[i+1])
		var nextCharacter string
		if(i+2 < len(values)) {
			nextCharacter = Sanitize(values[i+2])
		} else {
			nextCharacter = ""
		}
		var nextAction string
		if(i+3 < len(values)) {
			nextAction = Sanitize(values[i+3])
		} else {
			nextAction = "terminate"
		}
		// yes this is really how deep we have to go, since and/or sometimes takes the character
		// to a function that wants nextCharacter
		var nextNextCharacter string
		if(i+4 < len(values)) {
			nextNextCharacter = Sanitize(values[i+4])
		} else {
			nextNextCharacter = ""
		}

		// Finally, filter those messages based on the values we got earlier
		allmessages = FilterMessages(allmessages,campaign,character,action,nextCharacter,nextAction,nextNextCharacter)
	}

	// Sort all those by the time posted.
	sort.Slice(allmessages, func(a, b int) bool {
		return DateFormatted(allmessages[a].Timestamp).Before(DateFormatted(allmessages[b].Timestamp))
	})

	return allmessages
}

// Function for searching the messages from a specific character
func GetMessagesFrom(campaign, character string) ([]*Message) {
	var filtered []*Message
	// Get the authors that could possibly match the character
	matches := MatchNames(campaign, Sanitize(character))

	// And get the messages from them.
	for _, a := range matches {
		messages := Campaigns[campaign].Authors[a].Messages
		for _, v := range messages {
			filtered = append(filtered, v)
		}
	}
	return filtered
}


// Function for above for filtering messages based on a certain criteria
func FilterMessages(messages []*Message, campaign string, character, action, nextCharacter, nextAction, nextNextCharacter string) ([]*Message) {
	fmt.Printf("given %v messages\n",len(messages))
	var filtered []*Message
	fmt.Printf("section %v::%v\n",character,action)

	switch(action) {
		// get messages from A followed by B
		case "interacting-with":
			nextCharacterMatches := MatchNames(campaign, nextCharacter)
			// for each of the messages we got
			for _, v := range messages {
				nextAuthor := Sanitize(v.Next().Author)
				// check if the message following it is from any of the characters we're looking for.
				for _, a := range nextCharacterMatches {
					// if the message is followed by narrator, walk through those messages to see if its eventually leads to a match
					filtered_, add := addMessageIfFromNarrator(v.Next(),*(new([]*Message)),a)
					// If it does...
					if(add) {
						filtered = append(filtered, v)
						for _, v := range filtered_ {
							filtered = append(filtered,v)
						}
					}
					if(nextAuthor == a) {
						if(!add) {
							filtered = append(filtered,v)
							filtered = append(filtered,v.Next())
						}
						filtered = addLastMessageIfFromSame(v,filtered)
					}
				}
			}
		// get messages from A doing the next action and get messages of B doing the next action (or just existing).
		case "and-or":
			// If the next command is another and-or, just add the current character's messages
			if(nextAction == "and-or") {
				nextAction = "add"
			}
			filtered = FilterMessages(messages, campaign, character, nextAction, nextNextCharacter, "", "")
		// get messages from A with a certain keyword.
		case "mentioning":
			fmt.Println(nextCharacter)
			// for each of the messages we got
			for _, v := range messages {
				// check if the message has the keyword
				// (nextCharacter in this case would be the search string)
				if(strings.Contains(Sanitize(v.Content),Sanitize(nextCharacter))) {
					filtered = append(filtered, v)
				}
			}
		// narrow the search results to a specific area.
		case "in":
			// for each of the messages we got
			for _, v := range messages {
				if(Sanitize(v.Area) == nextCharacter) {
					filtered = append(filtered,v)
				}
			}
		// "and" is basically terminate but we add the last character's messages to the queue
		case "add":
			// get the messages from the last character
			filtered_ := GetMessagesFrom(campaign, character)
			// get a copy of the messages
			messages_ := messages
			// add the last character's messages to the mesages clone
			for _, v := range filtered_ {
				messages_ = append(messages_,v)
			}
			// return the messages clone
			filtered = messages_
		case "terminate":
			return messages
	}
	return filtered
}
// recursive function for above to add the last message if it's from the same author
func addLastMessageIfFromSame(message *Message, messages []*Message) (newmessages []*Message) {
	newmessages = messages
	prevAuthor := Sanitize(message.Last().Author)
	if(prevAuthor == Sanitize(message.Author)) {
		newmessages = append(newmessages,message.Last())
		newmessages = addLastMessageIfFromSame(message.Last(),newmessages)
	}
	return newmessages
}
// recursive function for walking through a block of narrator messages to see if it eventually leads to another match.
func addMessageIfFromNarrator(message *Message, messages []*Message, nextCharacter string) (newmessages []*Message, eventual bool) {
	newmessages = messages
	if(message.Fictional == "False") {
		newmessages = append(newmessages,message)
		newmessages, eventual = addMessageIfFromNarrator(message.Next(),newmessages,nextCharacter)
	} else {
		if(Sanitize(message.Author) == nextCharacter) {
			eventual = true
			newmessages = append(newmessages,message)
		}
	}
	return newmessages, eventual
}


// Template-ready function for checking if a name is in an query array.
func NameInSearch(value_ string, query []string) (bool) {
	// The query is actually a weird array that has what we want as a fake array in the first value
	names_ := strings.Split(query[0],",")
	// We want to sanitize what we get before comparing them.
	var names []string
	for _, v := range names_ {
		name := strings.Split(v,"::")
		names = append(names, Sanitize(name[0]))
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

// Function for pretty-printing the query array
func PrettyPrintValues(query []string) (result string) {
	characters := strings.Split(query[0],",")
	for _, v := range characters {
		parts := strings.Split(v,"::")
		result += Capitalize(parts[0])
		if(len(parts) > 1) {
			switch(parts[1]) {
				case "interacting-with":	result += " interacting with "
				case "and-or":				result += " and/or "
				case "in": 					result += " in "
				case "mentioning": 			result += " mentioning "
			}
		}
	}
	return
}