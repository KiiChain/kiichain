import React, { useState, useEffect } from "react";
import { ethers } from "ethers";

// Ganti dengan alamat kontrak token yang sudah kamu deploy
const tokenAddress = "0x72b0CbA97F4d3d3E811DA303B7873616c49f039b";

// ABI minimal ERC20 (transfer, balanceOf, mint jika ada)
const tokenABI = [
  "function balanceOf(address) view returns (uint256)",
  "function mint(address to, uint256 amount)",
  "function transfer(address to, uint256 amount) returns (bool)",
];

function App() {
  const [account, setAccount] = useState(null);
  const [balance, setBalance] = useState("0");
  const [provider, setProvider] = useState(null);
  const [signer, setSigner] = useState(null);

  // State untuk mint token
  const [mintAmount, setMintAmount] = useState("");

  // State untuk send token
  const [sendAddress, setSendAddress] = useState("");
  const [sendAmount, setSendAmount] = useState("");

  // Connect wallet function
  async function connectWallet() {
    if (window.ethereum) {
      try {
        const accounts = await window.ethereum.request({
          method: "eth_requestAccounts",
        });
        setAccount(accounts[0]);
        const prov = new ethers.BrowserProvider(window.ethereum);
        setProvider(prov);
        const sign = await prov.getSigner();
        setSigner(sign);
        updateBalance(accounts[0], prov);
      } catch (err) {
        alert("Connection failed: " + err.message);
      }
    } else {
      alert("Please install MetaMask!");
    }
  }

  // Update balance function
  async function updateBalance(accountAddr, prov = provider) {
    if (!prov || !accountAddr) return;
    try {
      const contract = new ethers.Contract(tokenAddress, tokenABI, prov);
      const bal = await contract.balanceOf(accountAddr);
      setBalance(ethers.formatUnits(bal, 18));
    } catch (err) {
      alert("Failed to get balance: " + err.message);
    }
  }

  // Mint token function
  async function mintToken() {
    if (!signer) {
      alert("Connect wallet first!");
      return;
    }
    if (!mintAmount || isNaN(mintAmount) || mintAmount <= 0) {
      alert("Enter a valid mint amount");
      return;
    }
    try {
      const contract = new ethers.Contract(tokenAddress, tokenABI, signer);
      const amount = ethers.parseUnits(mintAmount, 18);
      const tx = await contract.mint(account, amount);
      await tx.wait();
      alert(`Minted ${mintAmount} MTK to your address`);
      updateBalance(account);
      setMintAmount("");
    } catch (err) {
      alert("Mint failed: " + err.message);
    }
  }

  // Send token function
  async function sendToken() {
    if (!signer) {
      alert("Connect wallet first!");
      return;
    }
    if (!ethers.isAddress(sendAddress)) {
      alert("Invalid recipient address");
      return;
    }
    if (!sendAmount || isNaN(sendAmount) || sendAmount <= 0) {
      alert("Enter a valid amount to send");
      return;
    }
    try {
      const contract = new ethers.Contract(tokenAddress, tokenABI, signer);
      const amount = ethers.parseUnits(sendAmount, 18);
      const tx = await contract.transfer(sendAddress, amount);
      await tx.wait();
      alert(`Sent ${sendAmount} MTK to ${sendAddress}`);
      updateBalance(account);
      setSendAddress("");
      setSendAmount("");
    } catch (err) {
      alert("Send failed: " + err.message);
    }
  }

  // Reload balance if account or provider changes
  useEffect(() => {
    if (account && provider) {
      updateBalance(account, provider);
    }
  }, [account, provider]);

  return (
    <div style={{ padding: 20, fontFamily: "Arial, sans-serif" }}>
      <h1>MyToken DApp</h1>
      {account ? (
        <>
          <p>
            <strong>Connected:</strong> {account}
          </p>
          <p>
            <strong>Balance:</strong> {balance} MTK
          </p>

          {/* Mint Token */}
          <div style={{ marginBottom: 30 }}>
            <h3>Mint Token</h3>
            <input
              type="number"
              placeholder="Amount to mint"
              value={mintAmount}
              onChange={(e) => setMintAmount(e.target.value)}
              min="0"
              style={{ marginRight: 10, width: "150px" }}
            />
            <button onClick={mintToken}>Mint Token</button>
          </div>

          {/* Send Token */}
          <div>
            <h3>Send Token</h3>
            <input
              type="text"
              placeholder="Recipient address"
              value={sendAddress}
              onChange={(e) => setSendAddress(e.target.value)}
              style={{ width: "300px", marginRight: 10 }}
            />
            <input
              type="number"
              placeholder="Amount to send"
              value={sendAmount}
              onChange={(e) => setSendAmount(e.target.value)}
              min="0"
              style={{ width: "120px", marginRight: 10 }}
            />
            <button onClick={sendToken}>Send Token</button>
          </div>
        </>
      ) : (
        <button onClick={connectWallet}>Connect MetaMask Wallet</button>
      )}
    </div>
  );
}

export default App;
