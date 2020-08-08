// SPDX-License-Identifier: MIT
pragma solidity >=0.6.0 <0.7.0;
pragma experimental ABIEncoderV2;

import "../Course.sol";

contract CourseMock is Course {
    constructor(
        address[] memory owners,
        uint256 quorum
    ) public Course(owners, quorum) {
        // solhint-disable-previous-line no-empty-blocks
    }

    function enrollStudents(address[] memory students) public {
        for (uint256 i; i < students.length; i++) {
            enrolledStudents[students[i]] = true;
        }
    }
}
