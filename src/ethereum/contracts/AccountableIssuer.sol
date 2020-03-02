pragma solidity >=0.5.13 <0.7.0;

import "./Issuer.sol";
// import "@openzeppelin/contracts/math/SafeMath.sol";
// import "@openzeppelin/contracts/cryptography/ECDSA.sol";

/**
 * @title AccountableIssuer's contract
 * This contract consider implicit signatures verification.
 * TODO Implement using EIP712:
 * https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md
 */
abstract contract AccountableIssuer is Issuer {
    address[] public issuers;

    // Map of all issuers sub-contracts
    mapping(address => bool) public isIssuer;

    // Logged when an issuer added.
    event IssuerAdded(
        address indexed issuerAddress,
        address indexed addedBy
    );

    //TODO: blacklist issuers?

    constructor(address[] memory owners, uint256 quorum)
        public
        Issuer(owners, quorum)
    {
        // solhint-disable-previous-line no-empty-blocks
    }

    function addIssuer(address issuer) public {
        require(!isIssuer[issuer], "AccountableIssuer: issuer already added");
        isIssuer[issuer] =  true;
        issuers.push(issuer);
        emit IssuerAdded(issuer, msg.sender);
    }

    /**
     * @dev collectCredentials collect all the aggregated digests of 
     * a given subject in all sub-contracts levels.
     */
    function collectCredentials(address subject, address[] memory issuersAddresses)
        public
        view
        onlyOwner
        returns (bytes32[] memory)
    {
        require(issuersAddresses.length > 0, "AccountableIssuer: require at least one issuer");
        bytes32[] memory digests = new bytes32[](issuersAddresses.length);
        for (uint256 i = 0; i < issuersAddresses.length; i++) {
            address issuerAddress = address(issuersAddresses[i]);
            require(isIssuer[issuerAddress], "AccountableIssuer: issuer's address doesn't found");
            Issuer issuer = Issuer(issuerAddress);
            bytes32 proof = issuer.getProof(subject);
            require(proof != bytes32(0), "AccountableIssuer: aggregation on sub-contract not found");
            digests[i] = proof;
        }
        return digests;
    }

    function registerCredential(
        address subject,
        bytes32 digest,
        bytes32 digestRoot,
        address[] memory issuersAddresses
    ) public onlyOwner {
        require(aggregatedProofs.proofs(subject) == bytes32(0), "Issuer: credentials already aggregated, not possible to issue new credentials");
        bytes32[] memory d = collectCredentials(subject, issuersAddresses);
        bytes32[] memory digests = new bytes32[](d.length + 1);
        uint256 i = 0;
        for (; i < d.length; i++) {
            digests[i] = d[i];
        }
        // Add current credential
        digests[i] = digest;
        bytes32 aggregatedDigest =  aggregatedProofs.generateProof(subject, digests);
        require(aggregatedDigest == digestRoot, "AccountableIssuer: root is not equal");
        _issue(subject, digest);
        emit CredentialSigned(msg.sender, digest, block.number);
    }

    /**
     * @dev verifyCredential iteractivally verifies if a given credential
     * (i.e. represented by it's digest) corresponds to the aggregation 
     * of all stored credentials of a particular subject in all given contracts.
     */
    function verifyCredential(address subject, bytes32[] memory digest, address[] memory issuersAddresses) public view {
        require(issuersAddresses.length > 0, "Issuer: require at least one issuer");
        for (uint256 i = 0; i < issuersAddresses.length; i++) {
            address issuerAddress = address(issuersAddresses[i]); 
            require(isIssuer[issuerAddress], "AccountableIssuer: address not registered");
            Issuer issuer = Issuer(issuerAddress);
            issuer.verifyCredential(subject, digest[i]);
            //TODO: CATCH if assert fail
        }
    }
}
