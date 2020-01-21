pragma solidity >=0.5.13;

/**
 * @title Owners contract
*/
contract Owners {
    // Map of owners
    mapping (address => bool) public isOwner;
    address[] public owners;
    
    // The required number of owners to authorize actions
    uint public quorum;

    modifier onlyOwner {
        require(isOwner[msg.sender], "Owners: sender is not an owner");
        _;
    }

    /**
     * @dev Constructor
     * @param _owners is the array of all owners
     * @param _quorum is the required number of owners to perform actions
     */
    constructor (address[] memory _owners, uint _quorum) public {
        require(
            _owners.length > 0 && _owners.length < 256,
            "Owners: not enough owners"
        );
        require(
            _quorum > 0 && _quorum <= _owners.length,
            "Owners: quorum out of range"
        );
        for (uint i = 0; i < _owners.length; ++i) {
            // prevent duplicate and zero value address attack
            assert(!isOwner[_owners[i]] && _owners[i] != address(0));
            isOwner[_owners[i]] = true;
        }
        owners = _owners;
        quorum = _quorum;
    }

    /**
     * @return the length of the owners array
     */
    function ownersLength() public view returns (uint) {
        return owners.length;
    }
}