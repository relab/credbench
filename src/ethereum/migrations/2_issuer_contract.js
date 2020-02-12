var Issuer = artifacts.require("IssuerMock");

module.exports = async function (deployer, network, accounts) {
    const [issuer1, issuer2] = accounts;
    return await deployer.deploy(Issuer, [issuer1, issuer2], 2);
};