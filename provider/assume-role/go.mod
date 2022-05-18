module github.com/fuller-inc/actions-aws-assume-role/provider/assume-role

go 1.18

require (
	github.com/aws/aws-lambda-go v1.31.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.16.4
	github.com/aws/aws-sdk-go-v2/config v1.15.5
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.4
	github.com/aws/smithy-go v1.11.2
	github.com/golang-jwt/jwt/v4 v4.4.1
	github.com/shogo82148/aws-xray-yasdk-go v1.5.1
	github.com/shogo82148/aws-xray-yasdk-go/xrayaws-v2 v1.1.1
	github.com/shogo82148/ridgenative v1.2.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.12.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.4 // indirect
)
