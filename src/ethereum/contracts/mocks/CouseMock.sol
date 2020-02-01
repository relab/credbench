pragma solidity >=0.5.13;
pragma experimental ABIEncoderV2;

import "../Course.sol";

contract CourseMock is Course {
    constructor(
        address[] memory owners,
        uint256 quorum,
        uint256 startingTime,
        uint256 endingTime
    ) public Course(owners, quorum, startingTime, endingTime) {
        // solhint-disable-previous-line no-empty-blocks
    }

    function enrollStudents(address[] memory students) public {
        for (uint256 i; i < students.length; i++) {
            enrolledStudents[students[i]] = true;
        }
    }
}
