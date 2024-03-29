# Configure AWS Credentials by Assuming Roles

The action configures AWS Credential by assuming roles and [OpenID Connect (OIDC)](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect).

## Usage

At first, create an IAM role for your repository.
The role's trust policy must allow an AWS account `053160724612` to assume the role and check external id:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::053160724612:root"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "your-name/your-repo"
        }
      }
    }
  ]
}
```

And then, add the following step to your workflow:

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      id-token: write # needed to interact with GitHub's OIDC Token endpoint.
      contents: read

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Configure AWS Credentials
        uses: fuller-inc/actions-aws-assume-role@v1
        with:
          aws-region: us-east-2
          role-to-assume: arn:aws:iam::123456789012:role/GitHubRepoRole-us-east-2
```

### Session tagging

You can enable session tagging by adding `role-session-tagging: true`.

```yaml
- uses: fuller-inc/actions-aws-assume-role@v1
  with:
    aws-region: us-east-2
    role-to-assume: arn:aws:iam::123456789012:role/GitHubRepoRole-us-east-2
    role-session-tagging: true
```

The session will have the name "GitHubActions" and be tagged with the following tags:

| Key         | Value                         |
| ----------- | ----------------------------- |
| GitHub      | "Actions"                     |
| Repository  | `GITHUB_REPOSITORY`           |
| Workflow    | `GITHUB_WORKFLOW`             |
| RunId       | `GITHUB_RUN_ID`               |
| Actor       | `GITHUB_ACTOR`                |
| Branch      | `GITHUB_REF`                  |
| Commit      | `GITHUB_SHA`                  |
| Audience    | `aud` of the ID token         |
| Subject     | `sub` of the ID token         |
| Environment | `environment` of the ID token |

_Note: all tag values must conform to [the requirements](https://docs.aws.amazon.com/STS/latest/APIReference/API_Tag.html). Particularly, `GITHUB_WORKFLOW` will be truncated if it's too long. If `GITHUB_ACTOR` or `GITHUB_WORKFLOW` contain invalid characters, the characters will be replaced with an '\_'._

The role's trust policy need extra permission `sts:TagSession` for session tagging:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::053160724612:root"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "your-name/your-repo"
        }
      }
    },
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::053160724612:root"
      },
      "Action": "sts:TagSession"
    }
  ]
}
```

### Use the node id of the repository

You can use the global node id of the repository instead of its name as ExternalId.
By adding `use-node-id: true`, the action sends the node id (e.g. `R_kgDOFMsDjw`) as ExternalId.

```yaml
- uses: fuller-inc/actions-aws-assume-role@v1
  with:
    aws-region: us-east-2
    role-to-assume: arn:aws:iam::123456789012:role/GitHubRepoRole-us-east-2
    use-node-id: true
```

To get the node id, call [Get a repository REST API](https://docs.github.com/en/rest/reference/repos#get-a-repository).

```console
# with curl command
curl -i -u username:token -H "X-Github-Next-Global-ID: 1" "https://api.github.com/repos/{owner}/{repo}"

# with GitHub CLI
gh api "repos/{owner}/{repo}" -H "X-Github-Next-Global-ID: 1" --jq ".node_id"
```

You'll get a response that includes the `node_id` of the repository:

```json
{
  "id": 348849039,
  "node_id": "R_kgDOFMsDjw",
  "name": "actions-aws-assume-role",
  "full_name": "fuller-inc/actions-aws-assume-role",
  "private": false,
  "owner": {
    "login": "(... snip ...)"
  }
}
```

In this example, the `node_id` value is `R_kgDOFMsDjw`.
The role's trust policy looks like this:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::053160724612:root"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "R_kgDOFMsDjw"
        }
      }
    }
  ]
}
```

For more information about global node IDs, see [Using global node IDs](https://docs.github.com/en/graphql/guides/using-global-node-ids).

### Migrate your node id to the next format

If you get the following warning, you need to migrate your node IDs to the next format.

> It looks that you use legacy node IDs. You need to migrate them.
> See https://github.com/fuller-inc/actions-aws-assume-role#migrate-your-node-id-to-the-next-format for more detail.

Because GitHub is now [rolling out the new GraphQL global ID format](https://github.blog/2021-11-16-graphql-global-id-migration-update/#how-do-i-migrate-my-service).

To get the next IDs, call [Get a repository REST API](https://docs.github.com/en/rest/reference/repos#get-a-repository) with `X-Github-Next-Global-ID: 1` header.
If you use `curl` or `gh` (GitHub CLI), add the `-H "X-Github-Next-Global-ID: 1"` option.

```console
# with curl command
curl -i -u username:token -H "X-Github-Next-Global-ID: 1" "https://api.github.com/repos/{owner}/{repo}"

# with GitHub CLI
gh api "repos/{owner}/{repo}" -H "X-Github-Next-Global-ID: 1" --jq ".node_id"
```

Your role's trust policy looks like this:

```diff
 {
   "Version": "2012-10-17",
   "Statement": [
     {
       "Effect": "Allow",
       "Principal": {
         "AWS": "arn:aws:iam::053160724612:root"
       },
       "Action": "sts:AssumeRole",
       "Condition": {
         "StringEquals": {
-          "sts:ExternalId": "MDEwOlJlcG9zaXRvcnkzNDg4NDkwMzk="
+          "sts:ExternalId": "R_kgDOFMsDjw"
         }
       }
     }
   ]
 }
```

See [Migrating GraphQL global node IDs](https://docs.github.com/en/graphql/guides/migrating-graphql-global-node-ids)
and [GraphQL global ID migration update](https://github.blog/2021-11-16-graphql-global-id-migration-update/)
for more detail.

## How to Work

![How to Work](how-to-work.svg)

1. Request a new credential\
   The `fuller-inc/actions-aws-assume-role` action sends an ID token of OpenID connect to the credential provider. The credential provider works on AWS Lambda owned by @fuller-inc.
2. Check Permission of GitHub Repository\
   The Lambda function validates the ID token.
3. Request AssumeRole to an IAM Role on your AWS account
4. Check Permission of the IAM Role\
   The AWS IAM Service checks the role's trust policy.
5. Return the Credential
6. Configure the Credential to the workflow

## Caution

- You can use the credential provider for free, but note that it works on my personal AWS Account.
- Your AWS Account ID, the name of your IAM Role, and the name of your GitHub Repository will be logged by AWS CloudTrail on my AWS Account.
- If you enable tagging session, `GITHUB_WORKFLOW`, `GITHUB_RUN_ID`, `GITHUB_ACTOR`, `GITHUB_REF`, and `GITHUB_SHA` are also logged.
- If you want to use this action on your private repository, I recommend building your own credential provider. You can find its source code on [the provider directory](https://github.com/fuller-inc/actions-aws-assume-role/tree/main/provider)

## License

The scripts and documentation in this project are released under the [MIT License](LICENSE).
