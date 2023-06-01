package utils

import (
	"encoding/hex"
	mathbig "math/big"
	"strconv"
	"strings"

	"github.com/filecoin-project/go-state-types/big"

	addr "github.com/filecoin-project/go-address"
)

// ParseHexToUint64 parse start with hex str to uint64
func ParseHexToUint64(str string) (uint64, error) {
	parsedInt, err := strconv.ParseUint(strings.Replace(str, "0x", "", -1), 16, 64)
	if err != nil {
		return 0, err
	}
	return parsedInt, nil
}

// ParseHexToInt64 parse start with hex str to int64
func ParseHexToInt64(str string) (int64, error) {
	parsedInt, err := strconv.ParseInt(strings.Replace(str, "0x", "", -1), 16, 64)
	if err != nil {
		return 0, err
	}
	return parsedInt, nil
}

// ParseHexToBigInt parse hex to big int
func ParseHexToBigInt(str string) big.Int {
	replaced := strings.Replace(str, "0x", "", -1)
	if len(replaced)%2 == 1 {
		replaced = "0" + replaced
	}

	i := new(mathbig.Int)
	i.SetString(replaced, 16)
	return big.NewFromGo(i)
}

// ParseStrToHex parse str to hex
func ParseStrToHex(str string) (string, error) {
	str = strings.Replace(str, "0x", "", -1)
	if len(str)%2 == 1 {
		str = "0" + str
	}
	decoded, err := hex.DecodeString(str)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(decoded), nil
}

// https://github.com/filecoin-project/go-state-types/blob/master/builtin/singletons.go#L22
func MustMakeAddress(id uint64) addr.Address {
	address, err := addr.NewIDAddress(id)
	if err != nil {
		panic(err)
	}
	return address
}
