pragma solidity >=0.5.13;
pragma experimental ABIEncoderV2;

import "./Notary.sol";
import "./TimedNotary.sol";

/**
 * @title Academic Course
 */
contract Course is TimedNotary, Notary {
    // The teacher and the evaluator are owners of the contract
    mapping(address => bool) public enrolledStudents;

    event StudentAdded(address indexed student, address indexed requester);
    event StudentRemoved(address indexed student, address indexed requester);

    modifier registeredStudent(address student) {
        require(enrolledStudents[student], "Course: student not registered");
        _;
    }

    /**
    * @dev Constructor creates a Notary contract
    */
    constructor(
        address[] memory owners,
        uint256 quorum,
        uint256 startingTime,
        uint256 endingTime
    ) public Notary(owners, quorum) TimedNotary(startingTime, endingTime) {
        // solhint-disable-previous-line no-empty-blocks
    }

    /**
     * @dev Check if a student is enrolled in the course
     * @param student the address of the student to be checked.
     * @return Whether the student is enrolled in the course
     */
    function isEnrolled(address student) public view returns (bool) {
        require(student != address(0), "Course: student is the zero address");
        return enrolledStudents[student];
    }

    /**
     * @dev Adds a student to the course
     * @param student the address of the student to be added.
     */
    function addStudent(address student) public onlyOwner whileNotEnded {
        require(
            !isEnrolled(student),
            "Course: student already registered in this course"
        );
        enrolledStudents[student] = true;
        emit StudentAdded(student, msg.sender);
    }

    /**
     * @dev Register a student course removal
     * @param student the address of the student to be removed.
     */
    function _removeStudent(address student) internal {
        require(
            isEnrolled(student),
            "Course: student does not registered in this course"
        );
        enrolledStudents[student] = false;
        emit StudentRemoved(student, msg.sender);
    }

    /**
     * @dev Removes a student from the course
     */
    function removeStudent(address student) public onlyOwner whileNotEnded {
        _removeStudent(student);
    }

    /**
     * @dev Gives up the course
     */
    function renounceCourse() public onlyAfterStart whileNotEnded {
        _removeStudent(msg.sender);
    }

    // TODO: add tests for extenting time
    function extendTime(uint256 NewEndingTime) public {
        _extendTime(NewEndingTime);
    }

    /**
     * @dev issue a credential proof for enrolled students
     */
    function issue(address student, bytes32 digest)
        public
        override
        onlyOwner
        whileNotEnded
        registeredStudent(student)
    {
        super.issue(student, digest);
    }

    // TODO: only allow onwer to call the aggregation, then the faculty contract will not be able to call the method, but the teacher will need to call it
    function aggregate(address student)
        public
        override
        onlyAfterStart
        registeredStudent(student)
        returns (bytes32, uint256, uint256)
    {
        require(hasEnded(), "Course: course not ended yet");
        return super.aggregate(student);
    }
}
