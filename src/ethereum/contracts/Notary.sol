pragma solidity >=0.5.7 <0.7.0;

import './Owners.sol';

// TODO: how to manage key changes? e.g. a student that lost his previous key. Reissue the certificates may not work, since the time ordering, thus a possible solution is the contract to store a key update information for the subject, or something like that.

// TODO: Make it a library

/**
 * @title Notary's contract
*/
contract Notary is Owners {

    struct CredentialProof {
        uint signed;            // Amount of owners who signed
        bool subjectSigned;     // Whether the subject signed
        uint insertedBlock;     // The block number of the proof creation
        uint blockTimestamp;    // The block timestamp of the proof creation
        uint nonce;
        address issuer;         // The issuer address of this proof
        address subject;        // The entity address refered by a proof
        bytes32 digest;         // The digest of the credential stored (e.g. Swarm hash)
        bytes32 previousDigest;  // The hash of the previous certificate
        // TODO: add "bytes signature" field to allow external signatures and on-chain verification
        // TODO: add "uint256 signatureType" to inform what type of signature was used
        // TODO: add "string uri" field to identify the storage type (bzz, ipfs)
    }

    struct RevocationProof {
        address issuer;
        address subject;
        uint revokedBlock;  // The block number of the revocation (0 if not revoked)
    }
    // TODO: add key revocation

    // Incremental-only counter for issued credentials per subject
    mapping(address => uint) private nonce;

    // Maps credential digests by subjects
    mapping(address => bytes32[]) public digestsBySubject; //TODO: keep updated with the last hash manifest instead of a list

    // Maps issued credential proof by document digest
    mapping(bytes32 => CredentialProof) public issuedCredentials;

    // Maps document digest to revoked proof
    mapping(bytes32 => RevocationProof) public revokedCredentials;

    // Map digest to owners that already signed it
    mapping (bytes32 => mapping (address => bool)) public ownersSigned;
    
    // Logged when a credential is issued and signed by all parties (owners + subject).
    event CredentialIssued(bytes32 indexed digest, address indexed subject, address indexed issuer, bytes32 previousDigest, uint insertedBlock);

    // Logged when a credential is revoked by some owner.
    event CredentialRevoked(bytes32 indexed digest, address indexed subject, address indexed issuer, uint revokedBlock);

    /**
     * @dev Constructor creates an Owners contract
     */
    constructor (address[] memory _owners, uint _quorum) Owners(_owners, _quorum) public {}

    /**
     * @dev verify if a credential proof was revoked
     * @return the block number of revocation or 0 if not revoked
     */
    function wasRevoked(bytes32 _digest) public view returns(bool) {
        return revokedCredentials[_digest].revokedBlock != 0;
    }

    /**
     * @dev issue a credential proof ensuring an append-only property
     */
    function issue(
        address _subject,
        bytes32 _digest
    ) public onlyOwner {
        require(
            !ownersSigned[_digest][msg.sender],
            "Notary: sender already signed"
        );
        require(
            !isOwner[_subject],
            "Notary: subject cannot be the issuer"
        );
        if (issuedCredentials[_digest].insertedBlock == 0) {
            // Creation
            uint lastNonce;
            bytes32 previousDigest;
            if (nonce[_subject] == 0) {
                lastNonce = nonce[_subject];
                previousDigest = bytes32(0);
            } else {
                assert (nonce[_subject] > 0);
                lastNonce = nonce[_subject] - 1;
                assert (digestsBySubject[_subject].length > 0);
                previousDigest = digestsBySubject[_subject][lastNonce];
                CredentialProof memory c = issuedCredentials[previousDigest];
                require (c.subjectSigned, "Notary: previous credential must be signed before issue a new one");
                // Assert time constraints
                // TODO: multiply by credential time factor that will be defined in the Notary constructor
                assert(c.insertedBlock < block.number /* * TimeFactor */);
                assert(c.blockTimestamp < block.timestamp /* * TimeFactor */);
            }
            // TODO: assert the expect value here
            // previousDigest will be zero with didn't exists?
            issuedCredentials[_digest] = CredentialProof(
                1,
                false,
                block.number,
                block.timestamp,
                nonce[_subject],
                msg.sender,
                _subject,
                _digest, // Do we aggregate the hashes instead of just save?
                previousDigest
            );
            ++nonce[_subject];
            digestsBySubject[_subject].push(_digest); // append subject credential
        } else {
            // Signing
            ++issuedCredentials[_digest].signed;
        }
        ownersSigned[_digest][msg.sender] = true;
    }

    /**
     * @dev Verify if a digest was already certified (i.e. signed by all parties)
     */
    function certified(bytes32 _digest) public view returns(bool) {
        return issuedCredentials[_digest].subjectSigned;
    }

    /**
     * @dev request the emission of a quorum signed credential proof
     */
    function requestProof(bytes32 _digest) public {
        CredentialProof storage proof = issuedCredentials[_digest];
        require(
            proof.subject == msg.sender,
            "Notary: subject is not related with this credential"
        );
        require(
            !proof.subjectSigned,
            "Notary: subject already signed this credential"
        );
        require(
            proof.signed >= quorum,
            "Notary: not sufficient quorum of signatures"
        );
        proof.subjectSigned = true;
        emit CredentialIssued(_digest, proof.subject, proof.issuer, proof.previousDigest, proof.insertedBlock);
    }

    /**
     * @dev revoke a credential proof
     */
    function revoke(bytes32 _digest) public onlyOwner {
        require(
            issuedCredentials[_digest].insertedBlock != 0,
            "Notary: no credential proof found"
        );
        address subject = issuedCredentials[_digest].subject;
        revokedCredentials[_digest] = RevocationProof(
            msg.sender,
            subject,
            block.number
        );
        delete issuedCredentials[_digest];
        emit CredentialRevoked(_digest, subject, msg.sender, block.number);
    }
}