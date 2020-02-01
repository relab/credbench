pragma solidity >=0.5.13;

import "./Notary.sol";
// import "@openzeppelin/contracts/math/SafeMath.sol";

/**
 * @title TimedNotary
 * @dev Notary that emits certificates only within a time interval.
 * Based on Openzeppelin TimedCrowdsale
 */
abstract contract TimedNotary is Notary {
    // using SafeMath for uint256;

    uint256 private _startingTime;
    uint256 private _endingTime;

    /**
     * @param newEndingTime new ending time
     * @param prevEndingTime old ending time
     */
    event NotaryPeriodExtended(uint256 prevEndingTime, uint256 newEndingTime);

    /**
     * @dev Reverts if not in notary time range.
     */
    modifier onlyAfterStart {
        require(
            isStarted(),
            "TimedNotary: the notarization period didn't start yet"
        );
        _;
    }

    modifier whileNotEnded {
        require(
            stillRunning(),
            "TimedNotary: the notarization period has already ended"
        );
        _;
    }

    /**
     * @dev Constructor, takes notary starting and ending times.
     * @param startingTime notary starting time
     * @param endingTime notary ending time
     */
    constructor(uint256 startingTime, uint256 endingTime) public {
        // solhint-disable-next-line not-rely-on-time
        require(
            startingTime >= block.timestamp,
            "TimedNotary: starting time cannot be in the past"
        );
        // solhint-disable-next-line max-line-length
        require(
            endingTime > startingTime,
            "TimedNotary: ending time cannot be smaller than starting time"
        );

        _startingTime = startingTime;
        _endingTime = endingTime;
    }

    /**
     * @return the notary starting time.
     */
    function startingTime() public view returns (uint256) {
        return _startingTime;
    }

    /**
     * @return the notary ending time.
     */
    function endingTime() public view returns (uint256) {
        return _endingTime;
    }

    /**
     * @return true if the notary is started, false otherwise.
     */
    function isStarted() public view returns (bool) {
        // solhint-disable-next-line not-rely-on-time
        return block.timestamp >= _startingTime;
    }

    /**
     * @dev Checks whether the notarization period has already elapsed.
     * @return Whether notary period has elapsed
     */
    function hasEnded() public view returns (bool) {
        // solhint-disable-next-line not-rely-on-time
        return block.timestamp >= _endingTime;
    }

    /**
     * @return true if the notarization period still running.
     */
    function stillRunning() public view returns (bool) {
        // solhint-disable-next-line not-rely-on-time
        return isStarted() && !hasEnded();
    }

    /**
     * @dev Extend the notarization time.
     * @param newEndingTime the new notary ending time
     */
    function _extendTime(uint256 newEndingTime) internal whileNotEnded {
        // solhint-disable-next-line max-line-length
        require(
            newEndingTime > _endingTime,
            "TimedNotary: new ending time is before current ending time"
        );

        emit NotaryPeriodExtended(_endingTime, newEndingTime);
        _endingTime = newEndingTime;
    }

    function aggregate(address subject)
        public
        view
        override
        onlyAfterStart
        returns (bytes32)
    {
        //timed aggregation should only aggregate certificates within the    valid period
        return super.aggregate(subject);
    }
}
