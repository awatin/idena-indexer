package db

import "database/sql"

type context struct {
	epochId            int64
	identityIdsPerAddr map[string]int64
	flipIdsPerCid      map[string]int64
	txIdsPerHash       map[string]int64
	a                  *postgresAccessor
}

func newContext(a *postgresAccessor) *context {
	return &context{a: a}
}

func (c *context) identityId(tx *sql.Tx, addr string) (int64, error) {
	if c.identityIdsPerAddr == nil {
		c.identityIdsPerAddr = make(map[string]int64)
	}
	if id, present := c.identityIdsPerAddr[addr]; present {
		return id, nil
	}
	id, err := c.a.getIdentityId(tx, addr)
	if err != nil {
		return 0, err
	}
	c.identityIdsPerAddr[addr] = id
	return id, nil
}

func (c *context) flipId(tx *sql.Tx, cid string) (int64, error) {
	if c.flipIdsPerCid == nil {
		c.flipIdsPerCid = make(map[string]int64)
	}
	if id, present := c.flipIdsPerCid[cid]; present {
		return id, nil
	}
	id, err := c.a.getFlipId(tx, cid)
	if err != nil {
		return 0, err
	}
	c.flipIdsPerCid[cid] = id
	return id, nil
}

func (c *context) txId(hash string) int64 {
	return c.txIdsPerHash[hash]
}
