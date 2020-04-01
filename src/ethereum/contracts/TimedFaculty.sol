pragma solidity >=0.5.13 <0.7.0;

import "bbchain-contracts/contracts/AccountableIssuer.sol";
import "./TimedCourse.sol";

// TODO: contract to manage employees (addresses) - enhancement
/**
 * @title Academic TimedFaculty
 * This contract manage courses contracts.
 */
contract TimedFaculty is AccountableIssuer {
    // Map courses by semester
    mapping(bytes32 => address[]) public coursesBySemester;

    event CourseCreated(
        address indexed createdBy,
        bytes32 indexed semester,
        address indexed courseAddress,
        address[] teachers,
        uint256 quorum
    );

    constructor(address[] memory owners, uint256 quorum)
        public
        AccountableIssuer(owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }

    function createCourse(
        bytes32 semester,
        address[] memory teachers,
        uint256 quorum,
        uint256 beginTimestamp,
        uint256 endTimestamp
    ) public onlyOwner returns (address) {
        TimedCourse course = new TimedCourse(
            teachers,
            quorum,
            beginTimestamp,
            endTimestamp
        );
        coursesBySemester[semester].push(address(course));
        addIssuer(address(course));
        emit CourseCreated(
            msg.sender,
            semester,
            address(course),
            teachers,
            quorum
        );
        return address(course);
    }
}
