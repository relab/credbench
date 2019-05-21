pragma solidity >=0.5.7 <0.7.0;

import './Owners.sol';

/**
 * @title Notary's contract
*/
contract Notary is Owners {

    struct CredentialProof {
        uint signed;            // Amount of owners who signed
        bool subjectSigned;     // Whether the subject signed
        uint insertedBlock;     // The block number of the proof creation
        address issuer;         // The issuer address of this proof
        address subject;        // The entity address refered by a proof
        bytes32 digest;         // The digest of the credential stored (i.e. IPFSHash)
    }

    struct RevocationProof {
        address issuer;
        uint revokedBlock;  // The block number of the revocation (0 if not revoked)
    }

    // Maps document digest to issued credential proof
    mapping(bytes32 => CredentialProof) public issued;

    // Maps document digest to revoked proof
    mapping(bytes32 => RevocationProof) public revoked;

    // Map digest to owners that already signed it
    mapping (bytes32 => mapping (address => bool)) public ownersSigned;
    
    event CredentialIssued(bytes32 indexed digest, address indexed subject);
    event CredentialRevoked(bytes32 indexed digest, address indexed issuer, uint revokedBlock);

    /**
     * @dev Constructor create an Owners contract
     */
    constructor (address[] memory _owners, uint _quorum) Owners(_owners, _quorum) public {}

    /**
     * @dev verify if a credential proof was revoked
     * @return the block number of revocation or 0 if not revoked
     */
    function wasRevoked(bytes32 _digest) public view returns(bool) {
        return revoked[_digest].revokedBlock != 0;
    }

    /**
     * @dev issue a credential proof
     */
    function issue(
        address _subject,
        bytes32 _digest
    ) public onlyOwner {
        require(!ownersSigned[_digest][msg.sender], "Notary: sender already signed");
        if (issued[_digest].insertedBlock == 0) {
            // Creation
            issued[_digest] = CredentialProof(
                1,
                false,
                block.number,
                msg.sender,
                _subject,
                _digest
            );
        } else {
            // Signing
            ++issued[_digest].signed;
        }
        ownersSigned[_digest][msg.sender] = true;
    }

    /**
     * @dev Verify if a digest was already certified (i.e. signed by all parties)
     */
    function certified(bytes32 _digest) public view returns(bool) {
        return issued[_digest].subjectSigned;
    }

    /**
     * @dev request the emission of a quorum signed credential proof
     */
    function requestProof(bytes32 _digest) public {
        CredentialProof storage proof = issued[_digest];
        require(proof.subject == msg.sender, "Notary: subject isn't related with this credential");
        require(!proof.subjectSigned, "Notary: subject already signed this credential");
        require(proof.signed >= quorum, "Notary: not sufficient quorum of signatures");
        proof.subjectSigned = true;
        emit CredentialIssued(_digest, proof.subject);
    }

    /**
     * @dev revoke a credential proof
     */
    function revoke(bytes32 _digest) public onlyOwner {
        require(issued[_digest].insertedBlock != 0, "Notary: no credential proof found");
        revoked[_digest] = RevocationProof(
            msg.sender,
            block.number
        );
        delete issued[_digest];
        emit CredentialRevoked(_digest, msg.sender, block.number);
    }
}