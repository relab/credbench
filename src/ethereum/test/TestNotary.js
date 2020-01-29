const { BN, expectEvent, expectRevert, time, constants } = require('@openzeppelin/test-helpers');
const { expect } = require('chai');
const assertFailure = require('./helpers/assertFailure');

const Notary = artifacts.require('NotaryMock');

contract('Notary', accounts => {
    const [issuer1, issuer2, issuer3, subject1, subject2] = accounts;
    let notary = null;
    const reason = web3.utils.soliditySha3('revoked');
    const digest1 = web3.utils.soliditySha3('cert1');
    const digest2 = web3.utils.soliditySha3('cert2');
    const digest3 = web3.utils.soliditySha3('cert3');
    const zeroDigest = "0x0000000000000000000000000000000000000000000000000000000000000000";

    describe('constructor', () => {
        it('should successfully deploy the contract initializing the owners', async () => {
            notary = await Notary.new([issuer1, issuer2], 2);
            (await notary.isOwner(issuer1)).should.equal(true);
            (await notary.isOwner(issuer2)).should.equal(true);
            assert(notary.quorum(), 2);
        });

        it('should successfully get a deployed contract', async () => {
            notary = await Notary.deployed([issuer1, issuer2], 2);
            (await notary.isOwner(issuer1)).should.equal(true);
            (await notary.isOwner(issuer2)).should.equal(true);
            assert(notary.quorum(), 2);
        });
    });

    describe('issue', () => {
        it('should successfully create a signed credential proof', async () => {
            notary = await Notary.new([issuer1], 1);
            await notary.issue(subject1, digest1, { from: issuer1 });
            const credential = await notary.issuedCredentials(digest1);
            assert(credential.signed, 1);
            (credential.subjectSigned).should.equal(false);
            expect(await time.latestBlock()).to.be.bignumber.equal(new BN(credential.insertedBlock));
            assert(credential.subject, subject1);
            assert(credential.digest, digest1);
            assert(credential.previousDigest, zeroDigest);
            (await notary.ownersSigned(digest1, issuer1)).should.equal(true);
        });

        it('should not issue an already issued credential proof', async () => {
            notary = await Notary.new([issuer1], 1);
            await notary.issue(subject1, digest1, { from: issuer1 });

            await expectRevert(
                notary.issue(subject1, digest1, { from: issuer1 }),
                'Notary: sender already signed'
            );
        });

        it('should not allow an issuer to issue credential proof to themselves', async () => {
            notary = await Notary.new([issuer1, issuer2], 1);

            await expectRevert(
                notary.issue(issuer1, digest1, { from: issuer1 }),
                'Notary: subject cannot be the issuer'
            );

            await expectRevert(
                notary.issue(issuer1, digest1, { from: issuer2 }),
                'Notary: subject cannot be the issuer'
            );
        });

        it('should not issue a credential proof from a unauthorized address', async () => {
            notary = await Notary.new([issuer1], 1);
            await expectRevert(
                notary.issue(subject1, digest1, { from: issuer3 }),
                'Owners: sender is not an owner'
            );
        });

        it('should not issue a credential proof with the same digest for different subjects', async () => {
            notary = await Notary.new([issuer1, issuer2], 2);

            await notary.issue(subject1, digest1, { from: issuer1 });

            await expectRevert(
                notary.issue(subject2, digest1, { from: issuer2 }),
                'Notary: credential already issued for other subject'
            );
        });

        it('should compute a quorum of owners signatures', async () => {
            notary = await Notary.new([issuer1, issuer2, issuer3], 2);
            await notary.issue(subject1, digest1, { from: issuer1 });
            await notary.issue(subject1, digest1, { from: issuer2 });

            const length = await notary.ownersLength();
            let quorum = await notary.quorum();
            for (let i = 0; i < length; i++) {
                const owner = await notary.owners(i);
                const signed = await notary.ownersSigned(digest1, owner);
                if (signed)--quorum;
            }
            (quorum).should.equal(0);
        });

        it('should successfully create a signed credential proof linked with the previous of the same subject', async () => {
            notary = await Notary.new([issuer1, issuer2, issuer3], 2);
            await notary.issue(subject1, digest1, { from: issuer1 });
            await notary.issue(subject1, digest1, { from: issuer2 });
            await notary.requestProof(digest1, { from: subject1 });

            const credential1 = await notary.issuedCredentials(digest1);
            assert(credential1.signed, 2);
            assert(credential1.subject, subject1);
            assert(credential1.digest, digest1);
            assert(credential1.previousDigest, zeroDigest);
            (await notary.ownersSigned(digest1, issuer1)).should.equal(true);

            await time.increase(time.duration.seconds(1)); // mines a new block with timestamp 1 second ahead.

            // Issuing a new certificate
            await notary.issue(subject1, digest2, { from: issuer1 });

            const credential2 = await notary.issuedCredentials(digest2);

            expect(credential2.blockTimestamp).to.be.bignumber.above(credential1.blockTimestamp);

            assert(credential2.signed, 1);
            assert(credential2.subject, subject1);
            assert(credential2.digest, digest2);
            assert(credential2.previousDigest, digest1);
        });

        it('should not allow issue a new certificate before sign the previous', async () => {
            notary = await Notary.new([issuer1, issuer2, issuer3], 2);
            await notary.issue(subject1, digest1, { from: issuer1 });

            const credential1 = await notary.issuedCredentials(digest1);
            assert(credential1.signed, 2);
            assert(credential1.subject, subject1);
            assert(credential1.digest, digest1);
            assert(credential1.previousDigest, zeroDigest);
            (await notary.ownersSigned(digest1, issuer1)).should.equal(true);

            await expectRevert(
                notary.issue(subject1, digest2, { from: issuer1 }),
                'Notary: previous credential must be signed before issue a new one'
            );
        });
    });

    describe('request proof', () => {
        beforeEach(async () => {
            notary = await Notary.new([issuer1, issuer2, issuer3], 2);
        });

        it('should revert when requesting a credential proof without a quorum formed', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });

            await expectRevert(
                notary.requestProof(digest1, { from: subject1 }),
                'Notary: not sufficient quorum of signatures'
            );

            const credential = await notary.issuedCredentials(digest1);
            (credential.subjectSigned).should.equal(false);
        });

        it('should issue a requested credential proof if it was signed by a quorum', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            await notary.issue(subject1, digest1, { from: issuer2 });
            await notary.requestProof(digest1, { from: subject1 });

            const credential = await notary.issuedCredentials(digest1);
            (credential.subjectSigned).should.equal(true);
        });

        it('should emits an event when a requested credential proof is signed by all required parties', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            await notary.issue(subject1, digest1, { from: issuer2 });

            const { logs } = await notary.requestProof(digest1, { from: subject1 });
            expectEvent.inLogs(logs, 'CredentialIssued', {
                digest: digest1,
                subject: subject1,
                issuer: issuer1,
                previousDigest: zeroDigest
            });
        });

        it('should only allow credential proof requests from the correct subject', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            await notary.issue(subject1, digest1, { from: issuer2 });

            await expectRevert(
                notary.requestProof(digest1, { from: subject2 }),
                'Notary: subject is not related with this credential'
            );
        });

        it('should not allow a subject to re-sign a issued credential proof', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            await notary.issue(subject1, digest1, { from: issuer2 });
            await notary.requestProof(digest1, { from: subject1 });

            await expectRevert(
                notary.requestProof(digest1, { from: subject1 }),
                'Notary: subject already signed this credential'
            );
        });

        it('should certified that a credential proof was signed by all parties', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            await notary.issue(subject1, digest1, { from: issuer2 });

            (await notary.certified(digest1)).should.equal(false);

            await notary.requestProof(digest1, { from: subject1 });

            (await notary.certified(digest1)).should.equal(true);
        });
    });

    describe('revoke', () => {
        beforeEach(async () => {
            notary = await Notary.new([issuer1, issuer2], 2);
        });

        it('should not revoke a credential proof from a un-authorized address', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            await expectRevert(
                notary.revoke(digest1, reason, { from: issuer3 }),
                'Owners: sender is not an owner'
            );
        });

        it('should not revoke a not issued credential proof', async () => {
            await expectRevert(
                notary.revoke(digest1, reason, { from: issuer1 }),
                'Notary: no credential proof found'
            );
        });

        it('should verify if a credential proof was revoked based on the digest1', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            (await notary.wasRevoked(digest1)).should.equal(false);

            await notary.revoke(digest1, reason, { from: issuer1 });
            (await notary.wasRevoked(digest1)).should.equal(true);
        });

        it('should successfully create a revocation proof by any owner', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            await notary.revoke(digest1, reason, { from: issuer1 });

            const revocation = await notary.revokedCredentials(digest1);
            expect(await time.latestBlock()).to.be.bignumber.equal(new BN(revocation.revokedBlock));
            assert(revocation.reason, reason);
            assert(revocation.subject, subject1);
            assert(revocation.issuer, issuer1);
        });

        it('should emits an event when create a revocation proof', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            const { logs } = await notary.revoke(digest1, reason, { from: issuer2 });
            const blockNumber = await time.latestBlock();

            expectEvent.inLogs(logs, 'CredentialRevoked', {
                digest: digest1,
                subject: subject1,
                issuer: issuer2,
                revokedBlock: blockNumber,
                reason: reason
            });

            const credential = await notary.issuedCredentials(digest1);
            assert(credential.subject, constants.ZERO_ADDRESS);
            assert(credential.issuer, constants.ZERO_ADDRESS);
            assert(credential.insertedBlock, 0);

            (await notary.certified(digest1)).should.equal(false);
        });
    });

    describe('aggregate', () => {
        beforeEach(async () => {
            notary = await Notary.new([issuer1], 1);
        });

        it('should aggregate all certificates of a subject', async () => {
            for (d of [digest1, digest2, digest3]) {
                await notary.issue(subject1, d, { from: issuer1 });
                await notary.requestProof(d, { from: subject1 });
                await time.increase(time.duration.seconds(1));

                (await notary.certified(d)).should.equal(true);
            }

            let aggregated = await notary.aggregate(subject1);

            let expected = digest1;
            for (d of [digest2, digest3]) {
                expected = web3.utils.soliditySha3(expected, d);
            }
            (aggregated).should.equal(expected);
        });

        it('should revert if there is no certificate for a given subject', async () => {
            await expectRevert(
                notary.aggregate(subject1),
                'Notary: there is no certificate for the given subject'
            );
        });

        it('should return the certificate hash if only one certificate exists', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });
            await notary.requestProof(digest1, { from: subject1 });

            let aggregated = await notary.aggregate(subject1);

            (aggregated).should.equal(digest1);
        });

        it('should fail if there are any certificate of a subject that isn\'t signed by all parties', async () => {
            await notary.issue(subject1, digest1, { from: issuer1 });

            let assertion = await assertFailure(notary.aggregate(subject1));
            expect(assertion.message).to.include('invalid opcode');

            await notary.requestProof(digest1, { from: subject1 });
            await time.increase(time.duration.seconds(1));

            await notary.issue(subject1, digest2, { from: issuer1 });

            assertion = await assertFailure(notary.aggregate(subject1));
            expect(assertion.message).to.include('invalid opcode');
        });
    });
});
