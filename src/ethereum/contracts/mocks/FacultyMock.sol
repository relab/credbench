// SPDX-License-Identifier: MIT
pragma solidity >=0.6.0 <0.8.0;
pragma experimental ABIEncoderV2;

import "../Faculty.sol";

contract FacultyMock is Faculty {
    constructor(address[] memory owners, uint256 quorum)
        Faculty(owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }

    function addCourse(address course, bytes32 semester) public {
        coursesBySemester[semester].push(course);
        addIssuer(course);
    }

    function setBalance() public payable {
        // address(this).balance += msg.value;
    }

    function getBalance() public view returns (uint256) {
        return address(this).balance;
    }
}
