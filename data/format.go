package data

import (
	"encoding/binary"
	"strings"
)

type NodeType uint8
type NodeFormat uint8
type HashPrefix uint32
type LedgerNamespace uint16

const (
	// Hash Prefixes
	HP_TRANSACTION_ID   HashPrefix = 0x54584E00 // 'TXN' transaction
	HP_TRANSACTION_NODE HashPrefix = 0x534E4400 // 'SND' transaction plus metadata (probably should have been TND!)
	HP_LEAF_NODE        HashPrefix = 0x4D4C4E00 // 'MLN' account state
	HP_INNER_NODE       HashPrefix = 0x4D494E00 // 'MIN' inner node in tree
	HP_LEDGER_MASTER    HashPrefix = 0x4C575200 // 'LWR' ledger master data for signing (probably should have been LGR!)
	HP_TRANSACTION_SIGN HashPrefix = 0x53545800 // 'STX' inner transaction to sign
	HP_VALIDATION       HashPrefix = 0x56414C00 // 'VAL' validation for signing
	HP_PROPOSAL         HashPrefix = 0x50525000 // 'PRP' proposal for signing

	// Node Types
	NT_UNKNOWN          NodeType = 0
	NT_LEDGER           NodeType = 1
	NT_TRANSACTION      NodeType = 2
	NT_ACCOUNT_NODE     NodeType = 3
	NT_TRANSACTION_NODE NodeType = 4

	// Node Formats
	NF_PREFIX NodeFormat = 1
	NF_HASH   NodeFormat = 2
	NF_WIRE   NodeFormat = 3

	// Ledger index NameSpaces
	NS_ACCOUNT         LedgerNamespace = 'a'
	NS_DIRECTORY_NODE  LedgerNamespace = 'd'
	NS_RIPPLE_STATE    LedgerNamespace = 'r'
	NS_OFFER           LedgerNamespace = 'o' // Entry for an offer
	NS_OWNER_DIRECTORY LedgerNamespace = 'O' // Directory of things owned by an account
	NS_BOOK_DIRECTORY  LedgerNamespace = 'B' // Directory of order books
	NS_SKIP_LIST       LedgerNamespace = 's'
	NS_AMENDMENT       LedgerNamespace = 'f'
	NS_FEE             LedgerNamespace = 'e'
)

var nodeTypes = [...]string{
	NT_UNKNOWN:          "Unknown",
	NT_LEDGER:           "Ledger",
	NT_TRANSACTION:      "Transaction",
	NT_ACCOUNT_NODE:     "Account Node",
	NT_TRANSACTION_NODE: "Transaction Node",
}

type NodeHeader struct {
	LedgerIndex uint32
	_           uint32 //padding for repeated LedgerIndex
	NodeType    NodeType
}

type enc struct {
	typ, field uint8
}

const (
	ST_UINT16    uint8 = 1
	ST_UINT32    uint8 = 2
	ST_UINT64    uint8 = 3
	ST_HASH128   uint8 = 4
	ST_HASH256   uint8 = 5
	ST_AMOUNT    uint8 = 6
	ST_VL        uint8 = 7
	ST_ACCOUNT   uint8 = 8
	ST_OBJECT    uint8 = 14
	ST_ARRAY     uint8 = 15
	ST_UINT8     uint8 = 16
	ST_HASH160   uint8 = 17
	ST_PATHSET   uint8 = 18
	ST_VECTOR256 uint8 = 19
)

var encodings = map[enc]string{
	// 16-bit unsigned integers (common)
	enc{ST_UINT16, 1}: "LedgerEntryType",
	enc{ST_UINT16, 2}: "TransactionType",
	// 32-bit unsigned integers (common)
	enc{ST_UINT32, 2}:  "Flags",
	enc{ST_UINT32, 3}:  "SourceTag",
	enc{ST_UINT32, 4}:  "Sequence",
	enc{ST_UINT32, 5}:  "PreviousTxnLgrSeq",
	enc{ST_UINT32, 6}:  "LedgerSequence",
	enc{ST_UINT32, 7}:  "CloseTime",
	enc{ST_UINT32, 8}:  "ParentCloseTime",
	enc{ST_UINT32, 9}:  "SigningTime",
	enc{ST_UINT32, 10}: "Expiration",
	enc{ST_UINT32, 11}: "TransferRate",
	enc{ST_UINT32, 12}: "WalletSize",
	enc{ST_UINT32, 13}: "OwnerCount",
	enc{ST_UINT32, 14}: "DestinationTag",
	// 32-bit unsigned integers (uncommon)
	enc{ST_UINT32, 16}: "HighQualityIn",
	enc{ST_UINT32, 17}: "HighQualityOut",
	enc{ST_UINT32, 18}: "LowQualityIn",
	enc{ST_UINT32, 19}: "LowQualityOut",
	enc{ST_UINT32, 20}: "QualityIn",
	enc{ST_UINT32, 21}: "QualityOut",
	enc{ST_UINT32, 22}: "StampEscrow",
	enc{ST_UINT32, 23}: "BondAmount",
	enc{ST_UINT32, 24}: "LoadFee",
	enc{ST_UINT32, 25}: "OfferSequence",
	enc{ST_UINT32, 26}: "FirstLedgerSequence",
	enc{ST_UINT32, 27}: "LastLedgerSequence",
	enc{ST_UINT32, 28}: "TransactionIndex",
	enc{ST_UINT32, 29}: "OperationLimit",
	enc{ST_UINT32, 30}: "ReferenceFeeUnits",
	enc{ST_UINT32, 31}: "ReserveBase",
	enc{ST_UINT32, 32}: "ReserveIncrement",
	enc{ST_UINT32, 33}: "SetFlag",
	enc{ST_UINT32, 34}: "ClearFlag",
	// 64-bit unsigned integers (common)
	enc{ST_UINT64, 1}: "IndexNext",
	enc{ST_UINT64, 2}: "IndexPrevious",
	enc{ST_UINT64, 3}: "BookNode",
	enc{ST_UINT64, 4}: "OwnerNode",
	enc{ST_UINT64, 5}: "BaseFee",
	enc{ST_UINT64, 6}: "ExchangeRate",
	enc{ST_UINT64, 7}: "LowNode",
	enc{ST_UINT64, 8}: "HighNode",
	// 128-bit (common)
	enc{ST_HASH128, 1}: "EmailHash",
	// 256-bit (common)
	enc{ST_HASH256, 1}: "LedgerHash",
	enc{ST_HASH256, 2}: "ParentHash",
	enc{ST_HASH256, 3}: "TransactionHash",
	enc{ST_HASH256, 4}: "AccountHash",
	enc{ST_HASH256, 5}: "PreviousTxnID",
	enc{ST_HASH256, 6}: "LedgerIndex",
	enc{ST_HASH256, 7}: "WalletLocator",
	enc{ST_HASH256, 8}: "RootIndex",
	enc{ST_HASH256, 9}: "AccountTxnID",
	// 256-bit (uncommon)
	enc{ST_HASH256, 16}: "BookDirectory",
	enc{ST_HASH256, 17}: "InvoiceID",
	enc{ST_HASH256, 18}: "Nickname",
	enc{ST_HASH256, 19}: "Amendment",
	// currency amount (common)
	enc{ST_AMOUNT, 1}: "Amount",
	enc{ST_AMOUNT, 2}: "Balance",
	enc{ST_AMOUNT, 3}: "LimitAmount",
	enc{ST_AMOUNT, 4}: "TakerPays",
	enc{ST_AMOUNT, 5}: "TakerGets",
	enc{ST_AMOUNT, 6}: "LowLimit",
	enc{ST_AMOUNT, 7}: "HighLimit",
	enc{ST_AMOUNT, 8}: "Fee",
	enc{ST_AMOUNT, 9}: "SendMax",
	// currency amount (uncommon)
	enc{ST_AMOUNT, 16}: "MinimumOffer",
	enc{ST_AMOUNT, 17}: "RippleEscrow",
	enc{ST_AMOUNT, 18}: "DeliveredAmount",
	// variable length (common)
	enc{ST_VL, 1}:  "PublicKey",
	enc{ST_VL, 2}:  "MessageKey",
	enc{ST_VL, 3}:  "SigningPubKey",
	enc{ST_VL, 4}:  "TxnSignature",
	enc{ST_VL, 5}:  "Generator",
	enc{ST_VL, 6}:  "Signature",
	enc{ST_VL, 7}:  "Domain",
	enc{ST_VL, 8}:  "FundCode",
	enc{ST_VL, 9}:  "RemoveCode",
	enc{ST_VL, 10}: "ExpireCode",
	enc{ST_VL, 11}: "CreateCode",
	enc{ST_VL, 12}: "MemoType",
	enc{ST_VL, 13}: "MemoData",
	// account
	enc{ST_ACCOUNT, 1}: "Account",
	enc{ST_ACCOUNT, 2}: "Owner",
	enc{ST_ACCOUNT, 3}: "Destination",
	enc{ST_ACCOUNT, 4}: "Issuer",
	enc{ST_ACCOUNT, 7}: "Target",
	enc{ST_ACCOUNT, 8}: "RegularKey",
	// inner object
	enc{ST_OBJECT, 1}:  "EndOfObject",
	enc{ST_OBJECT, 2}:  "TransactionMetaData",
	enc{ST_OBJECT, 3}:  "CreatedNode",
	enc{ST_OBJECT, 4}:  "DeletedNode",
	enc{ST_OBJECT, 5}:  "ModifiedNode",
	enc{ST_OBJECT, 6}:  "PreviousFields",
	enc{ST_OBJECT, 7}:  "FinalFields",
	enc{ST_OBJECT, 8}:  "NewFields",
	enc{ST_OBJECT, 9}:  "TemplateEntry",
	enc{ST_OBJECT, 10}: "Memo",
	// array of objects
	enc{ST_ARRAY, 1}: "EndOfArray",
	enc{ST_ARRAY, 2}: "SigningAccounts",
	enc{ST_ARRAY, 3}: "TxnSignatures",
	enc{ST_ARRAY, 4}: "Signatures",
	enc{ST_ARRAY, 5}: "Template",
	enc{ST_ARRAY, 6}: "Necessary",
	enc{ST_ARRAY, 7}: "Sufficient",
	enc{ST_ARRAY, 8}: "AffectedNodes",
	enc{ST_ARRAY, 9}: "Memos",
	// 8-bit unsigned integers (common)
	enc{ST_UINT8, 1}: "CloseResolution",
	enc{ST_UINT8, 2}: "TemplateEntryType",
	enc{ST_UINT8, 3}: "TransactionResult",
	// 160-bit (common)
	enc{ST_HASH160, 1}: "TakerPaysCurrency",
	enc{ST_HASH160, 2}: "TakerPaysIssuer",
	enc{ST_HASH160, 3}: "TakerGetsCurrency",
	enc{ST_HASH160, 4}: "TakerGetsIssuer",
	// path set
	enc{ST_PATHSET, 1}: "Paths",
	// vector of 256-bit
	enc{ST_VECTOR256, 1}: "Indexes",
	enc{ST_VECTOR256, 2}: "Hashes",
	enc{ST_VECTOR256, 3}: "Amendments",
}

var reverseEncodings map[string]enc
var signingFields map[enc]struct{}

func init() {
	reverseEncodings = make(map[string]enc)
	signingFields = make(map[enc]struct{})
	for e, name := range encodings {
		reverseEncodings[name] = e
		if strings.Contains(name, "Signature") {
			signingFields[e] = struct{}{}
		}
	}
}

func (e enc) Priority() uint32 {
	return uint32(e.typ)<<16 | uint32(e.field)
}

func (e enc) SigningField() bool {
	_, ok := signingFields[e]
	return ok
}

func (h HashPrefix) String() string {
	return string(h.Bytes())
}

func (h HashPrefix) Bytes() []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(h))
	return b
}

func (n NodeType) String() string {
	return nodeTypes[n]
}
