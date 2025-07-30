// SPDX-License-Identifier: UNSPECIFIED
pragma solidity ^0.8.24;

contract Invoked {
    event Pinged(address sender);

    function ping() external {
        emit Pinged(msg.sender);
    }
}