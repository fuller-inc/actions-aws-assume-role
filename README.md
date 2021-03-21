# Configure AWS Credentials by Assuming Roles

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
- name: Configure AWS Credentials
  uses: shogo82148/actions-aws-assume-role@v1
  with:
    aws-region: us-east-2
    role-to-assume: arn:aws:iam::123456789012:role/GitHubRepoRole-us-east-2
```

## Session tagging

You can enable session tagging by adding `role-session-tagging: true`.

```yaml
- uses: shogo82148/actions-aws-assume-role@v1
  with:
    aws-region: us-east-2
    role-to-assume: arn:aws:iam::123456789012:role/GitHubRepoRole-us-east-2
    role-session-tagging: true
```

The session will have the name "GitHubActions" and be tagged with the following tags:

| Key        | Value               |
| ---------- | ------------------- |
| GitHub     | "Actions"           |
| Repository | `GITHUB_REPOSITORY` |
| Workflow   | `GITHUB_WORKFLOW`   |
| RunId      | `GITHUB_RUN_ID`     |
| Actor      | `GITHUB_ACTOR`      |
| Branch     | `GITHUB_REF`        |
| Commit     | `GITHUB_SHA`        |

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

## How to Work

![How to Work](how-to-work.svg)
