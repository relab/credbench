var Faculty = artifacts.require("FacultyMock");

module.exports = async function (deployer, network, accounts) {
    const [issuer1, issuer2] = accounts;
    return await deployer.deploy(Faculty, [issuer1, issuer2], 2);
};