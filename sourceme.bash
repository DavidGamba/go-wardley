#!/bin/bash

# Remove existing entries to ensure the right one is loaded
# This is not required when the completion one liner is loaded in your bashrc.
complete -r ./go-wardley 2>/dev/null

# Requires that the go-wardley binary is in your PATH
complete -o default -C go-wardley go-wardley
