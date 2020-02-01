pragma solidity >=0.5.13;
pragma experimental ABIEncoderV2;

// import "./CertificateSum.sol";
import "./Notary.sol";
import "./Course.sol";

// TODO: contract to manage employees (addresses) - enhancement
contract Faculty is Notary {
    // Map courses by semester
    mapping(bytes32 => address[]) public coursesBySemester;

    // Map of all courses
    mapping(address => bool) public courses;

    // Map diplomas by students
    mapping(address => bytes32) public diplomas;

    event CourseCreated(
        bytes32 indexed semester,
        address indexed courseAddress,
        address[] teachers,
        uint256 quorum
    );

    constructor(address[] memory owners, uint256 quorum)
        public
        Notary(owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }

    function createCourse(
        bytes32 semester,
        address[] memory teachers,
        uint256 quorum,
        uint256 beginTimestamp,
        uint256 endTimestamp
    ) public {
        // TODO: reause course contract instead of create a new one
        Course course = new Course(
            teachers,
            quorum,
            beginTimestamp,
            endTimestamp
        );
        coursesBySemester[semester].push(address(course));
        courses[address(course)] = true;
        emit CourseCreated(semester, address(course), teachers, quorum);
    }

    function issueDiploma(address subject, address[] memory courses_addresses)
        public
        onlyOwner
    {
        bytes32[] memory digests = new bytes32[](courses_addresses.length);
        for (uint256 i = 0; i < courses_addresses.length; i++) {
            require(
                courses[address(courses_addresses[i])],
                "Faculty: course must exists"
            ); // all courses should exists
            Course course = Course(address(courses_addresses[i]));

            // assert(course.courseCertificate(subject).insertedBlock != 0);
            Course.CourseProof memory proof = course.courseCertificate(subject);
            digests[i] = proof.digest;
        }
        require(digests.length > 0, "Faculty: Courses certificates not found");
        bytes32 diploma = keccak256(abi.encode(digests));
        // Two alternatives:
        // 1) call CertificateSum library and get the resulted proof.
        super.issue(subject, diploma);
        // 2) or, publish the proof as a notary
    }
}
