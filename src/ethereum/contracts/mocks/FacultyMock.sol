// SPDX-License-Identifier: MIT
pragma solidity >=0.6.0 <0.8.0;
pragma experimental ABIEncoderV2;

import "../Faculty.sol";

contract FacultyMock is Faculty {
    constructor(address[] memory owners, uint8 quorum)
        Faculty(owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }

    function addCourse(address course) public {
        addChild(course);
    }

    function getBalance() public view returns (uint256) {
        return address(this).balance;
    }
}
