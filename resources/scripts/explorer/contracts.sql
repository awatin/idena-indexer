SELECT fec.contract_tx_id, contra.address, fec.balance
FROM fact_evidence_contracts fec
         JOIN addresses contra on contra.id = fec.contract_address_id
WHERE $1 = $1 -- todo
  AND ($3::bigint IS NULL
    OR fec.balance <= $4 AND fec.contract_tx_id >= $3)
ORDER BY fec.balance DESC, fec.contract_tx_id
LIMIT $2