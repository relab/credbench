pragma solidity >=0.5.13;

// TODO update the interface fields
interface NotaryInterface {
    // Logged when a credential is issued and signed by all parties (owners + subject).
    event CredentialIssued(bytes32 indexed digest, address indexed subject, address indexed issuer, bytes32 previousDigest, uint insertedBlock);

    // Logged when a credential is revoked by some owner.
    event CredentialRevoked(bytes32 indexed digest, address indexed subject, address indexed issuer, uint revokedBlock);

    function wasRevoked(bytes32 digest) external view returns (bool);
    function issue(address subject, bytes32 digest) external;
    function certified(bytes32 digest) external view returns (bool);
    function requestProof(bytes32 digest) external;
    function revoke(bytes32 digest) external;
    function aggregate(address subject) external view returns (bytes32);
}