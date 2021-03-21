import * as core from '@actions/core';
import * as http from '@actions/http-client';

interface AssumeRoleParams {
  githubToken: string;
  awsRegion: string;
  roleToAssume: string;
  roleDurationSeconds: number;
  roleSessionName: string;
  providerEndpoint: string;
}

interface AssumeRoleResult {
  access_key_id: string;
  secret_access_key: string;
  session_token: string;
}

interface AssumeRoleError {
  message: string;
}

async function assumeRole(params: AssumeRoleParams) {
  const payload = {
    github_token: params.githubToken,
    role_to_assume: params.roleToAssume,
    role_session_name: params.roleSessionName,
    repository: process.env['GITHUB_REPOSITORY'],
    sha: process.env['GITHUB_SHA']
  };
  const client = new http.HttpClient('actions-aws-assume-role');
  const result = await client.postJson<AssumeRoleResult | AssumeRoleError>(params.providerEndpoint, payload);
  if (result.statusCode !== 200) {
    const resp = result.result as AssumeRoleError;
    core.setFailed(resp.message);
    return;
  }
  const resp = result.result as AssumeRoleResult;

  core.setSecret(resp.access_key_id);
  core.exportVariable('AWS_ACCESS_KEY_ID', resp.access_key_id);

  core.setSecret(resp.secret_access_key);
  core.exportVariable('AWS_SECRET_ACCESS_KEY', resp.secret_access_key);

  core.setSecret(resp.session_token);
  core.exportVariable('AWS_SESSION_TOKEN', resp.session_token);

  core.exportVariable('AWS_DEFAULT_REGION', params.awsRegion);
  core.exportVariable('AWS_REGION', params.awsRegion);
}

async function run() {
  try {
    const required = {
      required: true
    };
    const githubToken = core.getInput('github-token', required);
    const awsRegion = core.getInput('aws-region', required);
    const roleToAssume = core.getInput('role-to-assume', required);
    const roleDurationSeconds = Number.parseInt(core.getInput('role-duration-seconds', required));
    const roleSessionName = core.getInput('role-session-name', required);
    const providerEndpoint =
      core.getInput('provider-endpoint') || 'https://uw4qs7ndjj.execute-api.us-east-1.amazonaws.com/assume-role';
    await assumeRole({
      githubToken,
      awsRegion,
      roleToAssume,
      roleDurationSeconds,
      roleSessionName,
      providerEndpoint
    });
  } catch (error) {
    core.setFailed(error.message);
  }
}

run();
