module github.com/fuller-inc/actions-aws-assume-role/provider/assume-role

go 1.20

require (
	github.com/aws/aws-sdk-go-v2 v1.21.0
	github.com/aws/aws-sdk-go-v2/config v1.18.39
	github.com/aws/aws-sdk-go-v2/service/sts v1.21.5
	github.com/aws/smithy-go v1.14.2
	github.com/shogo82148/aws-xray-yasdk-go v1.6.0
	github.com/shogo82148/aws-xray-yasdk-go/xrayaws-v2 v1.1.2
	github.com/shogo82148/ctxlog v0.1.0
	github.com/shogo82148/goat v0.0.6
	github.com/shogo82148/ridgenative v1.4.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.13.37 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.41 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.42 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.13.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.15.6 // indirect
	github.com/shogo82148/memoize v0.0.2 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
