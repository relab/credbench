pragma solidity >=0.5.13;
// pragma experimental ABIEncoderV2;

import "./AccountableIssuer.sol";
import "./Course.sol";

// TODO: contract to manage employees (addresses) - enhancement
contract Faculty is AccountableIssuer {
    // Map courses by semester
    mapping(bytes32 => address[]) public coursesBySemester;

    event CourseCreated(
        bytes32 indexed semester,
        address indexed courseAddress,
        address[] teachers,
        uint256 quorum
    );

    constructor(address[] memory owners, uint256 quorum)
        public
        Issuer(owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }

    //method from issuer
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
        addIssuer(address(course));
        emit CourseCreated(semester, address(course), teachers, quorum);
    }

    // the diploma should be a hash of all student course certificates, that is a hash of all student exams(even the bad ones - this is why merkle tree would be good, so the student could choose what grade to present as the certificate (i.e. json) and still be a valid diploma hash).
    // Currently the diploma is build by hashing all digests in sequence following the given course contrats order, which can be wrong and produce different hashes. Should respect the timestamp order of certificates

}
