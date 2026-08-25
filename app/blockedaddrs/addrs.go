package blockedaddrs

import (
	"encoding/hex"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AttackerAddrs is the 22 Aug 2026 incident's frozen wallets (bech32) — the
// same list app/upgrades/v7_4 sweeps in its recovery handler.
var AttackerAddrs = []string{
	"kii1peafvgnleuyl20tyfwnyvtvvwwvnaujxmqe5qe",
	"kii1vvwu93nya4ku9yds3v6ns2uq0fsmrnf4cf4yht",
	"kii1p3zmn7m6xq82jna6me04p8awt7k4u4k2alwu99",
	"kii1zamzjyjcwl0dejjvr90rtrwttxx2zhspqx4sm5",
	"kii1zlqdn7706xym7q3k2mdleag0uqjnhv8wu4sfsj",
	"kii1rehngnge8qn3ngszw4a8xxf2kqwmact602wtm8",
	"kii1y8m0qyc4n3m0rw4rcd7qnqahjh3r7p9uu3ert8",
	"kii19p9h2nw2y4fs85sgwgj2qrhhx7jmz6zujldh3n",
	"kii183h7rz9p4r8a7j8q2ardnrc7pgwnjp9jvhc8kq",
	"kii1gp7ar4hdlqntl5qkerm5n8mfxhqkegm76zqskr",
	"kii1gf9a9jjnnv8q3zcr8kczx0r5425zcfgpdw72tt",
	"kii1t7gzjh4gsrcuyfx3xdsem05chluqfsa43j9g54",
	"kii1wucgj4wxe0zvmmew2000cltc5qrl99eedtrzv4",
	"kii1syetlh585kl6yv5hmflhfehla5re7ay4um2skh",
	"kii1s7jw5ffqgjfn4ywxtgtq3nhpgcn05z28fsmkhm",
	"kii13ndtp734ntzx0jqvr80rlmj62slztqm9agzwce",
	"kii13umhqxg56cxwa9wv4gu6l9v4vyz9e70g4hupvn",
	"kii156expaxlymu5uhepe2dh647c9lu4slxpyml28q",
	"kii1k8vyx8d9ru2hk3k207p3az84xedjxz2gkdyle0",
	"kii16tr429kvneexqf4jttueuecm75ptc5l3gtj34q",
	"kii1mkhdmdgklsskgcgzz699nzhafav2hkea4qp2dj",
	"kii1a5v3eaeaugdh3vk57nlh8q8xcu7z46w0ttlrw9",
}

// IsBlockedAccAddress reports whether addr is one of AttackerAddrs.
func IsBlockedAccAddress(addr sdk.AccAddress) bool {
	if len(addr) == 0 {
		return false
	}
	bech32 := addr.String()
	for _, blocked := range AttackerAddrs {
		if bech32 == blocked {
			return true
		}
	}
	return false
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
