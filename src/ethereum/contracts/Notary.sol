pragma solidity >=0.5.13;
pragma experimental ABIEncoderV2;

import "./NotaryInterface.sol";
import "./Owners.sol";
// import "@openzeppelin/contracts/math/SafeMath.sol";

// TODO: how to manage key changes? e.g. a student that lost his previous key. Reissue the certificates may not work, since the time ordering, thus a possible solution is the contract to store a key update information for the subject, or something like that.

/**
 * @title Notary's contract ensures that verifiable credentials are correctly
  * issued by untrusted issuers, discouraging fraudulent processes by
  * establishing a casual order between the certificates.
*/
abstract contract Notary is NotaryInterface, Owners {
    // using SafeMath for uint256;

    /**
     * @dev CredentialProof represents an on-chain proof that a
     * verifiable credential was created and signed by an issuer.
     */
    struct CredentialProof {
        uint256 signed; // Amount of owners who signed
        bool subjectSigned; // Whether the subject signed
        // FIXME: redundant time information. Decide on which one to use
        uint256 insertedBlock; // The block number of the proof creation
        uint256 blockTimestamp; // The block timestamp of the proof creation
        uint256 nonce;  // Increment-only counter of credentials of the same subject
        address issuer; // The issuer address of this proof
        address subject; // The entity address refered by a proof
        bytes32 digest; // The digest of the credential stored (e.g. Swarm/IPFS hash)
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
        uint256 revokedBlock; // The block number of the revocation (0 if not revoked)
        bytes32 reason; // digest of the reason of the revocation
    }
    // TODO: add key revocation

    // Incremental-only counter for issued credentials per subject
    mapping(address => uint256) public nonce;

    // Maps credential digests by subjects
    mapping(address => bytes32[]) private _digestsBySubject;

    // Maps issued credential proof by document digest
    mapping(bytes32 => CredentialProof) public issuedCredentials;

    // Maps document digest to revoked proof
    mapping(bytes32 => RevocationProof) public revokedCredentials;

    // Map digest to owners that already signed it
    mapping(bytes32 => mapping(address => bool)) public ownersSigned;

    /**
     * @dev Constructor creates an Owners contract
     */
    constructor(address[] memory owners, uint256 quorum)
        public
        Owners(owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }

    modifier notRevoked(bytes32 digest) {
        require(
            !wasRevoked(digest),
            "Notary: this credential was already revoked"
        );
        _;
    }

    function digestsBySubject(address subject) public view returns(bytes32[] memory) {
        return _digestsBySubject[subject];
    }

    /**
     * @dev verify if a credential proof was revoked
     * @return true if a revocation exists, false otherwise.
     */
    function wasRevoked(bytes32 digest) public view override returns (bool) {
        return revokedCredentials[digest].revokedBlock != 0;
    }

    function _issue(address subject, bytes32 digest)
        internal 
        notRevoked(digest)
    {
        require(
            !ownersSigned[digest][msg.sender],
            "Notary: sender already signed"
        );
        require(!isOwner[subject], "Notary: subject cannot be the issuer");
        if (issuedCredentials[digest].insertedBlock == 0) {
            // Creation
            uint256 lastNonce;
            if (nonce[subject] == 0) {
                lastNonce = nonce[subject];
            } else {
                assert(nonce[subject] > 0);
                lastNonce = nonce[subject] - 1;
                assert(_digestsBySubject[subject].length > 0);
                bytes32 previousDigest = _digestsBySubject[subject][lastNonce];
                CredentialProof memory c = issuedCredentials[previousDigest];
                require(
                    c.subjectSigned,
                    "Notary: previous credential must be signed before issue a new one"
                );
                // Assert time constraints
                // Ensure that a previous certificate happens before the new one.
                // solhint-disable-next-line expression-indent
                assert(c.insertedBlock < block.number);
                // solhint-disable-next-line not-rely-on-time, expression-indent
                assert(c.blockTimestamp < block.timestamp);
            }
            issuedCredentials[digest] = CredentialProof(
                1,
                false,
                block.number,
                block.timestamp, // solhint-disable-line not-rely-on-time
                nonce[subject],
                msg.sender,
                subject,
                digest
            );
            ++nonce[subject];
            _digestsBySubject[subject].push(digest); // append subject's credential hash
            emit CredentialIssued(
                digest,
                subject,
                msg.sender,
                block.number
            );
        } else {
            require(
                issuedCredentials[digest].subject == subject,
                "Notary: credential already issued for other subject"
            );
            // Register sign action
            ++issuedCredentials[digest].signed;
        }
        ownersSigned[digest][msg.sender] = true;
    }

    /**
     * @dev issue a credential proof ensuring an append-only property
     */
    function issue(address subject, bytes32 digest)
        public
        override
        onlyOwner
    {
        _issue(subject, digest);
        emit CredentialSigned(msg.sender, digest, block.number);
    }

    /**
     * @dev Verify if a digest was already certified (i.e. signed by all parties)
     */
    function certified(bytes32 digest) public view override returns (bool) {
        return issuedCredentials[digest].subjectSigned;
    }

    /**
     * @dev request the emission of a quorum signed credential proof
     */
    function confirmProof(bytes32 digest) public override notRevoked(digest) {
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
        emit CredentialSigned(msg.sender, digest, block.number);
    }

    /**
     * @dev revoke a credential proof
     */
    function revoke(bytes32 digest, bytes32 reason)
        public
        override
        onlyOwner
        notRevoked(digest)
    {
        require(
            issuedCredentials[digest].insertedBlock != 0,
            "Notary: no credential proof found"
        );
        address subject = issuedCredentials[digest].subject;
        assert(_digestsBySubject[subject].length > 0);
        revokedCredentials[digest] = RevocationProof(
            msg.sender,
            subject,
            block.number,
            reason
        );
        delete issuedCredentials[digest];
        emit CredentialRevoked(
            digest,
            subject,
            msg.sender,
            block.number,
            reason
        );
    }

    /**
     * @dev aggregate the digests of a given subject
     */
    function aggregate(address subject) public override virtual returns (bytes32, uint256, uint256) {
        bytes32[] memory digests = _digestsBySubject[subject];
        require(
            digests.length > 0,
            "Notary: there is no certificate for the given subject"
        );
        // TODO: ignore the revoke credentials in the aggregation
        for (uint256 i = 0; i < digests.length; i++) {
            require(certified(digests[i]), "Notary: impossible to aggregate. There are unsigned certificates"); //&& !wasRevoked(digests[i]));
            // all subject's certificates must be signed by all parties and should be valid
        }
        uint256 firstBlock = issuedCredentials[digests[0]].insertedBlock;
        uint256 lastBlock = issuedCredentials[digests[nonce[subject]]].insertedBlock;
        bytes32 digest = keccak256(abi.encode(digests));
        emit AggregatedCredential(msg.sender, subject, digest, firstBlock, lastBlock);
        return (digest, firstBlock, lastBlock);
    }
}
