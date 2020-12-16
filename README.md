Forked from [TriggerMesh](https://github.com/triggermesh/aws-event-sources)

## Installation

### Kubernetes

Using Helm:

```bash
$ helm repo add triggermesh https://storage.googleapis.com/triggermesh-charts
$ helm install triggermesh/aws-event-sources
```

Refer to the [aws-event-sources chart documentation](chart/README.md) for all available configuration options.

## Getting Started

The following table lists the AWS services currently supported by TriggerMesh Sources for AWS and their support level.

|                            AWS Service                            |                  Documentation                   | Support Level |
|-------------------------------------------------------------------|--------------------------------------------------|---------------|
| [CodeCommit](https://aws.amazon.com/codecommit/)                  | [README](cmd/awscodecommitsource/README.md)      | alpha         |
| [Cognito Identity Pool](https://aws.amazon.com/cognito/)               | [README](cmd/awscognitoidentitysource/README.md) | alpha         |
| [Cognito User Pool](https://aws.amazon.com/cognito/)               | [README](cmd/awscognitouserpoolsource/README.md) | alpha         |
| [DynamoDB](https://aws.amazon.com/dynamodb/)                      | [README](cmd/awsdynamodbsource/README.md)        | alpha         |
| [Kinesis](https://aws.amazon.com/kinesis/)                        | [README](cmd/awskinesissource/README.md)         | alpha         |
| [Simple Notifications Service (SNS)](https://aws.amazon.com/sns/) | [README](cmd/awssnssource/README.md)             | alpha         |
| [Simple Queue Service (SQS)](https://aws.amazon.com/sqs/)         | [README](cmd/awssqssource/README.md)             | alpha         |

For detailed usage instructions about a particular source, please refer to the linked `README.md` files.

## Roadmap

* Add a more customization properties
* Add a more generic SNS source using an operator architecture
* Add a CloudWatch source using an operator architecture
* Performance improvements

## Support

The sources listed in this repository are fully open source and can be used in any Knative cluster. They consist of event consumers for various AWS services. Most of them are packaged as `Container Sources` and make use of [CloudEvents](https://cloudevents.io/).

## Code of Conduct

This repository is not a part of [CNCF](https://www.cncf.io/) but we abide by its [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
