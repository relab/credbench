pragma solidity >=0.5.13;
pragma experimental ABIEncoderV2;

// import "./CertificateSum.sol";
import "./Notary.sol";
import "./Course.sol";

// TODO: contract to manage employees (addresses) - enhancement
contract Faculty is Notary {
    // Map courses by semester
    mapping(bytes32 => address[]) public coursesBySemester;

    // Map of all courses contracts
    mapping(address => bool) public courses;

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

    // the diploma should be a hash of all student course certificates, that is a hash of all student exams(even the bad ones - this is why merkle tree would be good, so the student could choose what grade to present as the certificate (i.e. json) and still be a valid diploma hash).
    // Currently the diploma is build by hashing all digests in sequence following the given course contrats order, which can be wrong and produce different hashes. Should respect the timestamp order of certificates
    // TODO: make a generic issue by aggregation
    function issueDiploma(
        address subject,
        bytes32 digest,
        bytes32 diplomaRoot,
        address[] memory courses_addresses // TODO: add period check
    ) public onlyOwner {
        // TODO: get all courses of a student
        bytes32[] memory digests = new bytes32[](courses_addresses.length + 1);
        require(courses_addresses.length > 0, "Faculty: No courses were given");
        uint256 i = 0;
        address courseAddr;
        for (; i < courses_addresses.length; i++) {
            courseAddr = address(courses_addresses[i]);
            //collect course certificates
            require(courses[courseAddr], "Faculty: course doesn't registered");
            Course course = Course(courseAddr);
            digests[i] = course.aggregate(subject);
        }
        // Add diploma
        digests[i] = digest;
        bytes32 diploma = keccak256(abi.encode(digests));
        assert(diploma == diplomaRoot);
        // TODO: store list of course addresses used to generate the proof
        super.issue(subject, digest); // create a diploma
    }
}
