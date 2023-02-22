// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

// Basic is a contract that implements a basic key/value store.
contract Basic {
    // Version is the current version of Store.
    string public Version;

    // Items is a mapping which will store key/value data.
    // We can only store/retrieve from the map, not iterate.
    mapping (string => uint256) public Items;

    // ItemSet is an even that outputs any updates to
    // the key/value data to the transaction receipt's logs
    event ItemSet(string key, uint256 value);

    // The constructor is automatically executed when the contract is deployed.
    constructor() {
        Version = "1.1";
    }


    // SetItem is an external-only function that accepts a key/value
    // pair and updates the contract's internal storage accordingly.
    function SetItems(string memory key, uint256 value) external {
        Items[key] = value;
        emit ItemSet(key, value);
    }
}