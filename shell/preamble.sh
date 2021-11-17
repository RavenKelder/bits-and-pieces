#!/bin/bash

# This line enforces the following, from 
# https://www.gnu.org/software/bash/manual/html_node/The-Set-Builtin.html
#
# -e -- Exit immediately if a pipeline (see Pipelines), which may consist of a
#       single simple command (see Simple Commands), a list (see Lists), or a 
#       compound command (see Compound Commands) returns a non-zero status. 
# -u -- Treat unset variables and parameters other than the special parameters
#       ‘@’ or ‘*’ as an error when performing parameter expansion. An error 
#       message will be written to the standard error, and a non-interactive 
#       shell will exit.
# -o pipefail -- If set, the return value of a pipeline is the value of the last
#                (rightmost) command to exit with a non-zero status, or zero if
#                all commands in the pipeline exit successfully. 
set -euo pipefail

