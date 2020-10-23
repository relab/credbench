var NotaryLib = artifacts.require("Notary");
var CredentialSumLib = artifacts.require("CredentialSum");
var Course = artifacts.require("CourseMock");

module.exports = async function (deployer, network, accounts) {
    const [teacher, evaluator] = accounts;

    console.log(`--- Deploying course at ${network} network ---`);
    await deployer.link(NotaryLib, Course);
    await deployer.link(CredentialSumLib, Course);
    return await deployer.deploy(Course, [teacher, evaluator], 2);
};