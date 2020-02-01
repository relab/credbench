pragma solidity >=0.5.13;
pragma experimental ABIEncoderV2;

import "./Notary.sol";
import "./TimedNotary.sol";

/**
 * @title Academic Course
 */
contract Course is TimedNotary {
    struct CourseProof {
        uint256 insertedBlock;
        uint256 blockTimestamp;
        address issuer;
        address subject;
        bytes32 digest;
        bytes32[] exams;
    }

    mapping(address => CourseProof) _courseCertificate;

    // The teacher and the evaluator are owners of the contract
    mapping(address => bool) public enrolledStudents;

    event StudentAdded(address indexed student, address indexed requester);
    event StudentRemoved(address indexed student, address indexed requester);

    event CourseCertificateCreated(
        address indexed subject,
        address indexed issuer,
        bytes32 indexed digest,
        bytes32[] certificates,
        uint256 blockNumber
    );

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

    function courseCertificate(address student)
        public
        view
        returns (CourseProof memory)
    {
        return _courseCertificate[student];
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

    /**
     * @dev issue a credential proof for enrolled students
     */
    function issueExam(address student, bytes32 digest)
        public
        onlyOwner
        whileNotEnded
    {
        require(enrolledStudents[student], "Course: student not registered");
        super.issue(student, digest);
    }

    // TODO: add tests for extenting time
    function extendTime(uint256 NewEndingTime) public {
        _extendTime(NewEndingTime);
    }

    function issueCourseCertificateFor(address subject) public onlyOwner {
        require(hasEnded(), "Course: course not ended yet");
        bytes32 courseDigest = aggregate(subject);
        _courseCertificate[subject] = CourseProof(
            block.number,
            block.timestamp,
            msg.sender,
            subject,
            courseDigest,
            digestsBySubject[subject]
        );
        emit CourseCertificateCreated(
            subject,
            msg.sender,
            courseDigest,
            digestsBySubject[subject],
            block.number
        );
    }
}
