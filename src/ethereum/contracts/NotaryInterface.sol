pragma solidity >=0.5.13;

// TODO update the interface fields
interface NotaryInterface {
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
    event AggregatedCredential(
        address indexed aggregator,
        address indexed subject,
        bytes32 indexed digestSum,
        uint256 firstBlock,
        uint256 lastBlock
    );

    function wasRevoked(bytes32 digest) external view returns (bool);
    function issue(address subject, bytes32 digest) external;
    function certified(bytes32 digest) external view returns (bool);
    function confirmProof(bytes32 digest) external;
    function revoke(bytes32 digest, bytes32 reason) external;
    function aggregate(address subject)
        external
        returns (bytes32, uint256, uint256);
}
