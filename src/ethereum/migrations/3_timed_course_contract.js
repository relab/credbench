var CredentialSum = artifacts.require("CredentialSum");
var TimedCourse = artifacts.require("TimedCourseMock");

module.exports = async function (deployer, network, accounts) {
    const [issuer1, issuer2] = accounts;

    deployer.link(CredentialSum, TimedCourse);

    let beginTimestamp = new Date().getTime();
    let endTimestamp = new Date().setTime(beginTimestamp + (1 * 60 * 1000));
    return await deployer.deploy(TimedCourse, [issuer1, issuer2], 2, beginTimestamp, endTimestamp);
};