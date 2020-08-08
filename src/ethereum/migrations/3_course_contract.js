var CredentialSum = artifacts.require("CredentialSum");
var Course = artifacts.require("CourseMock");

module.exports = async function (deployer, network, accounts) {
    const [issuer1, issuer2] = accounts;
    deployer.link(CredentialSum, Course);
    return await deployer.deploy(Course, [issuer1, issuer2], 2);
};