pragma solidity ^0.8.0;

contract Counter {
    int256 private counter;

    constructor() {
        counter = 0;
    }

    function increment() public {
        counter += 1;
    }

    function getCounter() public view returns (int256) {
        return counter;
    }
}
