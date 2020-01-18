pragma solidity >=0.5.0 <0.6.0; 

import "../Notary.sol";

contract NotaryMock is Notary {
    constructor(address[] memory owners, uint quorum) Notary(owners, quorum) public {}
}