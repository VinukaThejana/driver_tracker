#!/bin/bash

if [ -f .env ]; then
	export $(grep -v '^#' .env | xargs)
fi

rm -rf .kafka.yml

envsubst <.template.yml >.kafka.yml
