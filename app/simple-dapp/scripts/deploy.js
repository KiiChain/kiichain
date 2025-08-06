const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners(); // akun pemilik token

  const MyToken = await hre.ethers.getContractFactory("MyToken");
  const token = await MyToken.deploy(deployer.address); // owner sebagai constructor param
  await token.waitForDeployment(); // ganti dari .deployed() ke .waitForDeployment()

  console.log("✅ Token deployed to:", await token.getAddress());
  console.log("👑 Owner address:", deployer.address);
}

main().catch((error) => {
  console.error("❌ Deployment failed:", error);
  process.exitCode = 1;
});
