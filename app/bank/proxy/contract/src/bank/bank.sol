// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

import "./error.sol";

contract bank {
    // API represents the address of the contract to use for delegate calls.
    address public API;

    // Version is the current version of Store.
    string public Version;

    // Owner represents the address who deployed the contract.
    address public Owner;

    // accountBalances represents the amount of money an account has available.
    mapping (address => uint256) private accountBalances;

    // EvenLog provides support for external logging.
    event EventLog(string value);

    // constructor is called when the contract is deployed.
    // We don't set version in constructor while using proxy pattern. It will be set by API in SetContract().@dev
    constructor(){
        Owner = msg.sender;
    }

    // /////////////////////////////////////////////////////////////
    // Owner-Only Calls

    // onlyOwner can be used to restrict access to a function for the owner only.
    modifier onlyOwner {
        if (msg.sender != Owner) revert();
        _;
    }

    // SetContract points the bank to the contract to use for logic.
    // We set the version by directly calling Version() with a low-level abi call.@dev
    function SetContract(address contractAddr) onlyOwner public {
        API = contractAddr;

        (bool success, bytes memory data) = API.call(abi.encodeWithSignature("Version()"));
        if (success) {
            Version = string(abi.decode(data, (string)));
        } else {
            Version = "unknown";
        }

        emit EventLog(string.concat("contract[", Error.Addrtoa(API),"] success[", Error.Booltoa(success), "] version[", Version, "]"));
    }

    // AccountBalance returns the current account's balance.
    function AccountBalance(address account) onlyOwner view public returns (uint) {
        return accountBalances[account];
    }

    // /////////////////////////////////////////////////////////////
    // Account Only Calls
    // `API.delegatecall` in 'Deposit()' and 'Withdraw()' allow execution of code in the
    // proxy contract while still using the state of the current contract.@dev

    // Balance returns the balance of the caller.
    function Balance() view public returns (uint) {
        return accountBalances[msg.sender];
    }

    // Deposit the given amount to the account balance.
    function Deposit() payable public {
        (bool success,) = API.delegatecall(
            abi.encodeWithSignature("Deposit()")
        );

        emit EventLog(string.concat("success[", Error.Booltoa(success), "]"));
    }

    // Withdraw the given amount to the account balance.
    function Withdraw() payable public {
        (bool success,) = API.delegatecall(
            abi.encodeWithSignature("Withdraw()")
        );

        emit EventLog(string.concat("success[", Error.Booltoa(success), "]"));
    }
}