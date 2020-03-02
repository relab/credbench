pragma solidity >=0.5.13 <0.7.0;

import "../Issuer.sol";

contract IssuerMock is Issuer {
    constructor(address[] memory owners, uint256 quorum)
        public
        Issuer(owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }
}
