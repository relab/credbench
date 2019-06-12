# TODO

A list of things to do, or maybe not...

## General
 - Choose a better repository name
 - Move contracts to another repository
 - Move uis-dapp to another repository
 - Add documentation (running, compiling, testing)
 - Add the flow charts
 - Study anonymization approaches that allow smart contracts (https://www.aztecprotocol.com/, https://zokrates.github.io/)
  
## Contracts (Verifiable credential registry system)

### owners 
 - Perform ownership operations
 - transferOwnership

### notary
 - Allow multiple subjects
 - Add expiration date information?
 - Is RegistryEntry/Claim/VerifiableClaim a more appropriate name?
   A set of claims about the same subject issued by the same issuer would constitute a subject's credential
 - Is better to use a multi-signature scheme to aggregate signatures?

### course
 - Is necessary to have a list of enrolled students? store a array of student addresses in the contract solves the problem but has an inefficient removal operation, other option is let to the client to keep track of the enrolled students and just use the chain to check.
 - Bind a contract instance instead of inheritance in course, this will facilitate the decouple of the client implementation and allow multiple courses to share the same notary registry contract

### faculty
 - Not implemented
 - Implement admission process
 - Implement enrollment process
 - Implement issue and revocation process
 - Implement appeals process

## Client (a.k.a [user agent](https://www.w3.org/TR/verifiable-claims-data-model/#terminology))
 - IPFS interface
 - Produce JSON credential schemas
 - P2P handshake between agents (biometric based)
 - DID compatibility
 - Manage the keys/addresses properly and remove hardcoded

### Makefile
 - Generate clean code (without dependencies) using --exc parameter in abigen
 - Add support to others OS (osx, windows)
 
 # Verifiable credential references

 - https://www.w3.org/TR/verifiable-claims-data-model/
 - https://www.w3.org/TR/verifiable-claims-use-cases/
 - https://w3c.github.io/vc-imp-guide/