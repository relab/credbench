pragma solidity >=0.5.7 <0.7.0;

import "./Notary.sol";

/**
 * @title Academic Course
 */
contract Course is Notary {
    // The teacher and the evaluator are owners of the contract
    mapping(address => bool) public enrolled_students;

    event StudentAdded(address indexed student, address indexed requester);
    event StudentRemoved(address indexed student, address indexed requester);

    /**
    * @dev Constructor creates a Notary contract
    */
    constructor (address[] memory _owners, uint _quorum) Notary(_owners, _quorum) public {}

    /**
     * @dev Check if a student is enrolled in the course
     * @param _student the address of the student to be checked.
     * @return bool
     */
    function isEnrolled(address _student) public view returns (bool) {
        require(_student != address(0), "Course: student is the zero address");
        return enrolled_students[_student];
    }

    /**
     * @dev Adds a student to the course
     * @param _student the address of the student to be added.
     */
    function addStudent(address _student) public onlyOwner() {
        require(
            !isEnrolled(_student),
            "Course: student already registered in this course"
        );
        enrolled_students[_student] = true;
        emit StudentAdded(_student, msg.sender);
    }

    /**
     * @dev Register a student course removal
     * @param _student the address of the student to be removed.
     */
    function _removeStudent(address _student) internal {
        require(
            isEnrolled(_student),
            "Course: student does not registered in this course"
        );
        enrolled_students[_student] = false;
        emit StudentRemoved(_student, msg.sender);
    }

    /**
     * @dev Removes a student from the course
     */
    function removeStudent(address _student) public onlyOwner() {
        _removeStudent(_student);
    }

    /**
     * @dev Gives up the course
     */
    function renounceCourse() public {
        _removeStudent(msg.sender);
    }

    /**
     * @dev issue a credential proof for enrolled students
     */
    function issue(
        address _student,
        bytes32 _digest
    ) public onlyOwner {
        require(
            enrolled_students[_student],
            "Course: student not registered"
        );
        super.issue(_student, _digest);
    }
}