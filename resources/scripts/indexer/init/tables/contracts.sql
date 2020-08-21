CREATE TABLE IF NOT EXISTS dic_contract_types
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_contract_types_pkey PRIMARY KEY (id),
    CONSTRAINT dic_contract_types_name_key UNIQUE (name)
);
INSERT INTO dic_contract_types
VALUES (1, 'TimeLock')
ON CONFLICT DO NOTHING;
INSERT INTO dic_contract_types
VALUES (2, 'FactEvidence')
ON CONFLICT DO NOTHING;
INSERT INTO dic_contract_types
VALUES (3, 'EvidenceLock')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dic_fact_evidence_contract_states
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_fact_evidence_contract_states_pkey PRIMARY KEY (id),
    CONSTRAINT dic_fact_evidence_contract_states_name_key UNIQUE (name)
);
INSERT INTO dic_fact_evidence_contract_states
VALUES (0, 'Pending')
ON CONFLICT DO NOTHING;
INSERT INTO dic_fact_evidence_contract_states
VALUES (1, 'Running')
ON CONFLICT DO NOTHING;
INSERT INTO dic_fact_evidence_contract_states
VALUES (3, 'Completed')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS dic_fact_evidence_contract_state_reasons
(
    id   smallint                                           NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT dic_fact_evidence_contract_state_reasons_pkey PRIMARY KEY (id),
    CONSTRAINT dic_fact_evidence_contract_state_reasons_name_key UNIQUE (name)
);
INSERT INTO dic_fact_evidence_contract_state_reasons
VALUES (1, 'Deploy')
ON CONFLICT DO NOTHING;
INSERT INTO dic_fact_evidence_contract_state_reasons
VALUES (2, 'CallStartVoting')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS contracts
(
    tx_id bigint   NOT NULL,
    type  smallint NOT NULL,
    CONSTRAINT contracts_pkey PRIMARY KEY (tx_id),
    CONSTRAINT contracts_tx_id_fkey FOREIGN KEY (tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT contracts_type_fkey FOREIGN KEY (type)
        REFERENCES dic_contract_types (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS fact_evidence_contracts
(
    contract_tx_id      bigint          NOT NULL,
    contract_address_id bigint          NOT NULL,
    balance             numeric(30, 18) NOT NULL,
    state               smallint        NOT NULL,
    start_block_height  bigint,
    start_time          bigint          NOT NULL,
    CONSTRAINT fact_evidence_contracts_pkey PRIMARY KEY (contract_tx_id),
    CONSTRAINT fact_evidence_contracts_contract_tx_id_fkey FOREIGN KEY (contract_tx_id)
        REFERENCES contracts (tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fact_evidence_contracts_contract_address_id_fkey FOREIGN KEY (contract_address_id)
        REFERENCES addresses (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fact_evidence_contracts_state_fkey FOREIGN KEY (state)
        REFERENCES dic_fact_evidence_contract_states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS fact_evidence_contracts_contract_address_id_unique_idx on fact_evidence_contracts (contract_address_id);

CREATE TABLE IF NOT EXISTS fact_evidence_contract_call_starts
(
    call_tx_id         bigint NOT NULL,
    fe_contract_tx_id  bigint NOT NULL,
    voting_min_payment numeric(30, 18),
    CONSTRAINT fact_evidence_contract_call_starts_pkey PRIMARY KEY (call_tx_id),
    CONSTRAINT fact_evidence_contract_call_starts_call_tx_fkey FOREIGN KEY (call_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fact_evidence_contract_call_starts_contract_tx_id_fkey FOREIGN KEY (fe_contract_tx_id)
        REFERENCES fact_evidence_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE IF NOT EXISTS fact_evidence_contract_states
(
    state_tx_id       bigint   NOT NULL,
    fe_contract_tx_id bigint   NOT NULL,
    reason            smallint NOT NULL,
    state             smallint NOT NULL,
    CONSTRAINT fact_evidence_contract_states_pkey PRIMARY KEY (state_tx_id),
    CONSTRAINT fact_evidence_contract_states_state_tx_fkey FOREIGN KEY (state_tx_id)
        REFERENCES transactions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fact_evidence_contract_states_fe_contract_tx_id_fkey FOREIGN KEY (fe_contract_tx_id)
        REFERENCES fact_evidence_contracts (contract_tx_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fact_evidence_contract_states_state_fkey FOREIGN KEY (state)
        REFERENCES dic_fact_evidence_contract_states (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fact_evidence_contract_states_reason_fkey FOREIGN KEY (reason)
        REFERENCES dic_fact_evidence_contract_state_reasons (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);