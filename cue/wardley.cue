package wardley

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

c1c2: #connector & {
	from: n1.id
	to: n2.id
	label: "en -> es"
}

n3: #node & {
	id: "n3"
	visibility: n2.visibility + 1
	x: n2.x + 1
	evolution: "product"
}

#bodySchema & {
size: s
nodes: [n1, n2, n3]
connectors: [c1c2]
}
