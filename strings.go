package main

// File for functions relating to strings; editing them, returning them, etc.

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
)

var HTMLEscapeStrings = strings.NewReplacer(
	"<", "&lt;",
	">", "&gt;",
)

// Function for getting the similarites between two strings
func SimilaritiesBetweenStrings(a, b []string) float32 {
	var found float32
	for i, v := range a {
		if i < len(b) {
			if b[i] == v {
				found++
			}
		} else {
			break
		}
	}
	return found / float32(len(a))
}

// Function for stripping the name of a character from the beginning of a sentence.
func StripName(message Message) (result string, noSpace bool) {
	content := strings.ToLower(message.Content)
	content = strings.Replace(content, "*", "", 99)
	author := strings.ToLower(message.Author)
	authorSlice := strings.Split(author, " ")
	contentSlice := strings.Split(content, " ")
	noSpace = false

	// if they are the same width then it can be concluded that the character is just saying
	// their name so don't do anything.
	// if the message is shorter then the authors name then there's nothing to strip anyways
	if len(authorSlice) >= len(contentSlice) {
		return content, noSpace
	}

	// test 1: the message starts with the author's last name
	lastName := authorSlice[len(authorSlice)-1]
	if content[0:len(lastName)] == lastName {
		newContent := content[len(lastName):]
		if newContent[0] == 's' ||
			(newContent[0] == '\'' && newContent[1] == 's') {
			noSpace = true
		}
		return content[len(lastName):], noSpace
	}

	// test 2: the author's name or something close to it at the start of the message

	contentSlice = contentSlice[0:len(authorSlice)]
	similarities := SimilaritiesBetweenStrings(authorSlice, contentSlice)
	if similarities < 0.50 {
		return content, noSpace
	}

	return content[len(author):], noSpace
}

// Function for formatting the unix timestamp into a time object
func DateFormatted(date_ string) time.Time {
	date, err := strconv.ParseInt(date_, 10, 64)
	if err != nil {
		return time.Now() // too bad. templates can't handle errors.
	}
	unix := time.Unix(date, 0)
	return unix
}

// Function for getting the combined day/month/year from a formatted time;
// useful (and hence here in this file) for seperating messages based on what day
// they were posted
func CombinedDate(date_ string) int {
	date := DateFormatted(date_)
	return int(date.Month()) + date.Day() + date.Year()
}

// Return the date as as string.
func DateString(date_ string) string {
	date := DateFormatted(date_)
	return fmt.Sprintf("%v", date)
}

// Escape any potential HTML tags in a string.
func HTMLEscape(value string) string {
	return string(template.HTML(value))
}

// Capitalize a string
func Capitalize(value string) string {
	// Treat dashes as spaces
	value = strings.Replace(value, "-", " ", 99)
	valuesplit := strings.Split(value, " ")
	var result string
	for _, v := range valuesplit {
		result += strings.ToUpper(v[:1])
		result += v[1:] + " "
	}
	return result
}

func GetFileCategory(str string) string {
	if strings.HasSuffix(str, ".webm") || strings.HasSuffix(str, ".mov") || strings.HasSuffix(str, ".mp4") {
		return "video"
	}
	if strings.HasSuffix(str, ".ogg") || strings.HasSuffix(str, ".mp3") || strings.HasSuffix(str, ".wav") {
		return "audio"
	}
	return "image"
}
