"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.assumeRole = void 0;
const core = __importStar(require("@actions/core"));
const http = __importStar(require("@actions/http-client"));
function assertIsDefined(val) {
    if (val === undefined || val === null) {
        throw new Error(`Missing required environment value. Are you running in GitHub Actions?`);
    }
}
async function assumeRole(params) {
    const { GITHUB_REPOSITORY, GITHUB_WORKFLOW, GITHUB_RUN_ID, GITHUB_ACTOR, GITHUB_SHA, GITHUB_REF } = process.env;
    assertIsDefined(GITHUB_REPOSITORY);
    assertIsDefined(GITHUB_WORKFLOW);
    assertIsDefined(GITHUB_RUN_ID);
    assertIsDefined(GITHUB_ACTOR);
    assertIsDefined(GITHUB_SHA);
    const payload = {
        github_token: params.githubToken,
        role_to_assume: params.roleToAssume,
        role_session_name: params.roleSessionName,
        duration_seconds: params.roleDurationSeconds,
        repository: GITHUB_REPOSITORY,
        sha: GITHUB_SHA,
        role_session_tagging: params.roleSessionTagging,
        run_id: GITHUB_RUN_ID,
        workflow: GITHUB_WORKFLOW,
        actor: GITHUB_ACTOR,
        branch: GITHUB_REF || ''
    };
    const client = new http.HttpClient('actions-aws-assume-role');
    const result = await client.postJson(params.providerEndpoint, payload);
    if (result.statusCode !== 200) {
        const resp = result.result;
        core.setFailed(resp.message);
        return;
    }
    const resp = result.result;
    if (resp.message) {
        core.info(resp.message);
    }
    if (resp.warning) {
        core.warning(resp.warning);
    }
    core.setSecret(resp.access_key_id);
    core.exportVariable('AWS_ACCESS_KEY_ID', resp.access_key_id);
    core.setSecret(resp.secret_access_key);
    core.exportVariable('AWS_SECRET_ACCESS_KEY', resp.secret_access_key);
    core.setSecret(resp.session_token);
    core.exportVariable('AWS_SESSION_TOKEN', resp.session_token);
    core.exportVariable('AWS_DEFAULT_REGION', params.awsRegion);
    core.exportVariable('AWS_REGION', params.awsRegion);
}
exports.assumeRole = assumeRole;
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
        const roleSessionTagging = parseBoolean(core.getInput('role-session-tagging', required));
        const providerEndpoint = core.getInput('provider-endpoint') || 'https://uw4qs7ndjj.execute-api.us-east-1.amazonaws.com/assume-role';
        await assumeRole({
            githubToken,
            awsRegion,
            roleToAssume,
            roleDurationSeconds,
            roleSessionName,
            roleSessionTagging,
            providerEndpoint
        });
    }
    catch (error) {
        core.setFailed(error.message);
    }
}
function parseBoolean(s) {
    // YAML 1.0 compatible boolean values
    switch (s) {
        case 'y':
        case 'Y':
        case 'yes':
        case 'Yes':
        case 'YES':
        case 'true':
        case 'True':
        case 'TRUE':
            return true;
        case 'n':
        case 'N':
        case 'no':
        case 'No':
        case 'NO':
        case 'false':
        case 'False':
        case 'FALSE':
            return false;
    }
    throw `invalid boolean value: ${s}`;
}
if (require.main === module) {
    run();
}
