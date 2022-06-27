package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"subscriptions/src/config"
)

var S3Client *s3.Client
var AthenaClient *athena.Client

func SetupAWS() {
	if config.GetConfig().AwsConfig.ManuallySpecify {
		setupWithManuallyProvidedConfig()
		return
	}

	setupWithDefaults()
}

func setupWithManuallyProvidedConfig() {
	creds := credentials.NewStaticCredentialsProvider(
		*config.GetConfig().AwsConfig.AccessKeyId,
		*config.GetConfig().AwsConfig.AccessKeySecret,
		"",
	)

	var cfg aws.Config
	var athenaCfg aws.Config
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
			EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				//Despite being deprecated, it seems this is actually used rather than the above sometimes - so don't delete
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           *config.GetConfig().AwsConfig.Endpoint,
					SigningRegion: region,
				}, nil
			}),
		}

		if config.GetConfig().AwsConfig.AthenaEndpoint == nil {
			athenaCfg = cfg
		} else {
			athenaCfg = aws.Config{
				Credentials: creds,
				Region:      config.GetConfig().AwsConfig.Region,
				EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           *config.GetConfig().AwsConfig.AthenaEndpoint,
						SigningRegion: region,
					}, nil
				}),
				EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
					//Despite being deprecated, it seems this is actually used rather than the above sometimes - so don't delete
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           *config.GetConfig().AwsConfig.AthenaEndpoint,
						SigningRegion: region,
					}, nil
				}),
			}
		}
	} else {
		cfg = aws.Config{
			Credentials: creds,
			Region:      config.GetConfig().AwsConfig.Region,
		}
	}

	S3Client = s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = true
	})
	AthenaClient = athena.NewFromConfig(athenaCfg)
}

func setupWithDefaults() {
	configuration, err := cfg.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	S3Client = s3.NewFromConfig(configuration, func(o *s3.Options) {
		o.Region = config.GetConfig().AwsConfig.Region
	})
	AthenaClient = athena.NewFromConfig(configuration, func(o *athena.Options) {
		o.Region = config.GetConfig().AwsConfig.Region
	})
}
