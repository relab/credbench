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

### Issuer
 - Make the Issuer contract as a Library
 - Allow multiple subjects
 - Add expiration date information?
 - Is RegistryEntry/Claim/VerifiableClaim a more appropriate name?
   A set of claims about the same subject issued by the same issuer would constitute a subject's credential
 - Is better to use a multi-signature scheme to aggregate signatures?
 - Define a common interface like the token interface for certificates?

### course
 - Rename to leaf Issuer / leaf authority / subordinate CA or registration authority (like in PKI)
 - Is necessary to have a list of enrolled students? store a array of student addresses in the contract solves the problem but has an inefficient removal operation, other option is let to the client to keep track of the enrolled students and just use the chain to check.
 - Bind a contract instance instead of inheritance in course, this will facilitate the decouple of the client implementation and allow multiple courses to share the same Issuer registry contract

### faculty
 - Rename to intermediary Issuer / certificate authority
 - Deploy Course contracts
 - Implement admission process
 - Implement enrollment process
 - Implement issue and revocation process
 - Implement appeals process

## Client (a.k.a [user agent](https://www.w3.org/TR/verifiable-claims-data-model/#terminology))
 - SWARM interface
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


## Prototype

1. (ON) Deploy university contract and assign employees
2. (ON) University contract creates faculty contract and assign members
3. (ON) Faculty contract creates courses contracts for the semester and assign teachers/evaluators (optionally add contract address for conflict resolution)
4. For each course in the semester do:
   4.1. (OFF) Teacher create a verifiable credential as specified in the VC standard of the student assignment (optionally create only one for the whole course aggregating the assignments scores - save money/space but reduce accountability)
   4.2. (OFF) Teacher saves the credential and the necessary proofs on the storage (swarm)
   4.3. (OFF) Student get access to his credential on the storage
   4.4. (ON) Teacher/Evaluator register certificate for student on course (add student on course)
   4.5. (ON) Teacher/Evaluator sign certificate receipt
   4.6. (ON) Student accept or reject the certificate
     - If rejected, escalate for faculty
     - Otherwise publish it
   4.7. Repeat 4.1 to 4.6 until end of the course
   4.8. (ON) When the course ends, teacher/evaluator/faculty request course certificate by aggregating all students' assignments certificates in a tree. Leaves are the assignments cert. and root is the course certificate. Root is stored in the course contract. Delete the assignments after aggregation and publish course certificates per student
5. (ON) On the end of the semester, faculty member build semester certificate by requesting faculty contract to aggregate courses certificates for all courses (merkle tree of the semester, leaves are the courses certificate + course address hash)
6. Repeat 3 to 5 until the end of the bachelor's to generate the diploma by aggregating the semester certificates.

Meeting 05/12/2019
- Build the base line first (consider unlimited resources)
- Don't care about external courses for now, but think about it. How incorporate external certificates, e.g. student make a course in one institution and move to other place and continue his studies.
- Consider two hashes one for validation and another for the verifiable credential metadata

## TODO code

1. Finish the timed Issuer contract
2. Append the new certificate hash with the previous, creating a chain of certificates
3. Create a verifiable credential as a Issuer certificate
4. Verify the validity and authenticity of the certificates (implement tests)
5. Add commands in the client to perform the notarization process example
6. Allow contracts to verify on-chain and off-chain signatures (keep different versions for cost analyses):
on chain sign and verification, off chain sign and on chain verification, off chain sign and verification.
The current implementation made on chain sign and off chain verification
(https://github.com/OpenZeppelin/openzeppelin-contracts/tree/master/contracts/cryptography)
7. Estimate the cost of deployment of the contracts

## TODO tests

1. Setup the cluster
2. Setup private ethereum/swarm nodes
3. Deploy contracts
4. Create scripts to generate load
5. Collect metrics (cost, number of txs, tx/s) 

## TODO paper

1. Update the information about the diploma construction using the append-only approach
2. Collect information about the current costs of issuing diplomas
3. Send an email to the admission department asking about the current costs and statistics


HARD DEADLINE: 12 May 2020

When aggregate the previous can delete the information from the contract state