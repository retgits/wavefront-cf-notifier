# -----------------------------------------------------------------------------
# Description: Makefile
# Author(s): retgits
# 
# This software may be modified and distributed under the terms of the
# MIT license. See the LICENSE file for details.
# -----------------------------------------------------------------------------

# AWSBUCKET is the Amazon S3 bucket where AWS CloudFormation will store the binary
AWSBUCKET ?= no_bucket_set

# Suppress checking files and output
.PHONY: help deps clean build test
.SILENT: help deps clean build test

# Targets
help: ## Displays the help for each target (this message).
	echo
	echo Usage: make [TARGET]
	echo
	echo Makefile targets
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps: ## Get the dependencies for this project.
	go get ./...

clean: ## Removes all generated code
	rm -rf ./dist/wavefront-cf-notifier
	rm -rf ./packaged.yaml
	
build: ## Build the executable
	GOOS=linux GOARCH=amd64 go build -o ./dist/wavefront-cf-notifier

test: ## Runs go test -cover
	go test -cover

samtest: ## Runs SAM local
	sam local invoke -e ./test/snsevent.json

package: clean build ## Package your Lambda app and upload to S3
	sam package \
		--output-template-file packaged.yaml \
		--s3-bucket $(AWSBUCKET)

deploy: ## Create a Cloudformation Stack and deploy your SAM resources
	sam deploy \
		--template-file packaged.yaml \
		--stack-name wavefront-cf-notifier \
		--capabilities CAPABILITY_IAM