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
 - Make the Notary contract as a Library
 - Allow multiple subjects
 - Add expiration date information?
 - Is RegistryEntry/Claim/VerifiableClaim a more appropriate name?
   A set of claims about the same subject issued by the same issuer would constitute a subject's credential
 - Is better to use a multi-signature scheme to aggregate signatures?
 - Define a common interface like the token interface for certificates?

### course
 - Is necessary to have a list of enrolled students? store a array of student addresses in the contract solves the problem but has an inefficient removal operation, other option is let to the client to keep track of the enrolled students and just use the chain to check.
 - Bind a contract instance instead of inheritance in course, this will facilitate the decouple of the client implementation and allow multiple courses to share the same notary registry contract

### faculty
 - Rename to Certificate Authority
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