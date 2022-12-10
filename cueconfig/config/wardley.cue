package wardley

map: #Schema & {}

map: size: {
	font_size: 3
}

map: node: n1: {
	visibility: 1
	x:          1
	evolution:  "custom"
}

map: node: n2: {
	label:      "hola"
	visibility: 1
	x:          1
	evolution:  "product"
}

map: node: n3_is_coOL: {
	visibility: map.node.n2.visibility + 1
	x:          map.node.n2.x + 1
	evolution:  "product"
}

map: connector: n1n2: {
	from:  map.node.n1.id
	to:    map.node.n2.id
	label: "en -> es"
}

map: connector: n2n3: {
	from: map.node.n2.id
	to:   map.node.n3_is_coOL.id
}
