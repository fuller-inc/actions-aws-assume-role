module github.com/fuller-inc/actions-aws-assume-role/provider/assume-role

go 1.22

toolchain go1.24.1

require (
	github.com/aws/aws-sdk-go-v2 v1.38.3
	github.com/aws/aws-sdk-go-v2/config v1.31.6
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.2
	github.com/aws/smithy-go v1.23.0
	github.com/shogo82148/aws-xray-yasdk-go v1.8.1
	github.com/shogo82148/aws-xray-yasdk-go/xrayaws-v2 v1.1.10
	github.com/shogo82148/goat v0.1.0
	github.com/shogo82148/ridgenative v1.5.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.18.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.6 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.34.2 // indirect
	github.com/shogo82148/forwarded-header v0.1.0 // indirect
	github.com/shogo82148/memoize v0.0.4 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)
