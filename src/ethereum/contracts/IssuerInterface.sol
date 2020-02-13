pragma solidity >=0.5.13;

// TODO update the interface fields
interface IssuerInterface {
    // Logged when a credential is issued/created.
    event CredentialIssued(
        bytes32 indexed digest,
        address indexed subject,
        address indexed issuer,
        uint256 insertedBlock
    );

    // Logged when a credential is revoked by some owner.
    event CredentialRevoked(
        bytes32 indexed digest,
        address indexed subject,
        address indexed revoker,
        uint256 revokedBlock,
        bytes32 reason
    );

    // Logged when a credential is signed.
    event CredentialSigned(
        address indexed signer,
        bytes32 indexed digest,
        uint256 signedBlock
    );

    // Logged when a credential is aggregated.
    event VerifiedCredential(
        address indexed aggregator,
        address indexed subject,
        bytes32 indexed digestSum
    );

    // function isValid(bytes32 digest, uint256 period) external view returns (bool);

    /**
     * @dev isRevoked checks if the credential was revoked based on it's digest.
     */
    function isRevoked(bytes32 digest) external view returns (bool); //

    /**
     * @dev certified checks if a credential was signed by all parties.
     */
    function certified(bytes32 digest) external view returns (bool);

    /**
     * @dev certified registers the creation of a credential for
     * a particular subject
     */
    function registerCredential(address subject, bytes32 digest) external;

    /**
     * @dev confirmCredential confirms the agreement about the 
     * credential between the subject and the issuer.
     */
    function confirmCredential(bytes32 digest) external;

    /**
     * @dev revokeCredential revokes a credential for a given reason 
     * based on it's digest.
     */
    function revokeCredential(bytes32 digest, bytes32 reason) external;

    /**
     * @dev verifyCredential verifies if a given credential 
     * (i.e. represented by it's digest) corresponds to the aggregation 
     * of all stored credentials of a particular subject.
     */
    // TODO: add digest parameter
    function verifyCredential(address subject) external returns (bytes32);

    /**
     * @dev verifyCredential iteractivally verifies if a given credential
     * (i.e. represented by it's digest) corresponds to the aggregation 
     * of all stored credentials of a particular subject in all given contracts.
     */
    function verifyCredential(
        address subject,
        bytes32 digest,
        address[] calldata contracts
    ) external returns (bytes32);
}
