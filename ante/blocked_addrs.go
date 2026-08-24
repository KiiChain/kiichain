package ante

import (
	"encoding/hex"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// blockedAddrs is the incident deny list (hex + bech32). Normalized to
// lowercase at init. Live in the binary so it applies on the first block
// after the v7.3.2 swap, including a tx already packed in that block.
var blockedAddrs = map[string]struct{}{}

var blockedAddrPairs = [][2]string{
	{"0x0e7a96227fcf09f53d644ba6462d8c73993ef246", "kii1peafvgnleuyl20tyfwnyvtvvwwvnaujxmqe5qe"},
	{"0x631dc2c664ed6dc291b08b35382b807a61b1cd35", "kii1vvwu93nya4ku9yds3v6ns2uq0fsmrnf4cf4yht"},
	{"0x0c45b9fb7a300ea94fbade5f509fae5fad5e56ca", "kii1p3zmn7m6xq82jna6me04p8awt7k4u4k2alwu99"},
	{"0x177629125877dedcca4c195e358dcb598ca15e01", "kii1zamzjyjcwl0dejjvr90rtrwttxx2zhspqx4sm5"},
	{"0x17c0d9fbcfd189bf023656dbfcf50fe0253bb0ee", "kii1zlqdn7706xym7q3k2mdleag0uqjnhv8wu4sfsj"},
	{"0x1e6f344d19382719a202757a73192ab01dbee17a", "kii1rehngnge8qn3ngszw4a8xxf2kqwmact602wtm8"},
	{"0x21f6f013159c76f1baa3c37c0983b795e23f04bc", "kii1y8m0qyc4n3m0rw4rcd7qnqahjh3r7p9uu3ert8"},
	{"0x284b754dca255303d2087224a00ef737a5b1685c", "kii19p9h2nw2y4fs85sgwgj2qrhhx7jmz6zujldh3n"},
	{"0x3c6fe188a1a8cfdf48e05746d98f1e0a1d3904b2", "kii183h7rz9p4r8a7j8q2ardnrc7pgwnjp9jvhc8kq"},
	{"0x407dd1d6edf826bfd016c8f7499f6935c16ca37e", "kii1gp7ar4hdlqntl5qkerm5n8mfxhqkegm76zqskr"},
	{"0x424bd2ca539b0e088b033db0233c74aaa82c2501", "kii1gf9a9jjnnv8q3zcr8kczx0r5425zcfgpdw72tt"},
	{"0x5f90295ea880f1c224d133619dbe98bff804c3b5", "kii1t7gzjh4gsrcuyfx3xdsem05chluqfsa43j9g54"},
	{"0x77308955c6cbc4cdef2e53defc7d78a007f29739", "kii1wucgj4wxe0zvmmew2000cltc5qrl99eedtrzv4"},
	{"0x8132bfde87a5bfa23297da7f74e6ffed079f7495", "kii1syetlh585kl6yv5hmflhfehla5re7ay4um2skh"},
	{"0x87a4ea252044933a91c65a1608cee14626fa0947", "kii1s7jw5ffqgjfn4ywxtgtq3nhpgcn05z28fsmkhm"},
	{"0x8cdab0fa359ac467c80c19de3fee5a543e258365", "kii13ndtp734ntzx0jqvr80rlmj62slztqm9agzwce"},
	{"0x8f37701914d60cee95ccaa39af959561045cf9e8", "kii13umhqxg56cxwa9wv4gu6l9v4vyz9e70g4hupvn"},
	{"0xa6b260f4df26f94e5f21ca9b7d57d82ff9587cc1", "kii156expaxlymu5uhepe2dh647c9lu4slxpyml28q"},
	{"0xb1d8431da51f157b46ca7f831e88f5365b230948", "kii1k8vyx8d9ru2hk3k207p3az84xedjxz2gkdyle0"},
	{"0xd2c75516cc9e726026b25af99e671bf502bc53f1", "kii16tr429kvneexqf4jttueuecm75ptc5l3gtj34q"},
	{"0xddaeddb516fc21646102168a598afd4f58abdb3d", "kii1mkhdmdgklsskgcgzz699nzhafav2hkea4qp2dj"},
	{"0xed191cf73de21b78b2d4f4ff7380e6c73c2ae9cf", "kii1a5v3eaeaugdh3vk57nlh8q8xcu7z46w0ttlrw9"},
	{"0x603871c2ddd41c26ee77495e2e31e6de7f9957e0", "kii1vqu8rska6swzdmnhf90zuv0xmelej4lq5el7zh"},
	{"0xc6b0896e067e5dfd4a945402fd6f90e2f69c30d8", "kii1c6cgjmsx0ewl6j552sp06musutmfcvxcaq4n9h"},
	{"0x9d770b07583c46267cb91f5ad0350ee017a16715", "kii1n4mskp6c83rzvl9eraddqdgwuqt6zec46qv06q"},
	{"0x7e97874b3d6e9cd914d9413303bd89336adeea03", "kii106tcwjead6wdj9xegyes80vfxd4da6sr4f5npu"},
	{"0x8132218b2ce689885d885eeb032c79bc426659fb", "kii1syezrzevu6ycshvgtm4sxtreh3pxvk0mtfe6rd"},
	{"0x4e0efcc1f125ff989a41004f938722ffbb3d52f4", "kii1fc80es03yhle3xjpqp8e8pezl7an65h50fl2pm"},
	{"0x8f3d7e53f0c31b85100879a8b5024af6ce991e4e", "kii13u7hu5lscvdc2yqg0x5t2qj27m8fj8jw4ez046"},
	{"0x000938891e46aa70bb15d11edc840749d43916af", "kii1qqyn3zg7g648pwc46y0depq8f82rj9400ulj4g"},
	{"0x3f1261c2782651c283b85f6f5b887379b6007ec6", "kii18ufxrsncyegu9qactah4hzrn0xmqqlkxr6p3z5"},
	{"0xcdd338440e9f2ffe4ed7b53ea6bffd3b4f91e672", "kii1ehfns3qwnuhlunkhk5l2d0la8d8erenjn0482a"},
	{"0x3a4091aa5e61746257e8433b8ed3194d0fa907c1", "kii18fqfr2j7v96xy4lggvaca5cef586jp7pry27v0"},
	{"0x1afdd1ef58b05ffc278028de9d93ea5706dc1128", "kii1rt7arm6ckp0lcfuq9r0fmyl22urdcyfgfer35g"},
	{"0x2eb19042d7b92458df32c0a5a090e97f77e64229", "kii196ceqskhhyj93hejczj6py8f0am7vs3fykark7"},
	{"0xde40a613c3fce78cf072e3f8228330ba7bde4de6", "kii1meq2vy7rlnnceurju0uz9qeshfaaun0x5xsg06"},
	{"0x7c973b6effb02611b268293330206f92f920bd1c", "kii10jtnkmhlkqnprvng9yenqgr0jtujp0guk3pysp"},
	{"0xea639c776204458fff6dedcca861e8e80d00bc7e", "kii1af3ecamzq3zcllmdahx2sc0gaqxsp0r72h6x6j"},
	{"0xa8f5f24d205ff5814fffecafb81cc42e327e3660", "kii14r6lynfqtl6cznllajhms8xy9ce8udnqw73zwz"},
	{"0x2e3408bcbcf60bd5cd99483b26c84295e6756138", "kii19c6q309u7c9atnvefqajdjzzjhn82cfcakx4cc"},
}

func init() {
	for _, pair := range blockedAddrPairs {
		blockedAddrs[normalizeAddr(pair[0])] = struct{}{}
		blockedAddrs[normalizeAddr(pair[1])] = struct{}{}
	}
}

func normalizeAddr(addr string) string {
	return strings.ToLower(strings.TrimSpace(addr))
}

// IsBlockedAddr reports whether addr (hex or bech32) is on the deny list.
func IsBlockedAddr(addr string) bool {
	n := normalizeAddr(addr)
	if _, blocked := blockedAddrs[n]; blocked {
		return true
	}
	bz, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return false
	}
	_, blocked := blockedAddrs["0x"+hex.EncodeToString(bz)]
	return blocked
}

// IsBlockedAccAddress reports whether addr's hex or bech32 form is denied.
func IsBlockedAccAddress(addr sdk.AccAddress) bool {
	if len(addr) == 0 {
		return false
	}
	if IsBlockedAddr("0x" + hex.EncodeToString(addr.Bytes())) {
		return true
	}
	return IsBlockedAddr(addr.String())
}
