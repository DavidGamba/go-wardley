package wardley

#size: {
	width:     uint | *1280
	height:    uint | *768
	margin:    uint | *40
	font_size: uint | *12
}

#node: {
	id:           string & =~"^[a-z][a-z0-9_-]*$"
	label:        string | *id
	visibility:   uint
	evolution:    "genesis" | "custom" | "product" | "commodity"
	x:            uint
	description?: string
	fill:         string | *"white"
	color:        string | *"black"
}

#connector: {
	id:     string & =~"^[a-z][a-z0-9_-]*$"
	from:   #node.id
	to:     #node.id
	label?: string
	color:  string | *"black"
	type:   *"normal" | "bold" | "change" | "change-inertia"
}

#Schema: {
	size: #size
	node: [ID=_]: #node & {id: ID}
	connector: [ID=_]: #connector & {id: ID}
}
