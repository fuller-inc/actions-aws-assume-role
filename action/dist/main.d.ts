interface AssumeRoleParams {
    githubToken: string;
    awsRegion: string;
    roleToAssume: string;
    roleDurationSeconds: number;
    roleSessionName: string;
    roleSessionTagging: boolean;
    providerEndpoint: string;
    useNodeId: boolean;
}
export declare function assumeRole(params: AssumeRoleParams): Promise<void>;
export declare function run(): Promise<void>;
export {};
