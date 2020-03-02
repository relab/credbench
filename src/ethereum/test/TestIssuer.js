const { BN, expectEvent, expectRevert, time, constants } = require('@openzeppelin/test-helpers');
const { expect } = require('chai');
const assertFailure = require('./helpers/assertFailure');

const Issuer = artifacts.require('IssuerMock');

contract('Issuer', accounts => {
    const [issuer1, issuer2, issuer3, subject1, subject2, other] = accounts;
    let issuer = null;
    const reason = web3.utils.keccak256(web3.utils.toHex('revoked'));
    const digest1 = web3.utils.keccak256(web3.utils.toHex('cert1'));
    const digest2 = web3.utils.keccak256(web3.utils.toHex('cert2'));
    const digest3 = web3.utils.keccak256(web3.utils.toHex('cert3'));
    const zeroDigest = "0x0000000000000000000000000000000000000000000000000000000000000000";

    describe('constructor', () => {
        it('should successfully deploy the contract initializing the owners', async () => {
            issuer = await Issuer.new([issuer1, issuer2], 2);
            (await issuer.isOwner(issuer1)).should.equal(true);
            (await issuer.isOwner(issuer2)).should.equal(true);
            assert(issuer.quorum(), 2);
        });

        it('should successfully get a deployed contract', async () => {
            issuer = await Issuer.deployed([issuer1, issuer2], 2);
            (await issuer.isOwner(issuer1)).should.equal(true);
            (await issuer.isOwner(issuer2)).should.equal(true);
            assert(issuer.quorum(), 2);
        });
    });

    describe('issue', () => {
        it('should successfully create a signed credential proof', async () => {
            issuer = await Issuer.new([issuer1], 1);
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            const credential = await issuer.issuedCredentials(digest1);
            assert(credential.signed, 1);
            (credential.subjectSigned).should.equal(false);
            expect(await time.latestBlock()).to.be.bignumber.equal(new BN(credential.insertedBlock));
            assert(credential.subject, subject1);
            assert(credential.digest, digest1);
            (await issuer.ownersSigned(digest1, issuer1)).should.equal(true);
        });

        it('should emits an event when a credential proof is issued', async () => {
            issuer = await Issuer.new([issuer1], 1);
            let { logs } = await issuer.registerCredential(subject1, digest1, { from: issuer1 });

            let block = await time.latestBlock();
            expectEvent.inLogs(logs, 'CredentialIssued', {
                digest: digest1,
                subject: subject1,
                issuer: issuer1,
                insertedBlock: block
            });
        });

        it('should not issue an already issued credential proof', async () => {
            issuer = await Issuer.new([issuer1], 1);
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });

            await expectRevert(
                issuer.registerCredential(subject1, digest1, { from: issuer1 }),
                'Issuer: sender already signed'
            );
        });

        it('should not allow an issuer to issue credential proof to themselves', async () => {
            issuer = await Issuer.new([issuer1, issuer2], 1);

            await expectRevert(
                issuer.registerCredential(issuer1, digest1, { from: issuer1 }),
                'Issuer: subject cannot be the issuer'
            );

            await expectRevert(
                issuer.registerCredential(issuer1, digest1, { from: issuer2 }),
                'Issuer: subject cannot be the issuer'
            );
        });

        it('should not issue a credential proof from a unauthorized address', async () => {
            issuer = await Issuer.new([issuer1], 1);
            await expectRevert(
                issuer.registerCredential(subject1, digest1, { from: issuer3 }),
                'Owners: sender is not an owner'
            );
        });

        it('should not issue a credential proof with the same digest for different subjects', async () => {
            issuer = await Issuer.new([issuer1, issuer2], 2);

            await issuer.registerCredential(subject1, digest1, { from: issuer1 });

            await expectRevert(
                issuer.registerCredential(subject2, digest1, { from: issuer2 }),
                'Issuer: credential already issued for other subject'
            );
        });

        it('should compute a quorum of owners signatures', async () => {
            issuer = await Issuer.new([issuer1, issuer2, issuer3], 2);
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            await issuer.registerCredential(subject1, digest1, { from: issuer2 });

            const length = await issuer.ownersLength();
            let quorum = await issuer.quorum();
            for (let i = 0; i < length; i++) {
                const owner = await issuer.owners(i);
                const signed = await issuer.ownersSigned(digest1, owner);
                if (signed)--quorum;
            }
            (quorum).should.equal(0);
        });

        it('should not allow issue a new credential before sign the previous', async () => {
            issuer = await Issuer.new([issuer1, issuer2, issuer3], 2);
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });

            const credential1 = await issuer.issuedCredentials(digest1);
            assert(credential1.signed, 2);
            assert(credential1.subject, subject1);
            assert(credential1.digest, digest1);
            (await issuer.ownersSigned(digest1, issuer1)).should.equal(true);

            await expectRevert(
                issuer.registerCredential(subject1, digest2, { from: issuer1 }),
                'Issuer: previous credential must be signed before issue a new one'
            );
        });
    });

    describe('confirm the proof emission', () => {
        beforeEach(async () => {
            issuer = await Issuer.new([issuer1, issuer2, issuer3], 2);
        });

        it('should revert when attempt to confirm a credential proof without a quorum formed', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });

            await expectRevert(
                issuer.confirmCredential(digest1, { from: subject1 }),
                'Issuer: not sufficient quorum of signatures'
            );

            const credential = await issuer.issuedCredentials(digest1);
            (credential.subjectSigned).should.equal(false);
        });

        it('should mark a credential proof as signed when it was signed by a quorum and by the subject', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            await issuer.registerCredential(subject1, digest1, { from: issuer2 });
            await issuer.confirmCredential(digest1, { from: subject1 });

            const credential = await issuer.issuedCredentials(digest1);
            (credential.subjectSigned).should.equal(true);
        });

        it('should emit an event when a credential proof is signed by all required parties', async () => {
            const previousBlockNumber = await time.latestBlock();

            let { logs } = await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            let lastBlockNumber = await time.latestBlock();
            expectEvent.inLogs(logs, 'CredentialIssued', {
                digest: digest1,
                subject: subject1,
                issuer: issuer1,
                insertedBlock: lastBlockNumber
            });
            expectEvent.inLogs(logs, 'CredentialSigned', {
                signer: issuer1,
                digest: digest1,
                signedBlock: lastBlockNumber
            });

            ({ logs } = await issuer.registerCredential(subject1, digest1, { from: issuer2 }));
            lastBlockNumber = await time.latestBlock();
            expectEvent.inLogs(logs, 'CredentialSigned', {
                signer: issuer2,
                digest: digest1,
                signedBlock: lastBlockNumber
            });

            ({ logs } = await issuer.confirmCredential(digest1, { from: subject1 }));
            lastBlockNumber = await time.latestBlock();
            expectEvent.inLogs(logs, 'CredentialSigned', {
                signer: subject1,
                digest: digest1,
                signedBlock: lastBlockNumber
            });

            const eventList = await issuer.getPastEvents("allEvents", { fromBlock: previousBlockNumber, toBlock: lastBlockNumber });
            (eventList.length).should.equal(4);
        });

        it('should only allow credential proof requests from the correct subject', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            await issuer.registerCredential(subject1, digest1, { from: issuer2 });

            await expectRevert(
                issuer.confirmCredential(digest1, { from: subject2 }),
                'Issuer: subject is not related with this credential'
            );
        });

        it('should not allow a subject to re-sign a issued credential proof', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            await issuer.registerCredential(subject1, digest1, { from: issuer2 });
            await issuer.confirmCredential(digest1, { from: subject1 });

            await expectRevert(
                issuer.confirmCredential(digest1, { from: subject1 }),
                'Issuer: subject already signed this credential'
            );
        });

        it('should certified that a credential proof was signed by all parties', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            await issuer.registerCredential(subject1, digest1, { from: issuer2 });

            (await issuer.certified(digest1)).should.equal(false);

            await issuer.confirmCredential(digest1, { from: subject1 });

            (await issuer.certified(digest1)).should.equal(true);
        });

        it('should check if a list of credentials was signed by all parties', async () => {
            const digests = [digest1, digest2, digest3];

            for (let i = 0; i < digests.length; i++) {
                await issuer.registerCredential(subject1, digests[i], { from: issuer1 });
                await issuer.registerCredential(subject1, digests[i], { from: issuer2 });
                if (i < digests.length - 1) {
                    await issuer.confirmCredential(digests[i], { from: subject1 });
                }
                await time.increase(time.duration.seconds(1));
            }
            let result = await issuer.checkCredentials(digests);
            (result).should.equal(false);

            await issuer.confirmCredential(digest3, { from: subject1 });
            result = await issuer.checkCredentials(digests);
            (result).should.equal(true);
        });
    });

    describe('revoke', () => {
        beforeEach(async () => {
            issuer = await Issuer.new([issuer1, issuer2], 2);
        });

        it('should not revoke a credential proof from a un-authorized address', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            await expectRevert(
                issuer.revokeCredential(digest1, reason, { from: issuer3 }),
                'Owners: sender is not an owner'
            );
        });

        it('should not revoke a not issued credential proof', async () => {
            await expectRevert(
                issuer.revokeCredential(digest1, reason, { from: issuer1 }),
                'Issuer: no credential proof found'
            );
        });

        it('should verify if a credential proof was revoked based on the digest1', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            (await issuer.isRevoked(digest1)).should.equal(false);

            await issuer.revokeCredential(digest1, reason, { from: issuer1 });
            (await issuer.isRevoked(digest1)).should.equal(true);
        });

        it('should successfully create a revocation proof by any owner', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            await issuer.revokeCredential(digest1, reason, { from: issuer1 });

            const revocation = await issuer.revokedCredentials(digest1);
            expect(await time.latestBlock()).to.be.bignumber.equal(new BN(revocation.revokedBlock));
            assert(revocation.reason, reason);
            assert(revocation.subject, subject1);
            assert(revocation.issuer, issuer1);
        });

        it('should emits an event when create a revocation proof', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            const { logs } = await issuer.revokeCredential(digest1, reason, { from: issuer2 });
            const blockNumber = await time.latestBlock();

            expectEvent.inLogs(logs, 'CredentialRevoked', {
                digest: digest1,
                subject: subject1,
                revoker: issuer2,
                revokedBlock: blockNumber,
                reason: reason
            });

            const credential = await issuer.issuedCredentials(digest1);
            assert(credential.subject, constants.ZERO_ADDRESS);
            assert(credential.issuer, constants.ZERO_ADDRESS);
            assert(credential.insertedBlock, 0);

            (await issuer.certified(digest1)).should.equal(false);
        });
    });

    describe('aggregate', () => {
        const digests = [digest1, digest2, digest3];

        beforeEach(async () => {
            issuer = await Issuer.new([issuer1], 1);
        });

        describe('list of credentials', () => {
            const expected = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', digests));

            beforeEach(async () => {
                for (d of digests) {
                    await issuer.registerCredential(subject1, d, { from: issuer1 });
                    await issuer.confirmCredential(d, { from: subject1 });
                    await time.increase(time.duration.seconds(1));

                    (await issuer.certified(d)).should.equal(true);
                }
            });

            it('should aggregate all credentials of a subject', async () => {
                const aggregated = await issuer.aggregateCredentials.call(subject1);// don't emit event
                (aggregated).should.equal(expected);
            });

            it('should emit an event when aggregate all credentials of a subject', async () => {
                const { logs } = await issuer.aggregateCredentials(subject1, { from: issuer1 });
                const blockNumber = await time.latestBlock();
                expectEvent.inLogs(logs, 'AggregatedProof', {
                    aggregator: issuer1,
                    subject: subject1,
                    proof: expected,
                    aggregatedBlock: blockNumber
                });
            });

            it('should return the already aggregated proof', async () => {
                await issuer.aggregateCredentials(subject1);
                const storedProof = await issuer.getProof(subject1);

                const aggregated = await issuer.aggregateCredentials.call(subject1);

                (aggregated).should.equal(storedProof);
            });
        });

        it('should fail if there are any credential of a subject that isn\'t signed by all parties', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });

            await expectRevert(
                issuer.aggregateCredentials(subject1),
                'Issuer: there are unsigned credentials'
            );

            await issuer.confirmCredential(digest1, { from: subject1 });
            (await issuer.certified(digest1)).should.equal(true);
        });

        it('should return the credential hash if only one credential exists', async () => {
            await issuer.registerCredential(subject1, digest1, { from: issuer1 });
            await issuer.confirmCredential(digest1, { from: subject1 });

            const aggregated = await issuer.aggregateCredentials.call(subject1);
            let expected = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', [digest1]));

            (aggregated).should.equal(expected);
        });

        it('should revert if there is no credential to be aggregated for a given subject', async () => {
            await expectRevert(
                issuer.aggregateCredentials(subject1),
                'Issuer: there is no credential for the given subject'
            );
        });
    });

    describe('verify', () => {
        const digests = [digest1, digest2, digest3];
        const expected = web3.utils.keccak256(web3.eth.abi.encodeParameter('bytes32[]', digests));

        beforeEach(async () => {
            issuer = await Issuer.new([issuer1], 1);
            for (d of digests) {
                await issuer.registerCredential(subject1, d, { from: issuer1 });
                await issuer.confirmCredential(d, { from: subject1 });
                await time.increase(time.duration.seconds(1));
            }
            issuer.aggregateCredentials(subject1);
        });

        it('should successfully verify the given credential', async () => {
            issuer.verifyCredential(subject1, expected);
            const proof = await issuer.getProof(subject1);
            (proof).should.equal(expected);
        });

        it('should revert if the given credential doesn\'t match the stored proof', async () => {
            await expectRevert(
                issuer.verifyCredential(subject1, digest1),
                'Issuer: given credential doesn\'t match with stored proof'
            );
        });

        it('should revert if there is no credential to be verified for a given subject', async () => {
            await expectRevert(
                issuer.verifyCredential(subject2, expected),
                'Issuer: there is no aggregated proof to verify'
            );
        });
    });
});
