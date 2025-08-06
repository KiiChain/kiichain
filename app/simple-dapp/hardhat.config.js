require("@nomicfoundation/hardhat-toolbox");
require("dotenv").config(); // agar bisa ambil dari file .env

module.exports = {
  solidity: "0.8.20",
  networks: {
    kiichain: {
      url: "https://json-rpc.uno.sentry.testnet.v3.kiivalidator.com/",
      accounts: [process.env.PRIVATE_KEY]
    }
  }
};
