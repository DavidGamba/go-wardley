package wardley

import "strings"

#Size: {
	width:     uint | *1280
	height:    uint | *768
	margin:    uint | *40
	font_size: uint | *12
}

#Node: {
	id:           string & =~"^[a-zA-Z][a-zA-Z0-9_-]*$"
	label:        string | *strings.ToTitle(strings.Replace(id, "_", " ", -1))
	visibility:   uint
	evolution:    "genesis" | "custom" | "product" | "commodity"
	x:            uint
	description?: string
	fill:         string | *"white"
	color:        string | *"black"
}

#Connector: {
	id:     string & =~"^[a-zA-Z][a-zA-Z0-9_-]*$"
	from:   #Node.id
	to:     #Node.id
	label?: string
	color:  string | *"black"
	type:   *"normal" | "bold" | "change" | "change-inertia"
}

#Schema: {
	size: #Size
	node: [ID=_]:      #Node & {id:      ID}
	connector: [ID=_]: #Connector & {id: ID}
}
