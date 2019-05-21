const { shouldFail } = require('openzeppelin-test-helpers');
const Owners = artifacts.require('Owners');

contract('Owners', accounts => {
    const [owner1, owner2] = accounts;
    let contract = null;

    describe('constructor', () => {
        it('should successfully deploy the contract', async () => {
            contract = await Owners.new([owner1, owner2], 2);
            (await contract.owners(owner1)).should.equal(true);
            (await contract.owners(owner2)).should.equal(true);
            assert(contract.quorum(), 2);
        });

        it('should require a non-empty array of owners', async () => {
            await shouldFail.reverting.withMessage(Owners.new([], 0), 'Owners: not enough owners');
        });

        it('should require a quorum value greater than 0', async () => {
            await shouldFail.reverting.withMessage(Owners.new([owner1], 0), 'Owners: quorum out of range');
        });

        it('should require a quorum value less than the amount of owners', async () => {
            await shouldFail.reverting.withMessage(Owners.new([owner1, owner2], 3), 'Owners: quorum out of range');
        });

        it('should verify if an address is an owner', async () => {
            contract = await Owners.new([owner1], 1);
            (await contract.isOwner({ from: owner1 })).should.equal(true);
            (await contract.isOwner({ from: owner2 })).should.equal(false);
        });
    });
});
