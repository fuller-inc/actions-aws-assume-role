module github.com/fuller-inc/actions-aws-assume-role/provider/assume-role

go 1.21.6
toolchain go1.22.5

require (
	github.com/aws/aws-sdk-go-v2 v1.36.2
	github.com/aws/aws-sdk-go-v2/config v1.29.7
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.15
	github.com/aws/smithy-go v1.22.2
	github.com/shogo82148/aws-xray-yasdk-go v1.8.1
	github.com/shogo82148/aws-xray-yasdk-go/xrayaws-v2 v1.1.10
	github.com/shogo82148/goat v0.1.0
	github.com/shogo82148/ridgenative v1.5.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.17.60 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.15 // indirect
	github.com/shogo82148/forwarded-header v0.1.0 // indirect
	github.com/shogo82148/memoize v0.0.4 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)
