#!/usr/bin/env bash

if  oc api-versions | grep -q 'devconsole.openshift.io'; then
   echo -e "\n\033[0;32m \xE2\x9C\x94 Devconsole Operator is already installed \033[0m\n"
else
   echo -e "Running Openshift Version 4.x \n"
   echo -e "\n Installing DevConsole Operator... \n"
   echo -e "\n Installing Catalog Source... \n"
   oc apply -f ./yamls/catalog_source_OS4.yaml
   echo -e "\n Waiting for catalog source to get installed before creating subscription \n"
   sleep 60s
   echo -e "\n Creating Subscription... \n"
   oc apply -f ./yamls/subscription_OS4.yaml
   sleep 5s
fi
