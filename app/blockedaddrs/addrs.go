package blockedaddrs

import (
	"encoding/hex"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AttackerAddrs is the 22 Aug 2026 incident's frozen wallets: bech32 -> hex
var AttackerAddrs = map[string]string{
	"kii1peafvgnleuyl20tyfwnyvtvvwwvnaujxmqe5qe": "0x0e7a96227fcf09f53d644ba6462d8c73993ef246",
	"kii1vvwu93nya4ku9yds3v6ns2uq0fsmrnf4cf4yht": "0x631dc2c664ed6dc291b08b35382b807a61b1cd35",
	"kii1p3zmn7m6xq82jna6me04p8awt7k4u4k2alwu99": "0x0c45b9fb7a300ea94fbade5f509fae5fad5e56ca",
	"kii1zamzjyjcwl0dejjvr90rtrwttxx2zhspqx4sm5": "0x177629125877dedcca4c195e358dcb598ca15e01",
	"kii1zlqdn7706xym7q3k2mdleag0uqjnhv8wu4sfsj": "0x17c0d9fbcfd189bf023656dbfcf50fe0253bb0ee",
	"kii1rehngnge8qn3ngszw4a8xxf2kqwmact602wtm8": "0x1e6f344d19382719a202757a73192ab01dbee17a",
	"kii1y8m0qyc4n3m0rw4rcd7qnqahjh3r7p9uu3ert8": "0x21f6f013159c76f1baa3c37c0983b795e23f04bc",
	"kii19p9h2nw2y4fs85sgwgj2qrhhx7jmz6zujldh3n": "0x284b754dca255303d2087224a00ef737a5b1685c",
	"kii183h7rz9p4r8a7j8q2ardnrc7pgwnjp9jvhc8kq": "0x3c6fe188a1a8cfdf48e05746d98f1e0a1d3904b2",
	"kii1gp7ar4hdlqntl5qkerm5n8mfxhqkegm76zqskr": "0x407dd1d6edf826bfd016c8f7499f6935c16ca37e",
	"kii1gf9a9jjnnv8q3zcr8kczx0r5425zcfgpdw72tt": "0x424bd2ca539b0e088b033db0233c74aaa82c2501",
	"kii1t7gzjh4gsrcuyfx3xdsem05chluqfsa43j9g54": "0x5f90295ea880f1c224d133619dbe98bff804c3b5",
	"kii1wucgj4wxe0zvmmew2000cltc5qrl99eedtrzv4": "0x77308955c6cbc4cdef2e53defc7d78a007f29739",
	"kii1syetlh585kl6yv5hmflhfehla5re7ay4um2skh": "0x8132bfde87a5bfa23297da7f74e6ffed079f7495",
	"kii1s7jw5ffqgjfn4ywxtgtq3nhpgcn05z28fsmkhm": "0x87a4ea252044933a91c65a1608cee14626fa0947",
	"kii13ndtp734ntzx0jqvr80rlmj62slztqm9agzwce": "0x8cdab0fa359ac467c80c19de3fee5a543e258365",
	"kii13umhqxg56cxwa9wv4gu6l9v4vyz9e70g4hupvn": "0x8f37701914d60cee95ccaa39af959561045cf9e8",
	"kii156expaxlymu5uhepe2dh647c9lu4slxpyml28q": "0xa6b260f4df26f94e5f21ca9b7d57d82ff9587cc1",
	"kii1k8vyx8d9ru2hk3k207p3az84xedjxz2gkdyle0": "0xb1d8431da51f157b46ca7f831e88f5365b230948",
	"kii16tr429kvneexqf4jttueuecm75ptc5l3gtj34q": "0xd2c75516cc9e726026b25af99e671bf502bc53f1",
	"kii1mkhdmdgklsskgcgzz699nzhafav2hkea4qp2dj": "0xddaeddb516fc21646102168a598afd4f58abdb3d",
	"kii1a5v3eaeaugdh3vk57nlh8q8xcu7z46w0ttlrw9": "0xed191cf73de21b78b2d4f4ff7380e6c73c2ae9cf",
}

// SortedAttackerAddresses returns AttackerAddrs' keys in a fixed,
// deterministic order. Range over AttackerAddrs directly only where
// iteration order can't matter (e.g. IsBlockedAccAddress's lookup) — Go
// randomizes map iteration order per process, and code that walks this list
// from a consensus-critical upgrade handler must behave identically, in the
// same order, on every validator.
func SortedAttackerAddresses() []string {
	addrs := make([]string, 0, len(AttackerAddrs))
	for addr := range AttackerAddrs {
		addrs = append(addrs, addr)
	}
	sort.Strings(addrs)
	return addrs
}

// IsBlockedAccAddress reports whether addr is one of AttackerAddrs.
func IsBlockedAccAddress(addr sdk.AccAddress) bool {
	if len(addr) == 0 {
		return false
	}
	_, blocked := AttackerAddrs[addr.String()]
	return blocked
}

// IsBlockedAddr reports whether addr (hex or bech32) is one of AttackerAddrs.
func IsBlockedAddr(addr string) bool {
	if accAddr, err := sdk.AccAddressFromBech32(addr); err == nil {
		return IsBlockedAccAddress(accAddr)
	}

	s := strings.TrimSpace(addr)
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s = s[2:]
	}
	bz, err := hex.DecodeString(s)
	if err != nil {
		return false
	}
	return IsBlockedAccAddress(sdk.AccAddress(bz))
}
