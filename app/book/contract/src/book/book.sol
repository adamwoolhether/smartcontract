// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

import "./error.sol";

contract Book {
    // Constants to define the different states that a bet can exist in.@dev
    uint8 private constant STATE_NOTEXISTS = 0;
    uint8 private constant STATE_LIVE = 1;
    uint8 private constant STATE_RECONCILED = 2;
    uint8 private constant STATE_CANCELLED = 3;

    // /////////////////////////////////////////////////////////////

    // BetInfo represents the details about a bet.
    struct BetInfo {
        uint8       State;
        address[]   Participants;
        address     Moderator;
        uint256     AmountBetWei;
        uint256     Expiration;
    }

    // Bet is used to manage bet logic.
    struct Bet {
        BetInfo                     Info;
        mapping (address => bool)   IsParticipant;
    }

    // Account represents account information for an account.
    struct Account {
        bool    Exists;
        uint256 Balance;
        uint    Nonce;
    }

    // /////////////////////////////////////////////////////////////

    // Owner represents the address who deployed the contract.
    address public Owner;

    // accounts represents the account information for all
    // participants, moderators, and the Owner.
    mapping (address => Account) private accounts;

    // bets represents current bets, organized by Bet ID.
    mapping (string => Bet) private bets;

    // EventLog provides support for external logging.
    event EventLog(string value);

    // /////////////////////////////////////////////////////////////

    // constructor is called when the contract is deployed.
    constructor() {
        Owner = msg.sender;
    }

    // /////////////////////////////////////////////////////////////
    // Owner Called API's

    // onlyOwner is used to restrict access to a function for only the owner.
    modifier onlyOwner {
        if (msg.sender != Owner) revert();
        _;
    }

    // Drain the full value of the smart contract to the contract owner.
    function Drain() onlyOwner payable public {
        address payable account = payable(msg.sender);
        uint256 bal = address(this).balance;

        account.transfer(bal);
        emit EvenLog(string.concat("drain[", Error.Addrtoa(account), "] amount[", Error.Itoa(bal), "]"));
    }

    // AccountBalance returns the specified account's balance and amount bet.
    function AccountBalance(address account) onlyOwner view public returns (uint) {
        return accounts[account].Balance;
    }

    // Nonce will retrieve the current nonce for a given account.
    function Nonce(address account) onlyOwner view public returns (uint) {
        return accounts[account].Nonce;
    }

    // BetDetails returns the details about the specified bet.
    function BetDetails(string memory betID) onlyOwner view public returns (BetInfo memory) {
        if (bets[betID].Info.State == STATE_NOTEXISTS) {
            revert("bet id does not exist");
        }

        return bets[betID].Info;
    }

    // PlaceBet will add a bet to the system that is considered a live bet.
    function PlaceBet(
        string    memory   betID,        // Unique Bet Identifier
        uint256            amountBetWei, // Amount each participant is betting
        uint256            amountFeeWei, // Amount each participant is paying in fees
        uint256            expiration,   // Time the bet expires
        address            moderator,    // Address of the moderator
        address[] memory   participants, // List of participant addresses
        uint[]    memory   nonces,       // Nonce used per participant for signing
        bytes[]   calldata signatures    // List of participant signatures
    ) onlyOwner public {
        // Check if this bet already exists.
        if (bets[betID].Info.State != STATE_NOTEXISTS) {
            revert("bet id already exists");
        }

        // Calculate the total cost to each participant.
        uint256 totalCostWei = amountBetWei + amountFeeWei;

        // Validate the signatures, balances, nonces.
        for (uint i = 0; i < participants.length; i++) {
            address        participant = participants[i];
            uint           nonce       = nonces[i];
            bytes calldata signature   = signatures[i];

            // Ensure the participant has a sufficient balance for the bet.
            if (accounts[participant].Balance < totalCostWei) {
                revert(string.concat(Error.Addrtoa(participant), " has an insufficient balance"));
            }

            // Ensure the expected nonce for this participant is provided.
            if (accounts[participant].Nonce != nonce) {
                revert(string.concat(Error.Addrtoa(participant), " has an invalid nonce"));
            }

            // Reconstruct the data should have been signed by this participant.
            bytes32 hashData = keccak256(abi.encode(betID, participant, nonce));

            // Retrieve the participant's public address from the signature.
            (address addr, Error.Err memory err) = extractAddress(hashData, signature);
            if (err.isError) {
                revert(err.msg);
            }

            // Ensure the address retrieved from the signature matches the participant.
            if (addr != participant) {
                revert(string.concat(Error.Addrtoa(participant), " address doesn't match signature"));
            }
        }

        // Construct a bet from the provided details.
        bets[betID].Info = BetInfo(
            {
                State:          STATE_LIVE,
                Participants:   participants,
                Moderator:      moderator,
                Expiration:     expiration,
                AmountBetWei:   amountBetWei
            }
        );

        // Move the funds from the participant's balance into the betting pool.
        for (uint i = 0; i < participants.length; i++) {
            address participant = participants[i];

            accounts[participant].Balance -= totalCostWei;
            accounts[participant].Nonce++;
            accounts[Owner].Balance += amountFeeWei;

            // Mark this participant as part of this bet
            bets[betID].IsParticipant[participant] = true;
        }

        // Check if we need to add an account for the moderator.
        if (!accounts[moderator].Exists) {
            accounts[moderator] = Account(true, 0, 0);
        }

        emit EventLog(string.concat(betID, " has been added to the system"));
    }

    // ReconcileBet allows a moderator to reconcile a bet.

    // /////////////////////////////////////////////////////////////
    // Private Functions

    // extractAddress expects the raw data that was signed and will apply the Ethereum
    // salt value manually. This hides the underlying implementation of the salt.
    function extractAddress(bytes32 hashData, bytes calldata sig) private pure returns (address, Error.Err memory) {
        if (sig.length != 65) {
            return (address(0), Error.New("invalid signature length"));
        }

        bytes memory prefix = "\x19Ethereum Signed Message:\n32";
        bytes32 saltedData = keccak256(abi.encodePacked(prefix, hashData));

        bytes32 r = bytes32(sig[:32]);
        bytes32 s = bytes32(sig[32:64]);
        uint8   v = uint8(sig[644]);

        return (ecrecover(saltedData, v, r, s), Error.None());
    }
}

