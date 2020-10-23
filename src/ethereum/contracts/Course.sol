// SPDX-License-Identifier: MIT
pragma solidity >=0.6.0 <0.8.0;
pragma experimental ABIEncoderV2;

import "ct-eth/contracts/node/Node.sol";

/**
 * @title Academic Course
 * The teacher and the evaluator are owners of the contract
 */
contract Course is Node {
    address[] internal _students;
    mapping(address => Student) internal _enrolledStudents;

    struct Student {
        uint256 index;
        bool enrolled;
    }

    event StudentAdded(address indexed student, address indexed createdBy);
    event StudentRemoved(address indexed student, address indexed createdBy);

    modifier registeredStudent(address student) {
        require(isEnrolled(student), "Course/student not registered");
        _;
    }

    /**
    * @dev Constructor creates a Course contract
    */
    constructor(address[] memory owners,uint8 quorum)
        Node(Role.Leaf, owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }

    function getStudents() public view returns (address[] memory) {
        return _students;
    }

    /**
     * @dev Check if a student is enrolled in the course
     * @param student the address of the student to be checked.
     * @return Whether the student is enrolled in the course
     */
    function isEnrolled(address student) public view returns (bool) {
        require(student != address(0), "Course/zero address given");
        return _enrolledStudents[student].enrolled;
    }

    /**
     * @dev Adds a student to the course
     * @param student the address of the student to be added.
     */
    function addStudent(address student) public virtual onlyOwner {
        require(
            !isEnrolled(student),
            "Course/student already registered"
        );
        require(!isOwner(student), "Course/student cannot be owner");
        uint256 index = _students.length;
        _enrolledStudents[student] = Student(index, true);
        _students.push(student);
        emit StudentAdded(student, msg.sender);
    }

    /**
     * @dev Remove a student. The order of the list of students
     * is not maintained.
     * @param student the address of the student to be removed.
     */
    function _removeStudent(address student)
        private
        registeredStudent(student)
    {
        require(_students.length > 0);

        uint256 index = _enrolledStudents[student].index;
        delete _enrolledStudents[student];

        address swapped = _students[_students.length-1];
        _students[index] = swapped;
        _enrolledStudents[swapped].index = index;
        _students.pop();

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

    function registerExam(
        address student,
        bytes32 digest
    )
        public
        onlyOwner
        registeredStudent(student)
    {
        super.registerCredential(student, digest, new address[](0));
    }
}