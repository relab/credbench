pragma solidity >=0.5.13;

import "./Issuer.sol";
// import "@openzeppelin/contracts/math/SafeMath.sol";
// import "@openzeppelin/contracts/cryptography/ECDSA.sol";

/**
 * @title AccountableIssuer's contract
 * This contract consider on-chain signatures verification.
 * TODO: Implement using EIP712:
 * https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md
 */
abstract contract AccountableIssuer is Issuer {
    address[] public issuers;

    // Map of all issuers sub-contracts
    mapping(address => bool) public isIssuer;

    // Logged when a credential is issued/created.
    event IssuerAdded(
        address indexed issuerAddress,
        address indexed addedBy
    );

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
            require(isIssuer[issuerAddress], "AccountableIssuer: issuer's address doesn't found"); //FIXME: Use assert
            Issuer issuer = Issuer(issuerAddress);
            bytes32 proof = issuer.getProof(subject);
            require(proof != bytes32(0), "AccountableIssuer: aggregation on sub-contract not found"); //TODO: use assert
            digests[i] = proof;
        }
        return digests;
    }

    // Register a credential with on-chain verification
    function registerCredential(
        address subject,
        bytes32 digest,
        bytes32 digestRoot,
        address[] memory issuersAddresses
    ) public onlyOwner {
        bytes32[] memory digests = collectCredentials(subject, issuersAddresses);
        // TODO: call aggregation library (implementing sum/tree methods)
        bytes32 aggregatedDigest = keccak256(abi.encode(digests, digest));
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
            require(isIssuer[issuerAddress], "AccountableIssuer: address doesn't registered"); //FIXME: Use assert
            Issuer issuer = Issuer(issuerAddress);
            issuer.verifyCredential(subject, digest[i]);
            //TODO: CATCH if assert fail
        }
    }
}
