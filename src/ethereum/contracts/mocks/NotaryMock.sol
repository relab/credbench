pragma solidity >=0.5.13;

import "../Notary.sol";

contract NotaryMock is Notary {
    constructor(address[] memory owners, uint quorum) public Notary(owners, quorum) {
        // solhint-disable-previous-line no-empty-blocks
    }
}