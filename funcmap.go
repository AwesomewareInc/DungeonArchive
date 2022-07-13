package main

// The FuncMap that main.go uses, taken out of there to keep the file clean a bit
// (this will inevitably get pretty big)

import (
	"text/template"
	"math"
)

var FuncMap = 	template.FuncMap{
					"GetQuery": GetQuery,
					"ConfigValue": ConfigValue,
					"ListCampaigns": ListCampaigns,
					"PrettyString": PrettyString,
					"ListAreas": ListAreas,
					"StringNoExtension": StringNoExtension,
					"ListMessages":	ListMessages,
					"GetMessageType": GetMessageType,
					"CombinedDate": CombinedDate,
					"DateString": DateString,
					"ParseMarkdown": ParseMarkdown,
					"ParseActionMessage": ParseActionMessage,
					"HTMLEscape": HTMLEscape,

					// "inc" stands for "incredible" because
					// what the fuck why can't i just do arithmetic in templates
					"Inc": func(i int) (int) {
						return i + 1
					},

					"Sub": func(a, b int) (int) {
						return int(math.Abs(float64(a - b)))
					},
				}