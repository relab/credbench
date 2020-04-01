pragma solidity >=0.5.13 <0.7.0;
pragma experimental ABIEncoderV2;

import "../TimedFaculty.sol";

contract TimedFacultyMock is TimedFaculty {
    constructor(address[] memory owners, uint256 quorum)
        public
        TimedFaculty(owners, quorum)
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
