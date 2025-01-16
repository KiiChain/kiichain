const { fundAddress, fundKiiAddress, getKiiBalance, associateKey, importKey, waitForReceipt, bankSend, evmSend, getNativeAccount} = require("./lib");
const { expect } = require("chai");

describe("Associate Balances", function () {

    const keys = {
        "test1": {
            kiiAddress: 'kii14jekmh7yruasqx4k372mrktsd7hwz454snw0us',
            evmAddress: '0x90684e7F229f2d8E2336661f79caB693E4228Ff7'
        },
        "test2": {
            kiiAddress: 'kii14jekmh7yruasqx4k372mrktsd7hwz454snw0us',
            evmAddress: '0x28b2B0621f76A2D08A9e04acb7F445E61ba5b7E7'
        },
        "test3": {
            kiiAddress: 'kii14jekmh7yruasqx4k372mrktsd7hwz454snw0us',
            evmAddress: '0xCb2FB25A6a34Ca874171Ac0406d05A49BC45a1cF',
            castAddress: 'kii14jekmh7yruasqx4k372mrktsd7hwz454snw0us',
        }
    }

    const addresses = {
        kiiAddress: 'kii14jekmh7yruasqx4k372mrktsd7hwz454snw0us',
        evmAddress: '0x90684e7F229f2d8E2336661f79caB693E4228Ff7'
    }

    function truncate(num, byThisManyDecimals) {
        return parseFloat(`${num}`.slice(0, 12))
    }

    async function verifyAssociation(kiiAddr, evmAddr, associateFunc) {
        const beforeKii = BigInt(await getKiiBalance(kiiAddr))
        const beforeEvm = await ethers.provider.getBalance(evmAddr)
        const gas = await associateFunc(kiiAddr)
        const afterKii = BigInt(await getKiiBalance(kiiAddr))
        const afterEvm = await ethers.provider.getBalance(evmAddr)

        console.log(`KII Balance (before): ${beforeKii}`)
        console.log(`EVM Balance (before): ${beforeEvm}`)
        console.log(`KII Balance (after): ${afterKii}`)
        console.log(`EVM Balance (after): ${afterEvm}`)

        const multiplier = BigInt(1000000000000)
        expect(afterEvm).to.equal((beforeKii * multiplier) + beforeEvm - (gas * multiplier))
        expect(afterKii).to.equal(truncate(beforeKii - gas))
    }

    before(async function(){
        await importKey("test1", "../contracts/test/test1.key")
        await importKey("test2", "../contracts/test/test2.key")
        await importKey("test3", "../contracts/test/test3.key")
    })

    it("should associate with kii transaction", async function(){
        const addr = keys.test1
        await fundKiiAddress(addr.kiiAddress, "10000000000")
        await fundAddress(addr.evmAddress, "200");

        await verifyAssociation(addr.kiiAddress, addr.evmAddress, async function(){
            await bankSend(addr.kiiAddress, "test1")
            return BigInt(20000)
        })
    });

    it("should associate with evm transaction", async function(){
        const addr = keys.test2
        await fundKiiAddress(addr.kiiAddress, "10000000000")
        await fundAddress(addr.evmAddress, "200");

        await verifyAssociation(addr.kiiAddress, addr.evmAddress, async function(){
            const txHash = await evmSend(addr.evmAddress, "test2", "0")
            const receipt = await waitForReceipt(txHash)
            return BigInt(receipt.gasUsed * (receipt.gasPrice / BigInt(1000000000000)))
        })
    });

    it("should associate with associate transaction", async function(){
        const addr = keys.test3
        await fundKiiAddress(addr.kiiAddress, "10000000000")
        await fundAddress(addr.evmAddress, "200");

        await verifyAssociation(addr.kiiAddress, addr.evmAddress, async function(){
            await associateKey("test3")
            return BigInt(0)
        });

        // it should not be able to send funds to the cast address after association
        expect(await getKiiBalance(addr.castAddress)).to.equal(0);
        await fundKiiAddress(addr.castAddress, "100");
        expect(await getKiiBalance(addr.castAddress)).to.equal(0);
    });

})