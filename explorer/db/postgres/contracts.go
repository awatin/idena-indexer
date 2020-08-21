package postgres

import (
	"github.com/idena-network/idena-indexer/explorer/types"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
)

const (
	contractsQuery = "contracts.sql"
)

func (a *postgresAccessor) Contracts(status string, count uint64, continuationToken *string) ([]types.Contract, *string, error) {
	parseToken := func(continuationToken *string) (contractId *uint64, balance *decimal.Decimal, err error) {
		if continuationToken == nil {
			return
		}
		strs := strings.Split(*continuationToken, "-")
		if len(strs) != 2 {
			err = errors.New("invalid continuation token")
			return
		}
		sContractId := strs[0]
		if contractId, err = parseUintContinuationToken(&sContractId); err != nil {
			return
		}
		var d decimal.Decimal
		d, err = decimal.NewFromString(strs[1])
		if err != nil {
			return
		}
		balance = &d
		return
	}
	contractId, balance, err := parseToken(continuationToken)
	if err != nil {
		return nil, nil, err
	}
	rows, err := a.db.Query(a.getQuery(contractsQuery), status, count+1, contractId, balance)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var res []types.Contract
	var nextBalance decimal.Decimal
	var nextContractId uint64
	for rows.Next() {
		item := types.Contract{}
		err = rows.Scan(
			&nextContractId,
			&item.ContractAddress,
			&item.Balance,
		)
		if err != nil {
			return nil, nil, err
		}
		nextBalance = item.Balance
		res = append(res, item)
	}
	var nextContinuationToken *string
	if len(res) > 0 && len(res) == int(count)+1 {
		t := strconv.FormatUint(nextContractId, 10) + "-" + nextBalance.String()
		nextContinuationToken = &t
		res = res[:len(res)-1]
	}
	return res, nextContinuationToken, nil
}
