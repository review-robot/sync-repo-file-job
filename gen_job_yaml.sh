#!/bin/bash

set -euo pipefail
pn=$#

log(){
    local s=$1
    echo -e "\n${s}\n"
}

gen(){
  local platform=$1
  local fileNames=$2
  local enpoint=$3
  local orgRepos=$4
  local image=$5

  orgRepos=${orgRepos//\//\\\/}

  cp -f job_tpl.yaml sync-repo-file-job.yaml

  sed -i -e "s/{PLATFORM}/${platform}/" ./sync-repo-file-job.yaml
  sed -i -e "s/{FILE_NAMES}/${fileNames}/" ./sync-repo-file-job.yaml
  sed -i -e "s/{ENDPOINT}/${enpoint}/" ./sync-repo-file-job.yaml
  sed -i -e "s/{ORG_REPOS}/${orgRepos}/" ./sync-repo-file-job.yaml
  sed -i -e "s/{IMAGE}/${image}/" ./sync-repo-file-job.yaml
  log "info: sync-repo-file-job.yaml has been successfully generated"
}

cmd_help(){
 local me="gen_job_yaml.sh"
cat << EOF
Usage: $me param1 param2 param3 param4 param5
For Example: $me gitee OWNERS,OWNER 192.168.1.123:9090 openEuler/core,openLookeng/hetu-core sync-repo-file-job:latest
The command above will
Generate a k8s job yaml file named sync-repo-file-job.yaml.
param1: The platform of the repository to which the files to be synced belong: gitee or github
param2: File names to be synchronized, multiple file names separated by , .
param3: RPC call address provided by sync-file-server.
param4: The full path of the repository to which the files to be synchronized belong,multiple file names separated by , .
param5: The container image of the job.
EOF
}

if [ $pn -lt 4 ]; then
    cmd_help
    exit 1
fi

gen $1 $2 $3 $4 $5
