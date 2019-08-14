# wavefront-cf-notifier

If you're using AWS services, changes are you're using AWS CloudFormation to deploy new apps into production. If you're also using Wavefront to monitor your AWS resources, you can use this Lambda app to automatically create events in Wavefront so you can overlay those events on your dashboards.

```bash
.
├── LICENSE                      <-- Because everything needs a license
├── Makefile                     <-- Make to automate build
├── README.md                    <-- This file
├── go.mod                       <-- Go modules file
├── go.sum                       <-- Go sum file
├── main.go                      <-- Lambda function code
├── main_test.go                 <-- Unit tests
├── template.yaml                <-- SAM template
└── test
    └── snsevent.json            <-- Sample event used to test the function locally
```

## Dependencies

To use this app, you'll need to have a few things ready:

* [AWS account](https://aws.amazon.com/free/?all-free-tier.sort-by=item.additionalFields.SortRank&all-free-tier.sort-order=asc) to deploy the app to
* [Wavefront Access Token](https://docs.wavefront.com/wavefront_api.html) so you can configure the `template.yaml` properly
* [Docker](https://www.docker.com/products/docker-desktop) in case you want to test the app locally

## Building the app

### Getting the Go modules

To get the Go modules needed to build the app, run

```bash
# Use a proxy for repeatable builds, this example uses GoCenter
export GOPROXY=https://gocenter.io
make deps
```

or, if you don't want to use `make`

```bash
export GOPROXY=https://gocenter.io
go get ./...
```

### Building an executable

AWS Lambda takes a compiled executable to deploy, so you can run

```bash
make build
```

or, if you don't want to use `make`

```bash
GOOS=linux GOARCH=amd64 go build -o ./dist/wavefront-cf-notifier
```

## Deploying to AWS Lambda

Before you can deploy the app, you'll need to update the `template.yaml` file in three places:

* [Line 26](https://github.com/retgits/wavefront-cf-notifier/blob/master/template.yaml#L26) should be updated with the ARN of your SNS topic
* [Line 33](https://github.com/retgits/wavefront-cf-notifier/blob/master/template.yaml#L33) should be updated to your correct API endpoint for Wavefront
* [Line 34](https://github.com/retgits/wavefront-cf-notifier/blob/master/template.yaml#L34) should be updated with your Wavefront API token

Once those are configured, you can run

```bash
# Package your Lambda app and upload to S3
make package
# Create a Cloudformation Stack and deploy your SAM resources
make deploy
```

or, if you don't want to use `make`

```bash
# Package your Lambda app and upload to S3
sam package \
    --output-template-file packaged.yaml \
    --s3-bucket <name of your S3 bucket>

# Create a Cloudformation Stack and deploy your SAM resources
sam deploy \
    --template-file packaged.yaml \
    --stack-name wavefront-cf-notifier \
    --capabilities CAPABILITY_IAM
```

## All Make targets

```bash
$ make

Usage: make [TARGET]

Makefile targets
build                          Build the executable
clean                          Removes all generated code
deploy                         Create a Cloudformation Stack and deploy your SAM resources
deps                           Get the dependencies for this project.
help                           Displays the help for each target (this message).
package                        Package your Lambda app and upload to S3
samtest                        Runs SAM local
test                           Runs go test -cover
```

## License

This Lambda app is provided under the [MIT license](./LICENSE).
