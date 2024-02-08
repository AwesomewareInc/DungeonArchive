package main

// The FuncMap that main.go uses, taken out of there to keep the file clean a bit
// (this will inevitably get pretty big)

import (
	"math"
	"strconv"
	"text/template"
)

var FuncMap = template.FuncMap{
	"ConfigValue":        ConfigValue,
	"ListCampaigns":      ListCampaigns,
	"PrettyString":       PrettyString,
	"ListAreas":          ListAreas,
	"StringNoExtension":  StringNoExtension,
	"ListMessages":       ListMessages,
	"GetMessageType":     GetMessageType,
	"CombinedDate":       CombinedDate,
	"DateString":         DateString,
	"ParseMarkdown":      ParseMarkdown,
	"ParseActionMessage": ParseActionMessage,
	"HTMLEscape":         HTMLEscape,
	"SearchMessages":     SearchMessages,
	"NameInSearch":       NameInSearch,
	"PrettyPrintValues":  PrettyPrintValues,

	"StrToInt": func(st string) int {
		in, err := strconv.Atoi(st)
		if err != nil {
			// we have no choice lmao
			panic(err)
		}
		return in
	},

	"GetArea": GetArea,

	// "inc" stands for "incredible" because
	// what the fuck why can't i just do arithmetic in templates
	"Inc": func(i int) int {
		return i + 1
	},

	"Dec": func(i int) int {
		return i - 1
	},

	"Sub": func(a, b int) int {
		return int(math.Abs(float64(a - b)))
	},

	"notnil": func(pointer *Message) bool {
		return (pointer != nil)
	},
}
