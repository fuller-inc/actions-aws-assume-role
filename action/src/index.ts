import * as core from '@actions/core';
import * as http from '@actions/http-client';

interface AssumeRoleParams {
  githubToken: string;
  awsRegion: string;
  roleToAssume: string;
  roleDurationSeconds: number;
  roleSessionName: string;
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
  const result = await client.postJson('https://uw4qs7ndjj.execute-api.us-east-1.amazonaws.com/assume-role', payload);
  if (result.statusCode !== 200) {
    throw new Error('unexpected status code');
  }
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
    await assumeRole({
      githubToken,
      awsRegion,
      roleToAssume,
      roleDurationSeconds,
      roleSessionName
    });
  } catch (error) {
    core.setFailed(error.message);
  }
}

run();
