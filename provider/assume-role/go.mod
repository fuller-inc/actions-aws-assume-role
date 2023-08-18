module github.com/fuller-inc/actions-aws-assume-role/provider/assume-role

go 1.20

require (
	github.com/aws/aws-sdk-go-v2 v1.20.2
	github.com/aws/aws-sdk-go-v2/config v1.18.33
	github.com/aws/aws-sdk-go-v2/service/sts v1.21.3
	github.com/aws/smithy-go v1.14.1
	github.com/shogo82148/aws-xray-yasdk-go v1.5.2
	github.com/shogo82148/aws-xray-yasdk-go/xrayaws-v2 v1.1.2
	github.com/shogo82148/ctxlog v0.1.0
	github.com/shogo82148/goat v0.0.6
	github.com/shogo82148/ridgenative v1.4.0
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.13.32 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.39 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.39 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.33 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.13.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.15.2 // indirect
	github.com/shogo82148/memoize v0.0.2 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
