# size {
# 	width     = 1000
# 	height    = 1000
# 	margin    = 20
# 	font_size = 9
# }

# leyend {
#   node {
#			label = "node"
#			fill = "node"
#			color = "node"
#   }
#   node {
#			label = "node"
#			fill = "node"
#			color = "node"
#   }
#   connector {
#			label = "node"
#			color = "node"
#			type = "node"
#   }
# }

# anchor
node user {
  label = "User"
	fill  = "black"
	color = "black"
	visibility = 1
	evolution = "custom"
	x = 1
	description = "User Description"
}

node vcs {
  label = "On Prem VCS"
	fill  = "black"
	color = "black"
	visibility = 2
	evolution = "product"
	x = 1
	description = "On prem VCS"
}

node code_commit {
  label = "Code Commit Mirror"
	fill  = "white"
	color = "red"
	visibility = 2
	evolution = "commodity"
	x = 1
	description = "Allows Code Pipeline to access the code."
}

node deployment_script {
  label = "Deployment\nScript"
	fill  = "black"
	color = "black"
	visibility = 3
	evolution = "genesis"
	x = 1
	description = ""
}

node rest_based_deployment {
  label = "Rest based deployment\nAPI Gateway/Lambda"
	fill  = "black"
	color = "red"
	visibility = 3
	evolution = "product"
	x = 2
	description = "Utopia world, ask for an environment using the browser for example."
}

node ci_cd {
  label = "On Prem CI/CD"
	fill  = "black"
	color = "black"
	visibility = 4
	evolution = "product"
	x = 1
	description = "Product we have to maintain and customize in house."
}

node code_pipeline {
  label = "Code Pipeline"
	fill  = "white"
	color = "red"
	visibility = 4
	evolution = "commodity"
	x = 1
	description = "Built in integrations with AWS, no need for maintaining plugins or build nodes, etc."
}

node tooling {
  label = "Tooling"
	fill  = "white"
	color = "blue"
	visibility = 4
	evolution = "custom"
	x = 1
	description = "Even though ansible is a product it requires codifying the procedure of how to get what we want and doesn't track state."
}

node ansible {
  label = "Ansible"
	fill  = "black"
	color = "black"
	visibility = 5
	evolution = "genesis"
	x = 1
	description = "Even though ansible is a product it requires codifying the procedure of how to get what we want and doesn't track state."
}

node terraform_v011 {
  label = "Terraform v0.11"
	fill  = "white"
	color = "black"
	visibility = 5
	evolution = "custom"
	x = 1
	description = "External because we don't have to write how to get to what we want, only describe it."
}

node terraform_v012 {
  label = "Terraform v0.12"
	fill  = "white"
	color = "black"
	visibility = 5
	evolution = "product"
	x = 1
	description = "Many fixes to syntax and to index management."
}

connector {
	from = "user"
	to   = "deployment_script"
	label = ""
	color = "black"
	type = "normal"
}

connector {
	from = "user"
	to   = "vcs"
	label = ""
	color = "black"
	type = "normal"
}

connector {
	from = "vcs"
	to   = "code_commit"
	label = ""
	color = "red"
	type = "change-inertia"
}

connector {
	from = "vcs"
	to   = "ci_cd"
	label = ""
	color = "black"
	type = "normal"
}

connector {
	from = "code_commit"
	to   = "code_pipeline"
	label = ""
	color = "red"
	type = "normal"
}

connector {
	from = "ci_cd"
	to   = "code_pipeline"
	label = ""
	color = "red"
	type = "change-inertia"
}

connector {
	from = "deployment_script"
	to   = "rest_based_deployment"
	label = ""
	color = "red"
	type = "change-inertia"
}

connector {
	from = "tooling"
	to   = "ansible"
	label = "EC2 instance provisioning"
	color = "black"
	type = "bold"
}

connector {
	from = "tooling"
	to   = "terraform_v011"
	label = ""
	color = "black"
	type = "normal"
}

connector {
	from = "tooling"
	to   = "terraform_v012"
	label = ""
	color = "red"
	type = "normal"
}

connector {
	from = "ansible"
	to   = "terraform_v011"
	label = ""
	color = "black"
	type = "change"
}

connector {
	from = "terraform_v011"
	to   = "terraform_v012"
	label = ""
	color = "red"
	type = "change-inertia"
}

# vim:ft=terraform
