module github.com/fuller-inc/actions-aws-assume-role/provider/assume-role

go 1.18

require (
	github.com/aws/aws-lambda-go v1.28.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.16.1
	github.com/aws/aws-sdk-go-v2/config v1.15.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.2
	github.com/aws/smithy-go v1.11.2
	github.com/golang-jwt/jwt/v4 v4.3.0
	github.com/shogo82148/aws-xray-yasdk-go v1.4.3
	github.com/shogo82148/aws-xray-yasdk-go/xrayaws-v2 v1.0.4
	github.com/shogo82148/ridgenative v1.1.2
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.10.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.0 // indirect
)
