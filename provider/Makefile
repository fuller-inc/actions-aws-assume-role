
.PHONY: build
build:
	sam build

.PHONY: deploy
deploy:
	sam deploy

.PHONY: oidc-provider
oidc-provider:
	aws cloudformation deploy \
		--region us-east-1 \
		--stack-name "oidc-provider" \
		--template-file "oidc-provider.yaml"

# prepare resources for CI/CD
.PHONY: cicd
cicd:
	aws cloudformation deploy \
		--region us-east-1 \
		--stack-name "aws-assume-role-cicd" \
		--template-file "cicd.yaml" \
		--capabilities CAPABILITY_NAMED_IAM
