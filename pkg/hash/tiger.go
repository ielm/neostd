package hash

import (
	"encoding/binary"
	"math/bits"
)

const (
	tigerBlockSize  = 64
	tigerDigestSize = 24
	tigerRounds     = 3
)

// TigerHasher implements the Tiger hash algorithm
type TigerHasher[K any] struct {
	a, b, c uint64
	x       [tigerBlockSize]byte
	nx      int
	len     uint64
}

// NewTigerHasher creates a new TigerHasher
func NewTigerHasher[K any]() *TigerHasher[K] {
	return &TigerHasher[K]{
		a: 0x0123456789ABCDEF,
		b: 0xFEDCBA9876543210,
		c: 0xF096A5B4C3B2E187,
	}
}

// Hash computes the Tiger hash of the given key
func (t *TigerHasher[K]) Hash(key K) ([]byte, error) {
	data, err := keyToBytes(key)
	if err != nil {
		return nil, err
	}

	t.Reset()
	t.Write(data)
	return t.Sum(nil), nil
}

// Write adds more data to the running hash
func (t *TigerHasher[K]) Write(p []byte) (n int, err error) {
	n = len(p)
	t.len += uint64(n)

	if t.nx > 0 {
		n := copy(t.x[t.nx:], p)
		t.nx += n
		if t.nx == tigerBlockSize {
			t.compress(t.x[:])
			t.nx = 0
		}
		p = p[n:]
	}

	if len(p) >= tigerBlockSize {
		n := len(p) &^ (tigerBlockSize - 1)
		t.compress(p[:n])
		p = p[n:]
	}

	if len(p) > 0 {
		t.nx = copy(t.x[:], p)
	}

	return
}

// Sum appends the current hash to b and returns the resulting slice
func (t *TigerHasher[K]) Sum(b []byte) []byte {
	t0 := *t
	hash := t0.checkSum()
	return append(b, hash[:]...)
}

// Reset resets the hash to its initial state
func (t *TigerHasher[K]) Reset() {
	t.a = 0x0123456789ABCDEF
	t.b = 0xFEDCBA9876543210
	t.c = 0xF096A5B4C3B2E187
	t.nx = 0
	t.len = 0
}

// checkSum generates the final hash value
func (t *TigerHasher[K]) checkSum() [tigerDigestSize]byte {
	len := t.len
	// Padding
	t.x[t.nx] = 0x01
	t.nx++
	if t.nx > 56 {
		for i := t.nx; i < tigerBlockSize; i++ {
			t.x[i] = 0
		}
		t.compress(t.x[:])
		t.nx = 0
	}
	for i := t.nx; i < 56; i++ {
		t.x[i] = 0
	}
	binary.LittleEndian.PutUint64(t.x[56:], len<<3)
	t.compress(t.x[:])

	var digest [tigerDigestSize]byte
	binary.LittleEndian.PutUint64(digest[0:], t.a)
	binary.LittleEndian.PutUint64(digest[8:], t.b)
	binary.LittleEndian.PutUint64(digest[16:], t.c)
	return digest
}

// compress is the core compression function of Tiger
func (t *TigerHasher[K]) compress(block []byte) {
	var x [8]uint64
	for i := 0; i < 8; i++ {
		x[i] = binary.LittleEndian.Uint64(block[i*8:])
	}

	aa, bb, cc := t.a, t.b, t.c

	for i := 0; i < tigerRounds; i++ {
		if i != 0 {
			x[0] -= x[7] ^ 0xA5A5A5A5A5A5A5A5
			x[1] ^= x[0]
			x[2] += x[1]
			x[3] -= x[2] ^ ((^x[1]) << 19)
			x[4] ^= x[3]
			x[5] += x[4]
			x[6] -= x[5] ^ ((^x[4]) >> 23)
			x[7] ^= x[6]
			x[0] += x[7]
			x[1] -= x[0] ^ ((^x[7]) << 19)
			x[2] ^= x[1]
			x[3] += x[2]
			x[4] -= x[3] ^ ((^x[2]) >> 23)
			x[5] ^= x[4]
			x[6] += x[5]
			x[7] -= x[6] ^ 0x0123456789ABCDEF
		}

		aa, bb, cc = t.round(aa, bb, cc, x[0], x[1], x[2], x[3], x[4], x[5], x[6], x[7])
		aa, bb, cc = cc, aa, bb
	}

	t.a ^= aa
	t.b = bb - t.b
	t.c += cc
}

// round performs a single round of the Tiger hash function
func (t *TigerHasher[K]) round(a, b, c, x0, x1, x2, x3, x4, x5, x6, x7 uint64) (uint64, uint64, uint64) {
	c ^= x0
	a -= t0[byte(c)] ^ t1[byte(c>>16)] ^ t2[byte(c>>32)] ^ t3[byte(c>>48)]
	b += t3[byte(c>>8)] ^ t2[byte(c>>24)] ^ t1[byte(c>>40)] ^ t0[byte(c>>56)]
	b *= 5

	a ^= x1
	b -= t0[byte(a)] ^ t1[byte(a>>16)] ^ t2[byte(a>>32)] ^ t3[byte(a>>48)]
	c += t3[byte(a>>8)] ^ t2[byte(a>>24)] ^ t1[byte(a>>40)] ^ t0[byte(a>>56)]
	c *= 5

	b ^= x2
	c -= t0[byte(b)] ^ t1[byte(b>>16)] ^ t2[byte(b>>32)] ^ t3[byte(b>>48)]
	a += t3[byte(b>>8)] ^ t2[byte(b>>24)] ^ t1[byte(b>>40)] ^ t0[byte(b>>56)]
	a *= 5

	c ^= x3
	a -= t0[byte(c)] ^ t1[byte(c>>16)] ^ t2[byte(c>>32)] ^ t3[byte(c>>48)]
	b += t3[byte(c>>8)] ^ t2[byte(c>>24)] ^ t1[byte(c>>40)] ^ t0[byte(c>>56)]
	b *= 5

	a ^= x4
	b -= t0[byte(a)] ^ t1[byte(a>>16)] ^ t2[byte(a>>32)] ^ t3[byte(a>>48)]
	c += t3[byte(a>>8)] ^ t2[byte(a>>24)] ^ t1[byte(a>>40)] ^ t0[byte(a>>56)]
	c *= 5

	b ^= x5
	c -= t0[byte(b)] ^ t1[byte(b>>16)] ^ t2[byte(b>>32)] ^ t3[byte(b>>48)]
	a += t3[byte(b>>8)] ^ t2[byte(b>>24)] ^ t1[byte(b>>40)] ^ t0[byte(b>>56)]
	a *= 5

	c ^= x6
	a -= t0[byte(c)] ^ t1[byte(c>>16)] ^ t2[byte(c>>32)] ^ t3[byte(c>>48)]
	b += t3[byte(c>>8)] ^ t2[byte(c>>24)] ^ t1[byte(c>>40)] ^ t0[byte(c>>56)]
	b *= 5

	a ^= x7
	b -= t0[byte(a)] ^ t1[byte(a>>16)] ^ t2[byte(a>>32)] ^ t3[byte(a>>48)]
	c += t3[byte(a>>8)] ^ t2[byte(a>>24)] ^ t1[byte(a>>40)] ^ t0[byte(a>>56)]
	c *= 5

	return a, b, c
}

// t0, t1, t2, t3 are the S-boxes used in Tiger
var t0, t1, t2, t3 [256]uint64

func init() {
	// Initialize S-boxes
	for i := 0; i < 256; i++ {
		t0[i] = tigerT1[i]
		t1[i] = bits.RotateLeft64(tigerT1[i], 23)
		t2[i] = bits.RotateLeft64(tigerT1[i], 46)
		t3[i] = bits.RotateLeft64(tigerT1[i], 5)
	}
}

// tigerT1 is the first S-box used in Tiger
var tigerT1 = [256]uint64{
	0x02AAB17CF7E90C5E, 0xAC424B03E243A8EC, 0x72CD5BE30DD5FCD3, 0x6D019B93F6F97F3A,
	0xCD9978FFD21F9193, 0x7573A1C9708029E2, 0xB164326B922A83C3, 0x46883EEE04915870,
	0xEAACE3057103ECE6, 0xC54169B808A3535C, 0x4CE754918DDEC47C, 0x0AA2F4DFDC0DF40C,
	0x10B76F18A74DBEFA, 0xC6CCB6235AD1AB6A, 0x13726121572FE2FF, 0x1A488C6F199D921E,
	0x4BC9F9F4DA0007CA, 0x26F5E6F6E85241C7, 0x859079DBEA5947B6, 0x4F1885C5C99E8C92,
	0xD78E761EA96F864B, 0x8E36428C52B5C17D, 0x69CF6827373063C1, 0xB607C93D9BB4C56E,
	0x7D820E760E76B5EA, 0x645C9CC6F07FDC42, 0xBF38A078243342E0, 0x5F6B343C9D2E7D04,
	0xF2C28AEB600B0EC6, 0x6C0ED85F7254BCAC, 0x71592281A4DB4FE5, 0x1967FA69CE0FED9F,
	0xFD5293F8B96545DB, 0xC879E9D7F2A7600B, 0x860248920193194E, 0xA4F9533B2D9CC0B3,
	0x9053836C15957613, 0xDB6DCF8AFC357BF1, 0x18BEEA7A7A370F57, 0x037117CA50B99066,
	0x6AB30A9774424A35, 0xF4E92F02E325249B, 0x7739DB07061CCAE1, 0xD8F3B49CECA42A05,
	0xBD56BE3F51382F73, 0x45FAED5843B0BB28, 0x1C813D5C11BF1F83, 0x8AF0E4B6D75FA169,
	0x33EE18A487AD9999, 0x3C26E8EAB1C94410, 0xB510102BC0A822F9, 0x141EEF310CE6123B,
	0xFC65B90059DDB154, 0xE0158640C5E0E607, 0x884E079826C3A3CF, 0x930D0D9523C535FD,
	0x35638D754E9A2B00, 0x4085FCCF40469DD5, 0xC4B17AD28BE23A4C, 0xCAB2F0FC6A3E6A2E,
	0x2860971A6B943FCD, 0x3DDE6EE212E30446, 0x6222F32AE01765AE, 0x5D550BB5478308FE,
	0xA9EFA98DA0EDA22A, 0xC351A71686C40DA7, 0x1105586D9C867C84, 0xDCFFEE85FDA22853,
	0xCCFBD0262C5EEF76, 0xBAF294CB8990D201, 0xE69464F52AFAD975, 0x94B013AFDF133E14,
	0x06A7D1A32823C958, 0x6F95FE5130F61119, 0xD92AB34E462C06C0, 0xED7BDE33887C71D2,
	0x79746D6E6518393E, 0x5BA419385D713329, 0x7C1BA6B948A97564, 0x31987C197BFDAC67,
	0xDE6C23C44B053D02, 0x581C49FED002D64D, 0xDD474D6338261571, 0xAA4546C3E473D062,
	0x928FCE349455F860, 0x48161BBACAAB94D9, 0x63912430770E6F68, 0x6EC8A5E602C6641C,
	0x87282515337DDD2B, 0x2CDA6B42034B701B, 0xB03D37C181CB096D, 0xE108438266C71C6F,
	0x2B3180C7EB51B255, 0xDF92B82F96C08BBC, 0x5C68C8C0A632F3BA, 0x5504CC861C3D0556,
	0xABBFA4E55FB26B8F, 0x41848B0AB3BACEB4, 0xB334A273AA445D32, 0xBCA696F0A85AD881,
	0x24F6EC65B528D56C, 0x0CE1512E90F4524A, 0x4E9DD79D5506D35A, 0x258905FAC6CE9779,
	0x2019295B3E109B33, 0xF8A9478B73A054CC, 0x2924F2F934417EB0, 0x3993357D536D1BC4,
	0x38A81AC21DB6FF8B, 0x47C4FBF17D6016BF, 0x1E0FAADD7667E3F5, 0x7ABCFF62938BEB96,
	0xA78DAD948FC179C9, 0x8F1F98B72911E50D, 0x61E48EAE27121A91, 0x4D62F7AD31859808,
	0xECEBA345EF5CEAEB, 0xF5CEB25EBC9684CE, 0xF633E20CB7F76221, 0xA32CDF06AB8293E4,
	0x985A202CA5EE2CA4, 0xCF0B8447CC8A8FB1, 0x9F765244979859A3, 0xA8D516B1A1240017,
	0x0BD7BA3EBB5DC726, 0xE54BCA55B86ADB39, 0x1D7A3AFD6C478063, 0x519EC608E7669EDD,
	0x0E5715A2D149AA23, 0x177D4571848FF194, 0xEEB55F3241014C22, 0x0F5E5CA13A6E2EC2,
	0x8029927B75F5C361, 0xAD139FABC3D6E436, 0x0D5DF1A94CCF402F, 0x3E8BD948BEA5DFC8,
	0xA5A0D357BD3FF77E, 0xA2D12E251F74F645, 0x66FD9E525E81A082, 0x2E0C90CE7F687A49,
	0xC2E8BCBEBA973BC5, 0x000001BCE509745F, 0x423777BBE6DAB3D6, 0xD1661C7EAEF06EB5,
	0xA1781F354DAACFD8, 0x2D11284A2B16AFFC, 0xF1FC4F67FA891D1F, 0x73ECC25DCB920ADA,
	0xAE610C22C2A12651, 0x96E0A810D356B78A, 0x5A9A381F2FE7870F, 0xD5AD62EDE94E5530,
	0xD225E5E8368D1427, 0x65977B70C7AF4631, 0x99F889B2DE39D74F, 0x233F30BF54E1D143,
	0x9A9675D3D9A63C97, 0x5470554FF334F9A8, 0x166ACB744A4F5688, 0x70C74CAAB2E4AEAD,
	0xF0D091646F294D12, 0x57B82A89684031D1, 0xEFD95A5A61BE0B6B, 0x2FBD12E969F2F29A,
	0x9BD37013FEFF9FE8, 0x3F9B0404D6085A06, 0x4940C1F3166CFE15, 0x09542C4DCDF3DEFB,
	0xB4C5218385CD5CE3, 0xC935B7DC4462A641, 0x3417F8A68ED3B63F, 0xB80959295B215B40,
	0xF99CDAEF3B8C8572, 0x018C0614F8FCB95D, 0x1B14ACCD1A3ACDF3, 0x84D471F200BB732D,
	0xC1A3110E95E8DA16, 0x430A7220BF1A82B8, 0xB77E090D39DF210E, 0x5EF4BD9F3CD05E9D,
	0x9D4FF6DA7E57A444, 0xDA1D60E183D4A5F8, 0xB287C38417998E47, 0xFE3EDC121BB31886,
	0xC7FE3CCC980CCBEF, 0xE46FB590189BFD03, 0x3732FD469A4C57DC, 0x7EF700A07CF1AD65,
	0x59C64468A31D8859, 0x762FB0B4D45B61F6, 0x155BAED099047718, 0x68755E4C3D50BAA6,
	0xE9214E7F22D8B4DF, 0x2ADDBF532EAC95F4, 0x32AE3909B4BD0109, 0x834DF537B08E3450,
	0xFA209DA84220728D, 0x9E691D9B9EFE23F7, 0x0446D288C4AE8D7F, 0x7B4CC524E169785B,
	0x21D87F0135CA1385, 0xCEBB400F137B8AA5, 0x272E2B66580796BE, 0x3612264125C2B0DE,
	0x057702BDAD1EFBB2, 0xD4BABB8EACF84BE9, 0x91583139641BC67B, 0x8BDC2DE08036E024,
	0x603C8156F49F68ED, 0xF7D236F7DBEF5111, 0x9727C4598AD21E80, 0xA08A0896670A5FD7,
	0xCB4A8F4309EBA9CB, 0x81AF564B0F7036A1, 0xC0B99AA778199ABD, 0x959F1EC83FC8E952,
	0x8C505077794A81B9, 0x3ACAAF8F056338F0, 0x07B43F50627A6778, 0x4A44AB49F5ECCC77,
	0x3BC3D6E4B679EE98, 0x9CC0D4D1CF14108C, 0x4406C00B206BC8A0, 0x82A18854C8D72D89,
	0x67E366B35C3C432C, 0xB923DD61102B37F2, 0x56AB2779D884271D, 0xBE83E1B0FF1525AF,
	0xFB7C65D4217E49A9, 0x6BDBE0E76D48E7D4, 0x08DF828745D9179E, 0x22EA6A9ADD53BD34,
	0xE36E141C5622200A, 0x7F805D1B8CB750EE, 0xAFE5C7A59F58E837, 0xE27F996A4FB1C23C,
	0xD3867DFB0775F0D0, 0xD0E673DE6E88891A, 0x123AEB9EAFB86C25, 0x30F1D5D5C145B895,
	0xBB434A2DEE7269E7, 0x78CB67ECF931FA38, 0xF33B0372323BBF9C, 0x52D66336FB279C74,
	0x505F33AC0AFB4EAA, 0xE8A5CD99A2CCE187, 0x534974801E2D30BB, 0x8D2D5711D5876D90,
	0x1F1A412891BC038E, 0xD6E2E71D82E56648, 0x74036C3A497732B7, 0x89B67ED96361F5AB,
	0xFFED95D8F1EA02A2, 0xE72B3BD61464D43D, 0xA6300F170BDC4820, 0xEBC18760ED78A77A,
}

// Ensure TigerHasher implements the Hasher interface
var _ Hasher[any] = (*TigerHasher[any])(nil)
