package wardley

map: #Schema & {}

#Schema: {
	size: #Size
	node: [ID=_]:      #Node & {id:      ID}
	connector: [ID=_]: #Connector & {id: ID}
}
