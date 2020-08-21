CREATE OR REPLACE PROCEDURE save_fact_evidence_contracts(p_block_height bigint,
                                                         p_items tp_fact_evidence_contract[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    CONTRACT_TYPE_FACT_EVIDENCE       CONSTANT smallint = 2;
    FACT_EVIDENCE_STATE_PENDING       CONSTANT smallint = 0;
    FACT_EVIDENCE_STATE_REASON_DEPLOY CONSTANT smallint = 1;
    l_item                                     tp_fact_evidence_contract;
    l_contract_address_id                      bigint;
    l_tx_id                                    bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];
            INSERT INTO addresses (address, block_height)
            VALUES (l_item.contract_address, p_block_height)
            RETURNING id INTO l_contract_address_id;

            SELECT id INTO l_tx_id FROM transactions WHERE lower(hash) = lower(l_item.tx_hash);

            INSERT INTO contracts (tx_id, type) VALUES (l_tx_id, CONTRACT_TYPE_FACT_EVIDENCE);

            INSERT INTO fact_evidence_contracts (contract_tx_id, contract_address_id, balance, state, start_time)
            VALUES (l_tx_id, l_contract_address_id, 0, FACT_EVIDENCE_STATE_PENDING, l_item.start_time);

            INSERT INTO fact_evidence_contract_states (state_tx_id, fe_contract_tx_id, reason, state)
            VALUES (l_tx_id, l_tx_id, FACT_EVIDENCE_STATE_REASON_DEPLOY, FACT_EVIDENCE_STATE_PENDING);
        end loop;
END
$BODY$;

CREATE OR REPLACE PROCEDURE save_fact_evidence_contract_call_starts(p_block_height bigint,
                                                                    p_items tp_fact_evidence_contract_call_start[])
    LANGUAGE 'plpgsql'
AS
$BODY$
DECLARE
    FACT_EVIDENCE_STATE_REASON_CALL_START CONSTANT smallint = 2;
    FACT_EVIDENCE_STATE_RUNNING           CONSTANT smallint = 1;
    l_item                                         tp_fact_evidence_contract_call_start;
    l_contract_tx_id                               bigint;
    l_tx_id                                        bigint;
BEGIN
    for i in 1..cardinality(p_items)
        loop
            l_item = p_items[i];

            SELECT id INTO l_tx_id FROM transactions WHERE lower(hash) = lower(l_item.tx_hash);

            SELECT contract_tx_id
            INTO l_contract_tx_id
            FROM fact_evidence_contracts
            WHERE contract_address_id =
                  (SELECT id FROM addresses WHERE lower(address) = lower(l_item.contract_address));

            INSERT INTO fact_evidence_contract_call_starts (call_tx_id, fe_contract_tx_id, voting_min_payment)
            VALUES (l_tx_id, l_contract_tx_id, null_if_zero(l_item.voting_min_payment));

            INSERT INTO fact_evidence_contract_states (state_tx_id, fe_contract_tx_id, reason, state)
            VALUES (l_tx_id, l_contract_tx_id, FACT_EVIDENCE_STATE_REASON_CALL_START, FACT_EVIDENCE_STATE_RUNNING);

            UPDATE fact_evidence_contracts
            SET state              = FACT_EVIDENCE_STATE_RUNNING,
                start_block_height = l_item.start_block_height
            WHERE contract_tx_id = l_contract_tx_id;
        end loop;
END
$BODY$;