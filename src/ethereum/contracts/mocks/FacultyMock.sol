pragma solidity >=0.5.13;
pragma experimental ABIEncoderV2;

import "../Faculty.sol";

contract FacultyMock is Faculty {
    constructor(address[] memory owners, uint256 quorum)
        public
        Faculty(owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }

    function setBalance() public payable {
        // address(this).balance += msg.value;
    }

    function getBalance() public view returns (uint256) {
        return address(this).balance;
    }
}
