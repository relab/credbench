pragma solidity >=0.5.13;

struct Proof { mapping(address => bytes32) _proofs; }

library CredentialSum {

    // Logged when a credential is aggregated.
    event AggregatedProof(
        address indexed aggregator,
        address indexed subject,
        bytes32 indexed proof,
        uint256 aggregatedBlock
    );

    // Aggregate credentials and produce a proof of it
    function generateProof(Proof storage self, address subject, bytes32[] memory certificates)
        public
        returns (bytes32)
    {
        require(self._proofs[subject] == bytes32(0), "CredentialSum: proof already generated");
        require(
            certificates.length > 0,
            "CredentialSum: there is no certificates"
        );
        bytes32 proof = keccak256(abi.encode(certificates));
        self._proofs[subject] = proof;
        // TODO: sender should be issuer not contract
        emit AggregatedProof(msg.sender, subject, proof, block.number);
        return proof;
    }

    function verifyProof(Proof storage self, address subject, bytes32[] memory certificates)
        public
        view
        returns (bool)
    {
        require(self._proofs[subject] != bytes32(0), "CredentialSum: proof not exists");
        bytes32 proof = keccak256(abi.encode(certificates));
        return (self._proofs[subject] == proof);
    }

    function proofs(Proof storage self, address subject)
        public
        view
        returns (bytes32)
    {
        return self._proofs[subject];
    }
}