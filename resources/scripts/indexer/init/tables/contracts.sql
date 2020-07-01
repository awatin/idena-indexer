CREATE TABLE IF NOT EXISTS contracts
(
    tx_id bigint NOT NULL,
    CONSTRAINT contracts_pkey PRIMARY KEY (tx_id),
    CONSTRAINT contracts_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS fact_evidence_contracts
(
    contract_tx_id bigint NOT NULL,
    CONSTRAINT fact_evidence_contracts_pkey PRIMARY KEY (contract_tx_id),
    CONSTRAINT fact_evidence_contracts_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES contracts (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);