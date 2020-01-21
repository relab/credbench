pragma solidity >=0.5.13;

import "../Notary.sol";

contract NotaryMock is Notary {
    constructor(address[] memory owners, uint quorum) Notary(owners, quorum) public {}
}