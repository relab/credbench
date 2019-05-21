pragma solidity >=0.5.7 <0.7.0;

/**
 * @title Owners contract
*/
contract Owners {
    // Map of owners
    mapping (address => bool) public owners;
    address[] public allOwners;
    
    // The required number of owners necessary to authorize actions
    uint public quorum;

    modifier onlyOwner {
        require(owners[msg.sender], "Owners: sender is not an owner");
        _;
    }

    /**
     * @dev Constructor
     * @param _owners is the array of all owners
     * @param _quorum is the required number of owners to perform actions
     */
    constructor (address[] memory _owners, uint _quorum) public {
        require(
            _owners.length > 0 && _owners.length <= 256,
            "Owners: not enough owners"
        );
        require(
            _quorum > 0 && _quorum <= _owners.length,
            "Owners: quorum out of range"
        );
        for (uint i = 0; i < _owners.length; ++i)
            owners[_owners[i]] = true;
        allOwners = _owners;
        quorum = _quorum;
    }

    /**
     * @return true if `msg.sender` is an owner
     */
    function isOwner() public view returns (bool) {
        return owners[msg.sender];
    }

    /**
     * @return the length of the owners array
     */
    function ownersLength() public view returns (uint) {
        return allOwners.length;
    }
}