#!/bin/bash
# Sets up a bash prompt showing the current git branch, and terraform workspace
# Place in ~/.bash_profile, or as required

no_colour=$(tput sgr0)
green=$(tput setaf 40)
yellow=$(tput setaf 3)
purple=$(tput setaf 69)
parse_terraform_env() {
  if [ -d ./.terraform ]; then
    if test -f ./.terraform/environment; then
      echo " <$(cat ./.terraform/environment)>"
    else
      echo " <>"
    fi
  fi
}

parse_git_branch() {
  branch=$(git branch --show-current 2>/dev/null)
  if [[ $branch == "" ]]; then
    echo ""
  else
    echo " ($branch)"
  fi
}

export PS1="\u:\001$green\002\w\001$yellow\002\$(parse_git_branch)\001$purple\002\$(parse_terraform_env)\001$no_colour\002 $ "
