package viewchain

import (
	"bytes"
	"runtime"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

// var bigIntType = reflect.TypeOf(&big.Int{})
// var amountType = reflect.TypeOf(&amount.Amount{})

type ViewCaller struct {
	cn *chain.Chain
}

func NewViewCaller(cn *chain.Chain) *ViewCaller {
	return &ViewCaller{
		cn: cn,
	}
}

func (m *ViewCaller) _exec(conAddr common.Address, from, method string, inputs []interface{}) (types.IInteractor, error) {
	types.ExecLock.Lock()
	defer types.ExecLock.Unlock()

	ctx := m.cn.NewContext()
	cont, err := ctx.Contract(conAddr)
	if err != nil {
		return nil, err
	}
	var cc *types.ContractContext
	if from == "" {
		cc = ctx.ContractContext(cont, cont.Address())
	} else {
		cc = ctx.ContractContext(cont, common.HexToAddress(from))
	}
	intr := types.NewInteractor(ctx, cont, cc, "000000000000", true)
	cc.Exec = intr.Exec
	_, err = intr.Exec(cc, conAddr, method, inputs)
	if err != nil {
		return nil, err
	}

	return intr, nil
}

func (m *ViewCaller) Execute(contAddr common.Address, from, method string, inputs []interface{}) ([]interface{}, uint64, error) {
	intr, err := m._exec(contAddr, from, method, inputs)
	if err != nil {
		return nil, 0, err
	}
	en := intr.EventList()
	gh := intr.GasHistory()
	if len(en) == 0 {
		return nil, 0, errors.New("no result")
	} else {
		e := en[0]

		bf := bytes.NewBuffer(e.Result)
		mc := &types.MethodCallEvent{}
		_, err := mc.ReadFrom(bf)
		return mc.Result, gh[0], err
	}
}

func (m *ViewCaller) MultiExecute(addr []common.Address, from string, methods []string, inputss [][]interface{}) ([][]interface{}, error) {
	if len(addr) != len(inputss) {
		return nil, errors.New("not match params count")
	}
	if len(methods) != len(inputss) {
		return nil, errors.New("not match params count")
	}

	result := make([][]interface{}, len(methods))
	for i, method := range methods {
		r, _, err := m.Execute(addr[i], from, method, inputss[i])
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}

func GetVersion() string {
	sb := strings.Builder{}
	sb.WriteString("MEVerse/")
	sb.WriteString(prefix.ClientVersion)
	sb.WriteString("/")
	sb.WriteString(runtime.GOOS)
	sb.WriteString("-")
	sb.WriteString(runtime.GOARCH)
	sb.WriteString("/")
	sb.WriteString(runtime.Version())
	return sb.String()
}
