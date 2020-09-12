// SPDX-License-Identifier: MIT
pragma solidity >=0.6.0 <0.8.0;
pragma experimental ABIEncoderV2;

import "ct-eth/contracts/Issuer.sol";

/**
 * @title Academic Course
 */
contract Course is Issuer {
    // The teacher and the evaluator are owners of the contract
    mapping(address => bool) public enrolledStudents;

    event StudentAdded(address indexed student, address indexed requester);
    event StudentRemoved(address indexed student, address indexed requester);

    modifier registeredStudent(address student) {
        require(isEnrolled(student), "Course: student not registered");
        _;
    }

    /**
    * @dev Constructor creates a Course contract
    */
    constructor(
        address[] memory owners,
        uint256 quorum
    ) Issuer(owners, quorum, true) {
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
    function addStudent(address student) public virtual onlyOwner {
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
        registeredStudent(student)
    {
        enrolledStudents[student] = false;
        emit StudentRemoved(student, msg.sender);
    }

    /**
     * @dev Removes a student from the course
     */
    function removeStudent(address student) public virtual onlyOwner {
        _removeStudent(student);
    }

    /**
     * @dev Gives up the course
     */
    function renounceCourse() public virtual {
        _removeStudent(msg.sender);
    }

    /**
     * @dev issue a credential proof for enrolled students
     */
    function registerCredential(address student, bytes32 digest)
        public
        virtual
        override
        onlyOwner
        registeredStudent(student)
    {
        super.registerCredential(student, digest);
    }

    function aggregateCredentials(address student)
        public
        virtual
        override
        onlyOwner
        registeredStudent(student)
        returns (bytes32)
    {
        return super.aggregateCredentials(student);
    }
}