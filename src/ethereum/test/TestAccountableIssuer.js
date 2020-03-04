const { expectEvent, expectRevert } = require('@openzeppelin/test-helpers');

const Issuer = artifacts.require('IssuerMock');
const AccountableIssuer = artifacts.require('AccountableIssuerMock');

contract('AccountableIssuer', accounts => {
    const [issuer1, issuer2, subject, other] = accounts;
    let acIssuer, issuer, issuerAddress = null;

    describe('constructor', () => {
        it('should successfully deploy the contract', async () => {
            acIssuer = await AccountableIssuer.new([issuer1, issuer2], 2);
            (await acIssuer.isOwner(issuer1)).should.equal(true);
            (await acIssuer.isOwner(issuer2)).should.equal(true);
            assert(acIssuer.quorum(), 2);
        });
    });

    describe('add issuer', () => {
        beforeEach(async () => {
            acIssuer = await AccountableIssuer.new([issuer1, issuer2], 2);
            issuer = await Issuer.deployed([issuer1], 1, { from: issuer1 });
            issuerAddress = issuer.address;
        });

        it('should not add an issuer from a unauthorized address', async () => {
            await expectRevert(
                acIssuer.addIssuer(issuerAddress, { from: other }),
                'Owners: sender is not an owner'
            );
        });

        it('should add an issuer', async () => {
            const { logs } = await acIssuer.addIssuer(issuerAddress, { from: issuer1 });

            expectEvent.inLogs(logs, 'IssuerAdded', {
                addedBy: issuer1,
                issuerAddress: issuerAddress
            });
        });

        it('should not add the same issuer twice', async () => {
            await acIssuer.addIssuer(issuerAddress, { from: issuer1 });

            await expectRevert(
                acIssuer.addIssuer(issuerAddress, { from: issuer2 }),
                'AccountableIssuer: issuer already added'
            );
        });
    });

    describe('collectCredentials', () => { /* TODO */ });

    describe('registerCredential', () => { /* TODO */ });

    describe('verifyCredential', () => { /* TODO */ });
});