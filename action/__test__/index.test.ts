import * as os from 'os';
import * as fs from 'fs';
import * as path from 'path';
import * as exec from '@actions/exec';
import * as io from '@actions/io';
import * as child_process from 'child_process';
import * as index from '../src/index';

const sep = path.sep;

// extension of executable files
const binExt = os.platform() === 'win32' ? '.exe' : '';

process.env.GITHUB_REPOSITORY = 'shogo82148/actions-aws-assume-role';
process.env.GITHUB_WORKFLOW = 'test';
process.env.GITHUB_RUN_ID = '1234567890';
process.env.GITHUB_ACTOR = 'shogo82148';
process.env.GITHUB_SHA = 'e3a45c6c16c1464826b36a598ff39e6cc98c4da4';
process.env.GITHUB_REF = 'ref/heads/main';

describe('tests', () => {
  let tmpdir = '';
  let subprocess: child_process.ChildProcess;
  beforeAll(async () => {
    tmpdir = await mkdtemp();
    const bin = `${tmpdir}${sep}dummy${binExt}`;

    console.log("compiling dummy server");
    let myOutput = '';
    let myError = '';
    try {
      await exec.exec('go', ['build', '-o', bin, './cmd/dummy'], {
        cwd: '../provider/assume-role',
        listeners: {
          stdout: (data: Buffer) => {
            myOutput += data.toString();
          },
          stderr: (data: Buffer) => {
            myError += data.toString();
          }
        }
      });
    } catch (e) {
      console.log(`error: ${e}`);
      console.log(`stdout: ${myOutput}`);
      console.log(`stderr: ${myError}`);
      throw e;
    }

    console.log("starting dummy server");
    subprocess = child_process.spawn(bin, [], {
      detached: true,
      stdio: 'ignore'
    });
    await sleep(1); // wait for starting process
  }, 60000);

  afterAll(async () => {
    console.log("killing dummy server");
    subprocess?.kill('SIGTERM');
    await sleep(1); // wait for stopping process
    await io.rmRF(tmpdir);
  });

  it('succeed', async () => {
    await index.assumeRole({
      githubToken: 'ghs_dummyGitHubToken',
      awsRegion: 'us-east-1',
      roleToAssume: 'arn:aws:iam::123456789012:role/assume-role-test',
      roleDurationSeconds: 900,
      roleSessionName: 'GitHubActions',
      roleSessionTagging: true,
      providerEndpoint: 'http://localhost:8080'
    });
    expect(process.env.AWS_ACCESS_KEY_ID).toBe('AKIAIOSFODNN7EXAMPLE');
    expect(process.env.AWS_SECRET_ACCESS_KEY).toBe('wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY');
    expect(process.env.AWS_SESSION_TOKEN).toBe('session-token');
    expect(process.env.AWS_DEFAULT_REGION).toBe('us-east-1');
    expect(process.env.AWS_REGION).toBe('us-east-1');
  });

  it('invalid GitHub Token', async () => {
    await expect(async () => {
      await index.assumeRole({
        githubToken: 'ghp_dummyPersonalGitHubToken',
        awsRegion: 'us-east-1',
        roleToAssume: 'arn:aws:iam::123456789012:role/assume-role-test',
        roleDurationSeconds: 900,
        roleSessionName: 'GitHubActions',
        roleSessionTagging: true,
        providerEndpoint: 'http://localhost:8080'
      });
    }).rejects.toThrow();
  });
});

function mkdtemp(): Promise<string> {
  const tmp = os.tmpdir();
  return new Promise(function (resolve, reject) {
    fs.mkdtemp(`${tmp}${sep}actions-aws-assume-role-`, (err, dir) => {
      if (err) {
        reject(err);
        return;
      }
      resolve(dir);
    });
  });
}

function sleep(waitSec: number): Promise<void> {
  return new Promise<void>(function (resolve) {
    setTimeout(() => resolve(), waitSec * 1000);
  });
}
