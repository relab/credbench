pragma solidity >=0.5.13;

import "./NotaryInterface.sol";
import "./Owners.sol";
import "@openzeppelin/contracts/math/SafeMath.sol";
// import "@openzeppelin/contracts/cryptography/ECDSA.sol";

// TODO: how to manage key changes? e.g. a student that lost his previous key. Reissue the certificates may not work, since the time ordering, thus a possible solution is the contract to store a key update information for the subject, or something like that.

// TODO: Make it a library

/**
 * @title Notary's contract ensures that verifiable credentials are correctly
  * issued by untrusted issuers, discouraging fraudulent processes by
  * establishing a casual order between the certificates.
  * This contract consider on-chain signatures verification.
  * TODO: Implement using EIP712:
  * https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md
*/
contract Notary is NotaryInterface, Owners {
    using SafeMath for uint;

    /**
     * @dev CredentialProof represents an on-chain proof that a
     * verifiable credential was created and signed by an issuer.
     */
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

    /**
     * @dev RevocationProof represents an on-chain proof that a
     * verifiable credential was revoked by an issuer.
     */
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
    
    // Logged when a credential is created by an issuer
    event CredentialCreated(bytes32 indexed digest, address indexed subject, address indexed issuer, bytes32 previousDigest, uint insertedBlock);

    /**
     * @dev Constructor creates an Owners contract
     */
    constructor (address[] memory owners, uint quorum) public Owners(owners, quorum) {
        // solhint-disable-previous-line no-empty-blocks
    }

    modifier notRevoked(bytes32 digest) {
        require(
            !wasRevoked(digest),
            "Notary: this credential was already revoked"
        );
        _;
    }

    /**
     * @dev verify if a credential proof was revoked
     * @return true if a revocation exists, false otherwise.
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
    ) public onlyOwner notRevoked(digest) {
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
                // Ensure that a previous certificate happens before the new one.
                // solhint-disable-next-line expression-indent
                assert(c.insertedBlock < block.number);
                // solhint-disable-next-line not-rely-on-time, expression-indent
                assert(c.blockTimestamp < block.timestamp);
            }
            // TODO: assert the expect value here
            // previousDigest will be zero if didn't exists?
            issuedCredentials[digest] = CredentialProof(
                1,
                false,
                block.number,
                block.timestamp, // solhint-disable-line not-rely-on-time
                _nonce[subject],
                msg.sender,
                subject,
                digest,
                previousDigest
            );
            ++_nonce[subject];
            digestsBySubject[subject].push(digest); // append subject's credential hash
            emit CredentialCreated(digest, subject, msg.sender, previousDigest, block.number);
        } else {
            require(issuedCredentials[digest].subject == subject, "Notary: credential already issued for other subject");
            // Register "signature"
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
    function requestProof (bytes32 digest) public notRevoked(digest) {
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
        proof.subjectSigned = true; // All parties signed
        emit CredentialIssued(digest, proof.subject, proof.issuer, proof.previousDigest, proof.insertedBlock);
    }

    /**
     * @dev revoke a credential proof
     */
    function revoke (bytes32 digest) public onlyOwner notRevoked(digest) {
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

    /**
     * @dev aggregate the digests of a given subject
     */
    function aggregate (address subject) public view returns (bytes32) {
        bytes32[] memory digests = digestsBySubject[subject];
        // TODO: array index validation
        require(digests.length > 0, "Notary: there is no certificate for the given subject");
        assert(certified(digests[0])); // certificate must be signed by all parties
        bytes32 computedHash = digests[0];

        for (uint256 i = 1; i < digests.length; i++) {
            assert(certified(digests[i])); // all subject's certificates must be signed by all parties
            computedHash = keccak256(abi.encodePacked(computedHash, digests[i]));
        }
        return computedHash;
        // TODO: after aggregation, the digest can potentially be erased and a root credential can be create as replacement.
    }
}