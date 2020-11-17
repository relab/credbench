// SPDX-License-Identifier: MIT
pragma solidity >=0.6.0 <0.8.0;
pragma experimental ABIEncoderV2;

import "ct-eth/contracts/node/Node.sol";

/**
 * @title Academic Faculty
 */
contract Faculty is Node {
    mapping(bytes32 => address[]) private _coursesBySemester;

    event SemesterRegistered(address registar, bytes32 semester);

    constructor(address[] memory owners, uint8 quorum)
        Node(Role.Inner, owners, quorum) {
    }

    function semesterExists(bytes32 semester) public view returns (bool) {
        return _coursesBySemester[semester].length != 0;
    }

    function registerSemester(bytes32 semester, address[] memory courses) public onlyOwner {
        // TODO: limit the number of courses
        require(!semesterExists(semester), "Faculty/already registered");
        require(courses.length > 0, "Faculty/empty courses list");
        _coursesBySemester[semester] = courses;
        emit SemesterRegistered(msg.sender, semester);
    }

    function getCoursesBySemester(bytes32 semester) public view returns (address[] memory) {
        require(semesterExists(semester), "Faculty/not registered");
        return _coursesBySemester[semester];
    }
}
