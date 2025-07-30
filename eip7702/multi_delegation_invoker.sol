// SPDX-License-Identifier: UNSPECIFIED
pragma solidity ^0.8.24;

contract MultiDelegationInvoker {
    event PingSuccess(address from);
    event PingStart(address from);

    function triggerPings(address[] calldata froms) external {
        emit PingStart(msg.sender);
        
        for (uint i = 0; i < froms.length; i++) {
            address from = froms[i];
            bytes memory code = new bytes(23);

            // Load the first 23 bytes of the code at `from`
            assembly {
                extcodecopy(from, add(code, 0x20), 0, 23)
            }

            // Check if it starts with 0xef0100
            if (
                code.length == 23 &&
                uint8(code[0]) == 0xef &&
                uint8(code[1]) == 0x01 &&
                uint8(code[2]) == 0x00
            ) {
                // After verifying code starts with 0xef0100
                address someModule;
                assembly {
                    someModule := shr(96, mload(add(code, 0x23)))
                }

                (bool ok, ) = someModule.call(abi.encodeWithSignature("ping()"));
                require(ok, "Ping failed");

                emit PingSuccess(from);
            } else {
                revert("Not delegated or invalid delegation format");
            }
        }
    }
}