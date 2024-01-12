import * as os from "os";
import * as fs from "fs";
import * as path from "path";
import * as exec from "@actions/exec";
import * as io from "@actions/io";
import * as core from "@actions/core";
import * as child_process from "child_process";
import * as index from "../src/index";
import { jest, describe, expect, beforeAll, afterAll, it } from "@jest/globals";

const sep = path.sep;

jest.mock("@actions/core");

// extension of executable files
const binExt = os.platform() === "win32" ? ".exe" : "";

process.env.GITHUB_REPOSITORY = "fuller-inc/actions-aws-assume-role";
process.env.GITHUB_WORKFLOW = "test";
process.env.GITHUB_RUN_ID = "1234567890";
process.env.GITHUB_ACTOR = "fuller-inc";
process.env.GITHUB_SHA = "e3a45c6c16c1464826b36a598ff39e6cc98c4da4";
process.env.GITHUB_REF = "ref/heads/main";

// set dummy id token endpoint
process.env.ACTIONS_ID_TOKEN_REQUEST_TOKEN = "dummy";
process.env.ACTIONS_ID_TOKEN_REQUEST_URL = "https://example.com";

describe("tests", () => {
  let tmpdir = "";
  let subprocess: child_process.ChildProcess;
  beforeAll(async () => {
    tmpdir = await mkdtemp();
    const bin = `${tmpdir}${sep}dummy${binExt}`;

    console.log("compiling dummy server");
    await exec.exec(
      "go",
      ["build", "-o", bin, "github.com/fuller-inc/actions-aws-assume-role/provider/assume-role/cmd/dummy"],
      {
        cwd: `..${sep}provider${sep}assume-role`,
      },
    );

    console.log("starting dummy server");
    subprocess = child_process.spawn(bin, [], {
      detached: true,
      stdio: "ignore",
    });
    await sleep(1); // wait for starting process
  }, 5 * 60000);

  afterAll(async () => {
    console.log("killing dummy server");
    subprocess?.kill("SIGTERM");
    await sleep(1); // wait for stopping process
    await io.rmRF(tmpdir);
  });

  it("succeed", async () => {
    const getIDToken = core.getIDToken as jest.Mock<typeof core.getIDToken>;
    getIDToken.mockResolvedValueOnce("dummyGitHubIDToken");

    await index.assumeRole({
      githubToken: "ghs_dummyGitHubToken",
      awsRegion: "us-east-1",
      roleToAssume: "arn:aws:iam::123456789012:role/assume-role-test",
      roleDurationSeconds: 900,
      roleSessionName: "GitHubActions",
      roleSessionTagging: true,
      providerEndpoint: "http://localhost:8080",
      useNodeId: false,
    });

    const exportVariable = core.exportVariable as jest.Mock;
    expect(exportVariable).toHaveBeenCalledWith("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE");
    expect(exportVariable).toHaveBeenCalledWith("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY");
    expect(exportVariable).toHaveBeenCalledWith("AWS_SESSION_TOKEN", "session-token");
    expect(exportVariable).toHaveBeenCalledWith("AWS_DEFAULT_REGION", "us-east-1");
    expect(exportVariable).toHaveBeenCalledWith("AWS_REGION", "us-east-1");

    const setSecret = core.setSecret as jest.Mock;
    expect(setSecret).toHaveBeenCalledWith("AKIAIOSFODNN7EXAMPLE");
    expect(setSecret).toHaveBeenCalledWith("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY");
    expect(setSecret).toHaveBeenCalledWith("session-token");
  });

  it("invalid GitHub ID Token", async () => {
    await expect(async () => {
      const getIDToken = core.getIDToken as jest.Mock<typeof core.getIDToken>;
      getIDToken.mockResolvedValueOnce("invalidGitHubIDToken");

      await index.assumeRole({
        githubToken: "ghp_dummyPersonalGitHubToken",
        awsRegion: "us-east-1",
        roleToAssume: "arn:aws:iam::123456789012:role/assume-role-test",
        roleDurationSeconds: 900,
        roleSessionName: "GitHubActions",
        roleSessionTagging: true,
        providerEndpoint: "http://localhost:8080",
        useNodeId: false,
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
