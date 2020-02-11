var Course = artifacts.require("CourseMock");

module.exports = async function (deployer, network, accounts) {
    const [issuer1, issuer2] = accounts;
    let beginTimestamp = new Date().getTime();
    let endTimestamp = new Date().setTime(beginTimestamp + (1 * 60 * 1000));
    return await deployer.deploy(Course, [issuer1, issuer2], 2, beginTimestamp.toString(), endTimestamp.toString());
};