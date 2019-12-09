pragma solidity >=0.5.8;

import "./Owners.sol";
import "@openzeppelin/contracts/math/SafeMath.sol";
// import "@openzeppelin/contracts/cryptography/ECDSA.sol";

// TODO: how to manage key changes? e.g. a student that lost his previous key. Reissue the certificates may not work, since the time ordering, thus a possible solution is the contract to store a key update information for the subject, or something like that.

// TODO: Make it a library

/**
 * @title Notary's contract
*/
contract Notary is Owners {
    using SafeMath for uint;

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
    mapping(address => uint) private _nonce;

    // Maps credential digests by subjects
    mapping(address => bytes32[]) public digestsBySubject;

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
    constructor (address[] memory owners, uint quorum) public Owners(owners, quorum) {
    }

    /**
     * @dev verify if a credential proof was revoked
     * @return the block number of revocation or 0 if not revoked
     */
    function wasRevoked (bytes32 digest) public view returns(bool) {
        return revokedCredentials[digest].revokedBlock != 0;
    }

    /**
     * @dev issue a credential proof ensuring an append-only property
     */
    function issue (
        address subject,
        bytes32 digest
    ) public onlyOwner {
        require(
            !ownersSigned[digest][msg.sender],
            "Notary: sender already signed"
        );
        require(
            !isOwner[subject],
            "Notary: subject cannot be the issuer"
        );
        if (issuedCredentials[digest].insertedBlock == 0) {
            // Creation
            uint lastNonce;
            bytes32 previousDigest;
            if (_nonce[subject] == 0) {
                lastNonce = _nonce[subject];
                previousDigest = bytes32(0);
            } else {
                assert(_nonce[subject] > 0);
                lastNonce = _nonce[subject] - 1;
                assert(digestsBySubject[subject].length > 0);
                previousDigest = digestsBySubject[subject][lastNonce];
                CredentialProof memory c = issuedCredentials[previousDigest];
                require(c.subjectSigned, "Notary: previous credential must be signed before issue a new one");
                // Assert time constraints
                // TODO: multiply by credential time factor that will be defined in the Notary constructor
                // solhint-disable-next-line expression-indent
                assert(c.insertedBlock < block.number/* * TimeFactor */);
                // solhint-disable-next-line not-rely-on-time, expression-indent
                assert(c.blockTimestamp < block.timestamp/* * TimeFactor */);
            }
            // TODO: assert the expect value here
            // previousDigest will be zero with didn't exists?
            issuedCredentials[digest] = CredentialProof(
                1,
                false,
                block.number,
                block.timestamp, // solhint-disable-line not-rely-on-time
                _nonce[subject],
                msg.sender,
                subject,
                digest, // Do we aggregate the hashes instead of just save?
                previousDigest
            );
            ++_nonce[subject];
            digestsBySubject[subject].push(digest); // append subject credential
        } else {
            // Signing
            ++issuedCredentials[digest].signed;
        }
        ownersSigned[digest][msg.sender] = true;
    }

    /**
     * @dev Verify if a digest was already certified (i.e. signed by all parties)
     */
    function certified (bytes32 digest) public view returns(bool) {
        return issuedCredentials[digest].subjectSigned;
    }

    /**
     * @dev request the emission of a quorum signed credential proof
     */
    function requestProof (bytes32 digest) public {
        CredentialProof storage proof = issuedCredentials[digest];
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
        emit CredentialIssued(digest, proof.subject, proof.issuer, proof.previousDigest, proof.insertedBlock);
    }

    /**
     * @dev revoke a credential proof
     */
    function revoke (bytes32 digest) public onlyOwner {
        require(
            issuedCredentials[digest].insertedBlock != 0,
            "Notary: no credential proof found"
        );
        address subject = issuedCredentials[digest].subject;
        revokedCredentials[digest] = RevocationProof(
            msg.sender,
            subject,
            block.number
        );
        delete issuedCredentials[digest];
        emit CredentialRevoked(digest, subject, msg.sender, block.number);
    }
}