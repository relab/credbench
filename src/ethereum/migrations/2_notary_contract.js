var Notary = artifacts.require("NotaryMock");

module.exports = async function (deployer, network, accounts) {
    const [issuer1, issuer2] = accounts;
    return await deployer.deploy(Notary, [issuer1, issuer2], 2);
};