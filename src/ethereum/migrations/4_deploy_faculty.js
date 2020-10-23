var NotaryLib = artifacts.require("Notary");
var CredentialSumLib = artifacts.require("CredentialSum");
var Faculty = artifacts.require("FacultyMock");

module.exports = async function (deployer, network, accounts) {
    const [dean, adm] = accounts;

    console.log(`--- Deploying faculty at ${network} network ---`);
    await deployer.link(NotaryLib, Faculty);
    await deployer.link(CredentialSumLib, Faculty);
    return await deployer.deploy(Faculty, [dean, adm], 2);
};