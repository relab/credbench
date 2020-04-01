pragma solidity >=0.5.13 <0.7.0;
pragma experimental ABIEncoderV2;

import "bbchain-contracts/contracts/Timed.sol";
import "./Course.sol";

/**
 * @title Academic TimedCourse
 */
contract TimedCourse is Timed, Course {
    /**
    * @dev Constructor creates a TimedCourse contract
    */
    constructor(
        address[] memory owners,
        uint256 quorum,
        uint256 startingTime,
        uint256 endingTime
    ) public Course(owners, quorum) Timed(startingTime, endingTime) {
        // solhint-disable-previous-line no-empty-blocks
    }

    /**
     * @dev Adds a student to the course while course is running.
     * @param student the address of the student to be added.
     */
    function addStudent(address student) public override whileNotEnded {
       super.addStudent(student);
    }

    /**
     * @dev Removes a student from the course while course is running.
     */
    function removeStudent(address student) public override whileNotEnded {
        super.removeStudent(student);
    }

    /**
     * @dev Gives up the course while course is running.
     */
    function renounceCourse() public override whileNotEnded {
        super.renounceCourse();
    }

    // TODO: add tests for extenting time
    function extendTime(uint256 NewEndingTime) public onlyOwner {
        _extendTime(NewEndingTime);
    }

    /**
     * @dev issue a credential proof for enrolled students while course is running.
     */
    function registerCredential(address student, bytes32 digest)
        public
        override
        whileNotEnded
    {
        super.registerCredential(student, digest);
    }

    function aggregateCredentials(address student)
        public
        override
        returns (bytes32)
    {
        require(hasEnded(), "TimedCourse: course not ended yet");
        return super.aggregateCredentials(student);
    }
}