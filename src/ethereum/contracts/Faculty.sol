// SPDX-License-Identifier: MIT
pragma solidity >=0.6.0 <0.8.0;
pragma experimental ABIEncoderV2;

import "ct-eth/contracts/node/Node.sol";

/**
 * @title Academic Faculty
 */
contract Faculty is Node {
    constructor(address[] memory owners, uint8 quorum)
        Node(Role.Inner, owners, quorum) {
    }
}
