package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"go.uber.org/zap"
	"subscriptions/src/config"
	"subscriptions/src/monitoring"
)

var cfg aws.Config

func SetupAWS() {
	creds := credentials.NewStaticCredentialsProvider(
		config.GetConfig().AwsConfig.AccessKeyId,
		config.GetConfig().AwsConfig.AccessKeySecret,
		"",
	)

	if config.GetConfig().AwsConfig.Endpoint != nil {
		cfg = aws.Config{
			Credentials: creds,
			Region:      config.GetConfig().AwsConfig.Region,
			EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           *config.GetConfig().AwsConfig.Endpoint,
					SigningRegion: region,
				}, nil
			}),
		}
	} else {
		cfg = aws.Config{
			Credentials: creds,
			Region:      config.GetConfig().AwsConfig.Region,
		}
	}

	client := kinesis.NewFromConfig(cfg)
	output, err := client.ListStreams(monitoring.GlobalContext, &kinesis.ListStreamsInput{
		ExclusiveStartStreamName: nil,
		Limit:                    nil,
	})

	if err != nil {
		monitoring.GlobalContext.Error("Something went wrong listing streams", zap.Error(err))
	} else {
		monitoring.GlobalContext.Info("Streams are...", zap.Any("streams", output))
	}
}
