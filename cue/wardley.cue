package wardley

#size: {
	width:     uint | *1280
	height:    uint | *768
	margin:    uint | *40
	font_size: uint | *12
}

#node: {
	id: =~"^[a-z][a-z0-9_-]*$"
	label:        string | *id
	visibility:   uint
	evolution:    "genesis" | "custom" | "product" | "commodity"
	x:            uint
	description?: string
	fill:         string | *"white"
	color:        string | *"black"
}

#connector: {
	from:   #node.id
	to:     #node.id
	label?: string
	color:  string | *"black"
	type:   *"normal" | "bold" | "change" | "change-inertia"
}

s: #size & {
	font_size: 3
}

n1: #node & {
	id: "hello"
	visibility: 1
	x: 1
	evolution: "custom"
}

n2: #node & {
	id: "hola"
	label: "hola"
	visibility: 1
	x: 1
	evolution: "product"
}

n3: #node & {
	id: "n3"
	visibility: n2.visibility + 1
	x: n2.x + 1
	evolution: "product"
}
