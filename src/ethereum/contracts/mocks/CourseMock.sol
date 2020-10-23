// SPDX-License-Identifier: MIT
pragma solidity >=0.6.0 <0.8.0;
pragma experimental ABIEncoderV2;

import "../Course.sol";

contract CourseMock is Course {
    constructor(
        address[] memory owners,
        uint8 quorum
    ) Course(owners, quorum) {
        // solhint-disable-previous-line no-empty-blocks
    }

    function enrollStudents(address[] memory students) public {
        for (uint256 i; i < students.length; i++) {
            uint256 index = _students.length;
            _enrolledStudents[students[i]] = Student(index, true);
            _students.push(students[i]);
        }
    }

    function getStudent(address student) public view returns(Student memory) {
        return _enrolledStudents[student];
    }
}
