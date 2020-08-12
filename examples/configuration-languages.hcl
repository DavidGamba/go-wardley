size {
  width     = 1400
  height    = 700
  margin    = 20
  font_size = 10
}

#################################################
# Humans
#################################################
# anchor
node user_humans {
  label      = "Humans"
  visibility = 1
  evolution  = "custom"
  x          = 0
}

node hcl_humans {
  label      = "HCL"
  visibility = node.user_humans.visibility + 1
  evolution  = "custom"
  x          = 0
}

node yaml_humans {
  label      = "YAML"
  visibility = node.user_humans.visibility + 1
  evolution  = "product"
  x          = 0
}

node json_humans {
  label      = "JSON"
  visibility = node.user_humans.visibility + 1
  evolution  = "product"
  x          = 1
}

node xml_humans {
  label      = "XML"
  visibility = node.user_humans.visibility + 1
  evolution  = "commodity"
  x          = 1
}

# Humans
connector {
  from  = "user_humans"
  to    = "yaml_humans"
  label = "Prone to whitespace errors"
  type  = "bold"
}

connector {
  from = "user_humans"
  to   = "hcl_humans"
  type = "bold"
}

connector {
  from = "user_humans"
  to   = "json_humans"
}

connector {
  from  = "user_humans"
  to    = "xml_humans"
  color = "red"
}

#################################################
# Machines
#################################################
# anchor
node user_machines {
  label      = "Machines"
  visibility = 4
  evolution  = "custom"
  x          = 0
}

node hcl_machines {
  label      = "HCL"
  visibility = node.user_machines.visibility + 1
  evolution  = "custom"
  x          = 0
}

node yaml_machines {
  label      = "YAML"
  visibility = node.user_machines.visibility + 1
  evolution  = "product"
  x          = 0
}

node json_machines {
  label      = "JSON"
  visibility = node.user_machines.visibility + 1
  evolution  = "product"
  x          = 1
}

node xml_machines {
  label      = "XML"
  visibility = node.user_machines.visibility + 1
  evolution  = "commodity"
  x          = 1
}

# machines
connector {
  from  = "user_machines"
  to    = "yaml_machines"
  label = "Ambiguous grammar"
  color = "red"
}

connector {
  from  = "user_machines"
  to    = "hcl_machines"
  label = "allows for JSON\nintermediate representation"
  type  = "bold"
}

connector {
  from = "user_machines"
  to   = "json_machines"
  type = "bold"
}

connector {
  from = "user_machines"
  to   = "xml_machines"
  type = "bold"
}

#################################################
# APIs
#################################################
# anchor
node user_apis {
  label      = "apis"
  visibility = 7
  evolution  = "custom"
  x          = 0
}

node hcl_apis {
  label      = "HCL"
  visibility = node.user_apis.visibility + 1
  evolution  = "custom"
  x          = 0
}

node yaml_apis {
  label      = "YAML"
  visibility = node.user_apis.visibility + 1
  evolution  = "product"
  x          = 0
}

node json_apis {
  label      = "JSON"
  visibility = node.user_apis.visibility + 1
  evolution  = "product"
  x          = 1
}

node xml_apis {
  label      = "XML"
  visibility = node.user_apis.visibility + 1
  evolution  = "commodity"
  x          = 1
}

# apis
connector {
  from  = "user_apis"
  to    = "yaml_apis"
  label = "Ambiguous grammar"
  color = "red"
}

connector {
  from = "user_apis"
  to   = "hcl_apis"
  type = "bold"
}

connector {
  from  = "user_apis"
  to    = "json_apis"
  label = "No standard schema validation"
}

connector {
  from = "user_apis"
  to   = "xml_apis"
  type = "bold"
}

#################################################
# Tooling
#################################################

node tooling_all {
  label      = "Tooling for all languages"
  visibility = 10
  evolution  = "commodity"
  x          = 0
}

node tooling_single {
  label      = "Tooling for Go\nGolang"
  visibility = node.tooling_all.visibility
  evolution  = "custom"
  x          = 0
}

node yaml_tooling {
  label      = "YAML"
  visibility = node.tooling_all.visibility + 1
  evolution  = "custom"
  x          = 0
}

node hcl_tooling {
  label      = "HCL"
  visibility = node.tooling_all.visibility + 1
  evolution  = "product"
  x          = 0
}

node json_tooling {
  label      = "JSON"
  visibility = node.tooling_all.visibility + 1
  evolution  = "product"
  x          = 1
}

node xml_tooling {
  label      = "XML"
  visibility = node.tooling_all.visibility + 1
  evolution  = "commodity"
  x          = 1
}

connector {
  from = "yaml_tooling"
  to   = "tooling_all"
}

connector {
  from = "json_tooling"
  to   = "tooling_all"
}

connector {
  from = "xml_tooling"
  to   = "tooling_all"
}

connector {
  from  = "hcl_tooling"
  to    = "tooling_single"
  color = "red"
  type  = "bold"
}

# vim:ft=terraform
