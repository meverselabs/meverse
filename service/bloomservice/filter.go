package bloomservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/chain"
	mtypes "github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/ethereum/core/bloombits"
	etypes "github.com/meverselabs/meverse/ethereum/core/types"
	"github.com/meverselabs/meverse/ethereum/params"
	"github.com/meverselabs/meverse/service/pack"
	"github.com/meverselabs/meverse/service/txsearch/itxsearch"
)

// FilterQuery contains options for contract log filtering.
type FilterQuery struct {
	BlockHash *common.Hash     // used by eth_getLogs, return logs only from block with this hash
	FromBlock *big.Int         // beginning of the queried range, nil means genesis block
	ToBlock   *big.Int         // end of the range, nil means latest block
	Addresses []common.Address // restricts matches to events created by specific contracts

	// The Topic list restricts matches to particular event topics. Each event has a list
	// of topics. Topics matches a prefix of that list. An empty element slice matches any
	// topic. Non-empty elements represent an alternative that matches any of the
	// contained topics.
	//
	// Examples:
	// [] or nil          matches any topic list
	// [A,B]              matches topic (A OR B) in first position
	// [[A]]              matches topic A in first position
	// [[], [B]]          matches any topic in first position AND B in second position
	// [[A], [B]]         matches topic A in first position AND B in second position
	// [[A, B], [C, D]]   matches topic (A OR B) in first position AND (C OR D) in second position
	Topics [][]common.Hash
}

// ToFilter converts map[string]interface{} to FilterQuery struct
func ToFilter(arg map[string]interface{}) FilterQuery {

	q := FilterQuery{}

	if v, ok := arg["address"]; ok {
		if list, isSlice := v.([]string); isSlice {
			q.Addresses = make([]common.Address, len(list))
			for _, address := range list {
				q.Addresses = append(q.Addresses, common.HexToAddress(address))
			}
		} else if item, isString := v.(string); isString {
			q.Addresses = []common.Address{common.HexToAddress(item)}
		}
	}

	if topicArg, ok1 := arg["topics"]; ok1 {
		if outList, ok2 := topicArg.([]interface{}); ok2 {
			nested := true
			if len(outList) > 0 {
				kind := reflect.ValueOf(outList[0]).Kind()
				if kind == reflect.String {
					nested = false
				}
			}

			if nested {
				for _, innerList := range outList {
					var topics []common.Hash
					if iList, ok3 := innerList.([]interface{}); ok3 {
						for _, topic := range iList {
							if s, ok := topic.(string); ok {
								topics = append(topics, hash.HexToHash(s))
							}
						}
					}
					q.Topics = append(q.Topics, topics)
				}
			} else {
				var topics []common.Hash
				for _, topic := range outList {
					if s, ok := topic.(string); ok {
						topics = append(topics, hash.HexToHash(s))
					}
				}
				q.Topics = append(q.Topics, topics)
			}
		}
	}

	if v, ok := arg["blockHash"]; ok {
		if h, isString := v.(string); isString {
			hh := hash.HexToHash(h)
			q.BlockHash = &hh
		}
	}

	if q.BlockHash == nil {
		if v, ok := arg["fromBlock"]; ok {
			if s, ok2 := v.(string); ok2 {
				if s == "latest" {
					q.FromBlock = big.NewInt(-1)
				} else {
					q.FromBlock = new(big.Int)
					if strings.HasPrefix(s, "0x") {
						s = s[2:]
						if q.FromBlock, ok = q.FromBlock.SetString(s, 16); !ok {
							q.FromBlock = big.NewInt(-1)
						}
					} else {
						if q.FromBlock, ok = q.FromBlock.SetString(s, 10); !ok {
							q.FromBlock = big.NewInt(-1)
						}
					}
				}
			} else if i, ok2 := v.(int); ok2 {
				if i < 0 {
					q.FromBlock = big.NewInt(0)
				} else {
					q.FromBlock = big.NewInt(int64(i))
				}
			} else if jN, ok2 := v.(json.Number); ok2 {
				if fromBlock, err := jN.Int64(); err == nil {
					q.FromBlock = big.NewInt(fromBlock)
				} else {
					q.FromBlock = big.NewInt(-1)
				}
			}
		} else {
			q.FromBlock = big.NewInt(-1)
		}

		if v, ok := arg["toBlock"]; ok {
			if s, ok2 := v.(string); ok2 {
				if s == "latest" {
					q.ToBlock = big.NewInt(-1)
				} else {
					q.ToBlock = new(big.Int)
					if strings.HasPrefix(s, "0x") {
						s = s[2:]
						if _, ok := q.ToBlock.SetString(s, 16); !ok {
							q.ToBlock = big.NewInt(-1)
						}
					} else {
						if q.ToBlock, ok = q.ToBlock.SetString(s, 10); !ok {
							q.ToBlock = big.NewInt(-1)
						}
					}
				}
			} else if i, ok2 := v.(int); ok2 {
				if i < 0 {
					q.ToBlock = big.NewInt(-1)
				} else {
					q.ToBlock = big.NewInt(int64(i))
				}
			} else if jN, ok2 := v.(json.Number); ok2 {
				if toBlock, err := jN.Int64(); err == nil {
					q.ToBlock = big.NewInt(toBlock)
				} else {
					q.ToBlock = big.NewInt(-1)
				}
			}
		} else {
			q.ToBlock = big.NewInt(-1)
		}
	}
	return q
}

// https://eth.wiki/json-rpc/API#eth_getlogs
func FilterLogs(cn *chain.Chain, ts itxsearch.ITxSearch, bs *BloomBitService, crit FilterQuery) ([]*types.Log, error) {
	var filter *Filter
	if crit.BlockHash != nil {
		// Block filter requested, construct a single-shot filter
		filter = NewBlockFilter(cn, ts, *crit.BlockHash, crit.Addresses, crit.Topics)
	} else {
		// Convert the RPC block numbers into internal representations
		begin := rpc.LatestBlockNumber.Int64()
		if crit.FromBlock != nil {
			begin = crit.FromBlock.Int64()
		}
		end := rpc.LatestBlockNumber.Int64()
		if crit.ToBlock != nil {
			end = crit.ToBlock.Int64()
		}
		// Construct the range filter
		filter = NewRangeFilter(cn, ts, bs, begin, end, crit.Addresses, crit.Topics)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*params.QueryTimeout)
	defer cancel()

	logs, err := filter.Logs(ctx)

	if err != nil {
		return nil, err
	}
	return returnLogs(logs), err
}

// Filter can be used to retrieve and filter logs.
type Filter struct {
	backend *chain.Chain
	ts      itxsearch.ITxSearch
	bs      *BloomBitService

	//db        ethdb.Database
	addresses []common.Address
	topics    [][]common.Hash

	block      common.Hash // Block hash if filtering a single block
	begin, end int64       // Range interval if filtering multiple blocks

	matcher *bloombits.Matcher
}

// NewBlockFilter creates a new filter which directly inspects the contents of
// a block to figure out whether it is interesting or not.
func NewBlockFilter(backend *chain.Chain, ts itxsearch.ITxSearch, block common.Hash, addresses []common.Address, topics [][]common.Hash) *Filter {
	// Create a generic filter and convert it into a block filter
	filter := newFilter(backend, ts, addresses, topics)
	filter.block = block
	return filter
}

// NewRangeFilter creates a new filter which uses a bloom filter on blocks to
// figure out whether a particular block is interesting or not.
func NewRangeFilter(backend *chain.Chain, ts itxsearch.ITxSearch, bs *BloomBitService, begin, end int64, addresses []common.Address, topics [][]common.Hash) *Filter {
	// Flatten the address and topic filter clauses into a single bloombits filter
	// system. Since the bloombits are not positional, nil topics are permitted,
	// which get flattened into a nil byte slice.
	var filters [][][]byte
	if len(addresses) > 0 {
		filter := make([][]byte, len(addresses))
		for i, address := range addresses {
			filter[i] = address.Bytes()
		}
		filters = append(filters, filter)
	}
	for _, topicList := range topics {
		filter := make([][]byte, len(topicList))
		for i, topic := range topicList {
			filter[i] = topic.Bytes()
		}
		filters = append(filters, filter)
	}
	size, _ := bs.BloomStatus()

	// Create a generic filter and convert it into a range filter
	filter := newFilter(backend, ts, addresses, topics)

	filter.matcher = bloombits.NewMatcher(size, filters)
	filter.bs = bs
	filter.begin = begin
	filter.end = end

	return filter
}

// newFilter creates a generic filter that can either filter based on a block hash,
// or based on range queries. The search criteria needs to be explicitly set.
func newFilter(backend *chain.Chain, ts itxsearch.ITxSearch, addresses []common.Address, topics [][]common.Hash) *Filter {
	return &Filter{
		backend:   backend,
		ts:        ts,
		addresses: addresses,
		topics:    topics,
		//db:        backend.ChainDb(),
	}
}

// returnLogs is a helper that will return an empty log array in case the given logs array is nil,
// otherwise the given logs array is returned.
func returnLogs(logs []*types.Log) []*types.Log {
	if logs == nil {
		return []*types.Log{}
	}
	return logs
}

// Logs searches the blockchain for matching log entries, returning all from the
// first block that contains matches, updating the start of the filter accordingly.
func (f *Filter) Logs(ctx context.Context) ([]*types.Log, error) {
	provider := f.backend.Provider()

	// If we're doing singleton block filtering, execute and return
	if f.block != (common.Hash{}) {
		hei, err := f.ts.BlockHeight(f.block)
		if err != nil {
			return nil, err
		}
		block, err := provider.Block(hei)
		if err != nil {
			return nil, err
		}
		return f.blockLogs(ctx, block)
	}
	// Figure out the limits of the filter range
	head := int64(provider.Height())
	initHeight := int64(provider.InitHeight())
	if f.begin == -1 {
		f.begin = head
	} else if f.begin <= initHeight {
		f.begin = initHeight + 1
	}
	end := uint64(f.end)
	if f.end == -1 {
		end = uint64(head)
	}
	// Gather all indexed logs, and finish with non indexed ones
	var (
		logs []*types.Log
		err  error
	)
	size, sections := f.bs.BloomStatus()
	if indexed := sections * size; indexed > uint64(f.begin) {
		if indexed > end {
			logs, err = f.indexedLogs(ctx, end)
		} else {
			logs, err = f.indexedLogs(ctx, indexed-1)
		}
		if err != nil {
			return logs, err
		}
	}
	//log.Println("indexedLogs", logs)
	rest, err := f.unindexedLogs(ctx, end)
	logs = append(logs, rest...)
	return logs, err
}

// blockLogs returns the logs matching the filter criteria within a single block.
func (f *Filter) blockLogs(ctx context.Context, block *mtypes.Block) (logs []*types.Log, err error) {
	bloom, err := BlockLogsBloom(f.backend, block)
	if err != nil {
		return nil, err
	}

	if bloomFilter(bloom, f.addresses, f.topics) {
		found, err := f.checkMatches(ctx, block)
		if err != nil {
			return logs, err
		}
		logs = append(logs, found...)
	}
	return logs, nil
}

// checkMatches checks if the receipts belonging to the given header contain any log events that
// match the filter criteria. This function is called when the bloom filter signals a potential match.
func (f *Filter) checkMatches(ctx context.Context, block *mtypes.Block) (logs []*types.Log, err error) {
	// Get the logs of the block
	logsList, err := blockShortLogs(f.backend, block)
	if err != nil {
		return nil, err
	}
	var unfiltered []*types.Log
	for _, logs := range logsList {
		unfiltered = append(unfiltered, logs...)
	}
	logs = filterLogs(unfiltered, nil, nil, f.addresses, f.topics)
	if len(logs) > 0 {
		// We have matching logs, check if we need to resolve full logs via the light client
		if logs[0].TxHash == (common.Hash{}) {
			logsList, err := blockFullLogs(f.backend, block)
			if err != nil {
				return nil, err
			}
			unfiltered = unfiltered[:0]
			for _, logs := range logsList {
				unfiltered = append(unfiltered, logs...)
			}
			logs = filterLogs(unfiltered, nil, nil, f.addresses, f.topics)
		}
		return logs, nil
	}
	return nil, nil
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []*types.Log, fromBlock, toBlock *big.Int, addresses []common.Address, topics [][]common.Hash) []*types.Log {
	var ret []*types.Log
Logs:
	for _, log := range logs {
		if fromBlock != nil && fromBlock.Int64() >= 0 && fromBlock.Uint64() > log.BlockNumber {
			continue
		}
		if toBlock != nil && toBlock.Int64() >= 0 && toBlock.Uint64() < log.BlockNumber {
			continue
		}

		if len(addresses) > 0 && !includes(addresses, log.Address) {
			continue
		}
		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(topics) > len(log.Topics) {
			continue Logs
		}
		for i, sub := range topics {
			match := len(sub) == 0 // empty rule set == wildcard
			for _, topic := range sub {
				if log.Topics[i] == topic {
					match = true
					break
				}
			}
			if !match {
				continue Logs
			}
		}
		ret = append(ret, log)
	}
	return ret
}

func includes(addresses []common.Address, a common.Address) bool {
	for _, addr := range addresses {
		if addr == a {
			return true
		}
	}

	return false
}

// blockShortLogs retrieves the logs generated by the transactions included in a given block
func blockShortLogs(chain *chain.Chain, block *mtypes.Block) ([][]*types.Log, error) {

	provider := chain.Provider()
	header := &block.Header

	receipts, err := provider.Receipts(header.Height)
	if err != nil {
		return nil, err
	}
	var logList [][]*types.Log
	for i := 0; i < len(receipts); i++ {
		tx := block.Body.Transactions[i]
		if tx.VmType == mtypes.Evm {
			logList = append(logList, receipts[i].Logs)
		} else {
			logs := []*types.Log{}
			evs, err := FindTransactionsEvents(block.Body.Transactions, block.Body.Events, i)
			if err != nil {
				return nil, err
			}
			for j := 0; j < len(evs); j++ {
				mc := &mtypes.MethodCallEvent{}
				_, err := mc.ReadFrom(bytes.NewReader(evs[j].Result))
				if err != nil {
					return nil, err
				}

				if mc.To == (common.Address{}) {
					return nil, errors.New("event To is null")
				}

				// provider.Receipts(header.Height) 에서 가져오는 것만 처리
				log := types.Log{}
				log.Address = mc.To
				topics, err := makeEventTopics(chain.NewContext(), mc)
				if err != nil {
					return nil, err
				}
				log.Topics = hashTopics(topics)
				log.Removed = false

				logs = append(logs, &log)
			}
			logList = append(logList, logs)
		}
	}

	return logList, nil
}

// blockFullLogs retrieves the logs generated by the transactions included in a given block
func blockFullLogs(chain *chain.Chain, block *mtypes.Block) ([][]*types.Log, error) {

	provider := chain.Provider()
	header := &block.Header

	receipts, err := provider.Receipts(header.Height)
	if err != nil {
		return nil, err
	}
	signer := etypes.MakeSigner(provider.ChainID(), header.Height)
	var logList [][]*types.Log
	for i := 0; i < len(receipts); i++ {
		tx := block.Body.Transactions[i]
		if tx.VmType == mtypes.Evm {
			etx := new(types.Transaction)
			if err := etx.UnmarshalBinary(tx.Args); err != nil {
				return nil, err
			}

			err := receipts.DeriveReceiptFields(bin.MustWriterToHash(header), uint64(header.Height), uint16(i), etx, signer)
			if err != nil {
				return nil, err
			}
			logList = append(logList, receipts[i].Logs)
		} else {
			logs, err := nonEvmTxLogs(chain, block, i)
			if err != nil {
				return nil, err
			}
			logList = append(logList, logs)
		}
	}

	return logList, nil
}

// nonEvmTxLogs returns logs converted from non-evm type transaction events
func nonEvmTxLogs(chain *chain.Chain, block *mtypes.Block, idx int) ([]*types.Log, error) {

	tx := block.Body.Transactions[idx]
	if tx.VmType == mtypes.Evm {
		return nil, nil
	}

	header := &block.Header
	evs, err := FindTransactionsEvents(block.Body.Transactions, block.Body.Events, idx)
	if err != nil {
		return nil, err
	}
	logs, err := EventsToLogs(chain, header, tx, evs, idx)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// EventsToLogs converts non-evm type events to ethereum type logs
func EventsToLogs(chain *chain.Chain, header *mtypes.Header, tx *mtypes.Transaction, evs []*mtypes.Event, idx int) ([]*types.Log, error) {
	logs := []*types.Log{}
	for j := 0; j < len(evs); j++ {
		mc := &mtypes.MethodCallEvent{}
		if _, err := mc.ReadFrom(bytes.NewReader(evs[j].Result)); err != nil {
			return nil, err
		}

		if mc.To == (common.Address{}) {
			return nil, errors.New("event To is null")
		}

		log := types.Log{}
		log.Address = mc.To

		topics, err := makeEventTopics(chain.NewContext(), mc)
		if err != nil {
			return nil, err
		}

		log.Topics = hashTopics(topics)
		log.Data, err = pack.Pack(mc.Args)
		if err != nil {
			return nil, err
		}
		log.BlockNumber = uint64(header.Height)
		log.TxHash = tx.Hash(header.Height)
		log.TxIndex = uint(idx)
		log.BlockHash = bin.MustWriterToHash(header)
		log.Index = uint(j)
		log.Removed = false

		logs = append(logs, &log)
	}

	return logs, nil
}

/*
// Pack packs MethodCallEvent arguments. When unpacking, the function(event) definition is necessary
func pack(args []interface{}) ([]byte, error) {

	var ret []byte
	for _, arg := range args {
		argByte, err := packElement(reflect.ValueOf(arg))
		if err != nil {
			return nil, err
		}
		ret = append(ret, argByte...)
	}
	return ret, nil
}

// packList packs only slice and array, excludes address, byte-slice([]byte), byte-array([k]byte)
func packList(v reflect.Value) ([]byte, error) {

	var ret []byte
	ret = append(ret, packNum(reflect.ValueOf(v.Len()))...)

	// slice, and length = 0
	if v.Len() == 0 {
		return ret, nil
	}

	// calculate offset if any
	offset := 0
	offsetReq := isDynamicType(v.Index(0))
	if offsetReq {
		offset = getTypeSize(v.Index(0)) * v.Len()
	}
	var tail []byte
	for i := 0; i < v.Len(); i++ {
		val, err := packElement(v.Index(i))
		if err != nil {
			return nil, err
		}
		if !offsetReq {
			ret = append(ret, val...)
			continue
		}
		ret = append(ret, packNum(reflect.ValueOf(offset))...)
		offset += len(val)
		tail = append(tail, val...)
	}
	return append(ret, tail...), nil
}

// isDynamicType returns true if the type is dynamic.
// The following types are called “dynamic”:
// * bytes
// * string
// * T[] for any T
// * T[k] for any dynamic T and any k >= 0
// * (T1,...,Tk) if Ti is dynamic for some 1 <= i <= k
// reflect.TypeOf(v).Elem().Kind() 의 경우 하위로 계속 내려갈 수
func isDynamicType(v reflect.Value) bool {
	switch kind := v.Kind(); kind {
	case reflect.String:
		return true
	case reflect.Slice:
		return true
	case reflect.Array:
		if v.Len() > 0 && isDynamicType(v.Index(0)) {
			return true
		}
		return false
	default:
		return false
	}
}

// getTypeSize returns the size that this type needs to occupy.
// We distinguish static and dynamic types. Static types are encoded in-place
// and dynamic types are encoded at a separately allocated location after the
// current block.
// So for a static variable, the size returned represents the size that the
// variable actually occupies.
// For a dynamic variable, the returned size is fixed 32 bytes, which is used
// to store the location reference for actual value storage.
func getTypeSize(v reflect.Value) int {
	kind := v.Kind()
	if kind == reflect.Array {
		if v.Len() > 0 && !isDynamicType(v.Index(0)) {
			if v.Index(0).Kind() == reflect.Array {
				return v.Len() * getTypeSize(v.Index(0))
			}
			return v.Len() * 32
		}
	}
	return 32
}

// packElement packs the given reflect value according to the abi specification in v.
func packElement(v reflect.Value) ([]byte, error) {
	switch kind := v.Kind(); kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return packNum(v), nil
	case reflect.Ptr:
		switch typ := v.Type(); typ {
		case bigIntType:
			return math.U256Bytes(new(big.Int).Set(v.Interface().(*big.Int))), nil
		case amountType:
			return math.U256Bytes(new(big.Int).Set(v.Interface().(*amount.Amount).Int)), nil
		default:
			return nil, fmt.Errorf("packElement, unknown type: %v", v.Type())
		}
	case reflect.String:
		return packBytesSlice([]byte(v.String()), v.Len()), nil
	case reflect.Array:
		switch typ := v.Type(); typ {
		case addressType:
			v = mustArrayToByteSlice(v)
			return common.LeftPadBytes(v.Bytes(), 32), nil
		default:
			// ex. [220]byte
			elem := v.Index(0)
			if elem.Kind() == reflect.Uint8 {
				v = mustArrayToByteSlice(v)
				return common.RightPadBytes(v.Bytes(), 32), nil
			}
			return packList(v)
		}
	case reflect.Slice:
		switch typ := v.Type(); typ {
		case byteSliceType:
			v = mustArrayToByteSlice(v)
			return packBytesSlice(v.Bytes(), v.Len()), nil
		}
		return packList(v)
	case reflect.Bool:
		if v.Bool() {
			return math.PaddedBigBytes(common.Big1, 32), nil
		}
		return math.PaddedBigBytes(common.Big0, 32), nil
	case reflect.Interface:
		ret, err := packElement(v.Elem())
		if err != nil {
			return nil, err
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("packElement, unknown type: %v", v.Type())
	}
}

// packBytesSlice packs the given bytes as [L, V] as the canonical representation
// bytes slice.
func packBytesSlice(bytes []byte, l int) []byte {
	len := packNum(reflect.ValueOf(l))
	return append(len, common.RightPadBytes(bytes, (l+31)/32*32)...)
}

// packNum packs the given number (using the reflect value) and will cast it to appropriate number representation.
func packNum(value reflect.Value) []byte {
	switch kind := value.Kind(); kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return math.U256Bytes(new(big.Int).SetUint64(value.Uint()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return math.U256Bytes(big.NewInt(value.Int()))
	default:
		panic("abi: fatal error")
	}
}

*/

// indexedLogs returns the logs matching the filter criteria based on the bloom
// bits indexed available locally or via the network.
func (f *Filter) indexedLogs(ctx context.Context, end uint64) ([]*types.Log, error) {
	// Create a matcher session and request servicing from the backend
	matches := make(chan uint64, 64)

	session, err := f.matcher.Start(ctx, uint64(f.begin), end, matches)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	f.bs.ServiceFilter(ctx, session)

	// Iterate over the matches until exhausted or context closed
	var logs []*types.Log

	for {
		select {
		case number, ok := <-matches:
			// Abort if all matches have been fulfilled
			if !ok {
				err := session.Error()
				if err == nil {
					f.begin = int64(end) + 1
				}
				return logs, err
			}
			f.begin = int64(number) + 1

			block, err := f.backend.Provider().Block(uint32(number))
			if block == nil || err != nil {
				return logs, err
			}

			found, err := f.checkMatches(ctx, block)
			if err != nil {
				return logs, err
			}
			logs = append(logs, found...)

		case <-ctx.Done():
			return logs, ctx.Err()
		}
	}
}

// unindexedLogs returns the logs matching the filter criteria based on raw block
// iteration and bloom matching.
func (f *Filter) unindexedLogs(ctx context.Context, end uint64) ([]*types.Log, error) {
	var logs []*types.Log

	for ; f.begin <= int64(end); f.begin++ {
		if f.begin == 0 {
			continue
		}
		block, err := f.backend.Provider().Block(uint32(f.begin))
		if block == nil || err != nil {
			return logs, err
		}
		found, err := f.blockLogs(ctx, block)
		if err != nil {
			return logs, err
		}
		logs = append(logs, found...)
	}
	return logs, nil
}

// bloomFilter checks whether addresses or topics are included in the given bloom
func bloomFilter(bloom types.Bloom, addresses []common.Address, topics [][]common.Hash) bool {
	if len(addresses) > 0 {
		var included bool
		for _, addr := range addresses {
			if types.BloomLookup(bloom, addr) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, sub := range topics {
		included := len(sub) == 0 // empty rule set == wildcard
		for _, topic := range sub {
			if types.BloomLookup(bloom, topic) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}
	return true
}
