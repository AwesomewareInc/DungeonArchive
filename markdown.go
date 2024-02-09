package main

// Passthrough function for markdown parsing to template shit.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var BasicReplacer = strings.NewReplacer(
	"<p>", "",
	"</p>", "",
	"<ul>", "",
	"</ul>", "",
	"<li>", "*",
	"</li>", "",
	"<hr>", " ",
	"\\n", "<br>",
)

func ParseMarkdown(value string, md string) string {

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))
	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	result := string(markdown.Render(doc, renderer))
	result = BasicReplacer.Replace(result)

	// Emojis
	emotes := Regexps["emote"].FindAllStringSubmatch(result, -1)

	for _, emote_parts := range emotes {
		result = strings.Replace(result, emote_parts[0], "<img class='emoji' src='https://cdn.discordapp.com/emojis/"+emote_parts[len(emote_parts)-1]+".webp?size=40'>", 1)
	}

	channels := Regexps["channel"].FindAllStringSubmatch(result, -1)

	for _, ch_parts := range channels {
		ch := ch_parts[1]
		ch_int, err := strconv.Atoi(ch)
		if err != nil {
			panic("j" + err.Error())
		}
		ar := GetArea(value, ch_int)
		if ar != nil {
			className := "channel-name"
			if !ar.Archived {
				className += " archived"
			}
			result = strings.Replace(result, ch_parts[0], "<a target='a_blank' href='/campaign/ddd_roleplay/messages/"+ch+"' class='"+className+"'>#"+ar.Name+"</a>", 1)
		} else {
			fmt.Println("Warning! Area is nil!")
		}
	}

	users := Regexps["user"].FindAllStringSubmatch(result, -1)

	for _, user_parts := range users {
		ch := user_parts[2]
		us_int, err := strconv.Atoi(ch)
		if err != nil {
			panic("j" + err.Error())
		}
		result = strings.Replace(result, user_parts[0], "<span class='user-id'>@"+GetUser(value, us_int).Name+"</span>", 1)
	}

	messages := Regexps["message"].FindAllStringSubmatch(result, -1)

	for _, user_parts := range messages {
		result = strings.Replace(result, user_parts[0], "/campaign/"+value+"/messages/"+user_parts[2]+"#"+user_parts[3], 1)
	}
	return result
}
