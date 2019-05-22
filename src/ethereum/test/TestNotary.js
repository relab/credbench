const { BN, expectEvent, shouldFail, time, constants } = require('openzeppelin-test-helpers');
const { expect } = require('chai');

const Notary = artifacts.require('Notary');

contract('Notary', accounts => {
    const [issuer1, issuer2, issuer3, subject1, subject2] = accounts;
    let notary = null;
    const digest = web3.utils.sha3('QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG');

    describe('constructor', () => {
        it('should successfully deploy the contract initializing the owners', async () => {
            notary = await Notary.new([issuer1, issuer2], 2);
            (await notary.owners(issuer1)).should.equal(true);
            (await notary.owners(issuer2)).should.equal(true);
            assert(notary.quorum(), 2);
        });
    });

    describe('issue', () => {
        it('should successfully create a signed credential proof', async () => {
            notary = await Notary.new([issuer1], 1);
            await notary.issue(subject1, digest, { from: issuer1 });
            const credential = await notary.issued(digest);
            assert(credential.signed, 1);
            (credential.subjectSigned).should.equal(false);
            expect(await time.latestBlock()).to.be.bignumber.equal(new BN(credential.insertedBlock));
            assert(credential.subject, subject1);
            assert(credential.digest, digest);
            (await notary.ownersSigned(digest, issuer1)).should.equal(true);
        });

        it('should not issue a already issued credential proof', async () => {
            notary = await Notary.new([issuer1], 1);
            await notary.issue(subject1, digest, { from: issuer1 });

            await shouldFail.reverting.withMessage(
                notary.issue(subject1, digest, { from: issuer1 }),
                'Notary: sender already signed'
            );
        });

        it('should not issue a credential proof from a un-authorized address', async () => {
            notary = await Notary.new([issuer1], 1);
            await shouldFail.reverting.withMessage(
                notary.issue(subject1, digest, { from: issuer2 }),
                'Owners: sender is not an owner'
            );
        });

        it('should compute a quorum of owners signatures', async () => {
            notary = await Notary.new([issuer1, issuer2, issuer3], 2);
            await notary.issue(subject1, digest, { from: issuer1 });
            await notary.issue(subject1, digest, { from: issuer2 });

            const length = await notary.ownersLength();
            let quorum = await notary.quorum();
            for (let i = 0; i < length; i++) {
                const owner = await notary.allOwners(i);
                const signed = await notary.ownersSigned(digest, owner);
                if (signed)--quorum;
            }
            (quorum).should.equal(0);
        });
    });

    describe('request proof', () => {
        beforeEach(async () => {
            notary = await Notary.new([issuer1, issuer2, issuer3], 2);
        });

        it('should revert when requesting a credential proof without a quorum formed', async () => {
            await notary.issue(subject1, digest, { from: issuer1 });

            await shouldFail.reverting.withMessage(
                notary.requestProof(digest, { from: subject1 }),
                'Notary: not sufficient quorum of signatures'
            );

            const credential = await notary.issued(digest);
            (credential.subjectSigned).should.equal(false);
        });

        it('should issue a requested credential proof if it was signed by a quorum', async () => {
            await notary.issue(subject1, digest, { from: issuer1 });
            await notary.issue(subject1, digest, { from: issuer2 });
            await notary.requestProof(digest, { from: subject1 });

            const credential = await notary.issued(digest);
            (credential.subjectSigned).should.equal(true);
        });

        it('should emits an event when a requested credential proof is signed by all required parties', async () => {
            await notary.issue(subject1, digest, { from: issuer1 });
            await notary.issue(subject1, digest, { from: issuer2 });

            const { logs } = await notary.requestProof(digest, { from: subject1 });
            expectEvent.inLogs(logs, 'CredentialIssued', {
                digest: digest,
                subject: subject1
            });
        });

        it('should only allow credential proof requests from the correct subject', async () => {
            await notary.issue(subject1, digest, { from: issuer1 });
            await notary.issue(subject1, digest, { from: issuer2 });

            await shouldFail.reverting.withMessage(
                notary.requestProof(digest, { from: subject2 }),
                'Notary: subject is not related with this credential'
            );
        });

        it('should not allow a subject to re-sign a issued credential proof', async () => {
            await notary.issue(subject1, digest, { from: issuer1 });
            await notary.issue(subject1, digest, { from: issuer2 });
            await notary.requestProof(digest, { from: subject1 });

            await shouldFail.reverting.withMessage(
                notary.requestProof(digest, { from: subject1 }),
                'Notary: subject already signed this credential'
            );
        });

        it('should certified that a credential proof was signed by all parties', async () => {
            await notary.issue(subject1, digest, { from: issuer1 });
            await notary.issue(subject1, digest, { from: issuer2 });

            (await notary.certified(digest)).should.equal(false);

            await notary.requestProof(digest, { from: subject1 });

            (await notary.certified(digest)).should.equal(true);
        });
    });

    describe('revoke', () => {
        beforeEach(async () => {
            notary = await Notary.new([issuer1, issuer2], 2);
        });

        it('should not revoke a credential proof from a un-authorized address', async () => {
            await notary.issue(subject1, digest, { from: issuer1 });
            await shouldFail.reverting.withMessage(
                notary.revoke(digest, { from: issuer3 }),
                'Owners: sender is not an owner'
            );
        });

        it('should not revoke a not issued credential proof', async () => {
            await shouldFail.reverting.withMessage(
                notary.revoke(digest, { from: issuer1 }),
                'Notary: no credential proof found'
            );
        });

        it('should verify if a credential proof was revoked based on the digest', async () => {
            await notary.issue(subject1, digest, { from: issuer1 });
            (await notary.wasRevoked(digest)).should.equal(false);

            await notary.revoke(digest, { from: issuer1 });
            (await notary.wasRevoked(digest)).should.equal(true);
        });

        it('should successfully emits a revocation proof by any owner', async () => {
            await notary.issue(subject1, digest, { from: issuer1 });
            const { logs } = await notary.revoke(digest, { from: issuer2 });
            const blockNumber = await time.latestBlock();

            expectEvent.inLogs(logs, 'CredentialRevoked', {
                digest: digest,
                issuer: issuer2,
                revokedBlock: blockNumber
            });

            const credential = await notary.issued(digest);
            assert(credential.subject, constants.ZERO_ADDRESS);
            assert(credential.issuer, constants.ZERO_ADDRESS);
            assert(credential.insertedBlock, 0);

            (await notary.certified(digest)).should.equal(false);
        });
    });
});
