package wardley

map: #Schema & {
	size: {
		font_size: 3
	}

	node: n1: {
		visibility: 1
		x:          1
		evolution:  "custom"
	}

	node: n2: {
		label:      "hola"
		visibility: 1
		x:          1
		evolution:  "product"
	}

	node: n3_is_coOL: {
		visibility: node.n2.visibility + 1
		x:          node.n2.x + 1
		evolution:  "product"
	}

	connector: n1n2: {
		from:  node.n1.id
		to:    node.n2.id
		label: "en -> es"
	}
}
