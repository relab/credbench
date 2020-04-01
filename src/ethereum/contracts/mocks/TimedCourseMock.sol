pragma solidity >=0.5.13 <0.7.0;
pragma experimental ABIEncoderV2;

import "../TimedCourse.sol";

contract TimedCourseMock is TimedCourse {
    constructor(
        address[] memory owners,
        uint256 quorum,
        uint256 startingTime,
        uint256 endingTime
    ) public TimedCourse(owners, quorum, startingTime, endingTime) {
        // solhint-disable-previous-line no-empty-blocks
    }

    function enrollStudents(address[] memory students) public {
        for (uint256 i; i < students.length; i++) {
            enrolledStudents[students[i]] = true;
        }
    }
}
