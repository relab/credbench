pragma solidity >=0.5.13 <0.7.0;
pragma experimental ABIEncoderV2;

import "bbchain-contracts/contracts/Issuer.sol";
import "bbchain-contracts/contracts/Timed.sol";

/**
 * @title Academic Course
 */
contract Course is Timed, Issuer {
    // The teacher and the evaluator are owners of the contract
    mapping(address => bool) public enrolledStudents;

    event StudentAdded(address indexed student, address indexed requester);
    event StudentRemoved(address indexed student, address indexed requester);

    modifier registeredStudent(address student) {
        require(isEnrolled(student), "Course: student not registered");
        _;
    }

    /**
    * @dev Constructor creates a Issuer contract
    */
    constructor(
        address[] memory owners,
        uint256 quorum,
        uint256 startingTime,
        uint256 endingTime
    ) public Issuer(owners, quorum) Timed(startingTime, endingTime) {
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
    function _removeStudent(address student)
        internal
        whileNotEnded
        registeredStudent(student)
    {
        enrolledStudents[student] = false;
        emit StudentRemoved(student, msg.sender);
    }

    /**
     * @dev Removes a student from the course
     */
    function removeStudent(address student) public onlyOwner {
        _removeStudent(student);
    }

    /**
     * @dev Gives up the course
     */
    function renounceCourse() public {
        _removeStudent(msg.sender);
    }

    // TODO: add tests for extenting time
    function extendTime(uint256 NewEndingTime) public onlyOwner {
        _extendTime(NewEndingTime);
    }

    /**
     * @dev issue a credential proof for enrolled students
     */
    function registerCredential(address student, bytes32 digest)
        public
        override
        onlyOwner
        whileNotEnded
        registeredStudent(student)
    {
        super.registerCredential(student, digest);
    }

    // TODO: only allow onwer to call the aggregation? If so, the faculty contract will not be able to call the method, and the teacher will need to call it
    function aggregateCredentials(address student)
        public
        override
        registeredStudent(student)
        returns (bytes32)
    {
        require(hasEnded(), "Course: course not ended yet");
        return super.aggregateCredentials(student);
    }
}