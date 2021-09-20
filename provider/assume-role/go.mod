module github.com/fuller-inc/actions-aws-assume-role/provider/assume-role

go 1.17

require (
	github.com/aws/aws-lambda-go v1.26.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.9.1
	github.com/aws/aws-sdk-go-v2/config v1.8.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.7.1
	github.com/aws/smithy-go v1.8.0
	github.com/shogo82148/aws-xray-yasdk-go v1.3.0
	github.com/shogo82148/aws-xray-yasdk-go/xrayaws-v2 v1.0.1
	github.com/shogo82148/ridgenative v1.1.1
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.2.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.3.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.4.0 // indirect
)
