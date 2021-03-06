SELECT b.height,
       b.Hash,
       b.timestamp,
       (SELECT count(*) FROM transactions WHERE block_height = b.height) TX_COUNT,
       coalesce(a.address, '')                                           proposer,
       coalesce(vs.vrf_score, 0)                                         proposer_vrf_score,
       b.is_empty,
       b.body_size,
       b.full_size,
       b.vrf_proposer_threshold,
       b.fee_rate,
       c.burnt,
       c.minted,
       c.total_balance,
       c.total_stake,
       bf.flags
FROM (SELECT *
      FROM blocks
      WHERE ($3::bigint IS NULL OR height <= $3::bigint)
        AND epoch = $1
      ORDER BY height DESC
      LIMIT $2) b
         LEFT JOIN block_proposers p ON p.block_height = b.height
         LEFT JOIN block_proposer_vrf_scores vs ON vs.block_height = b.height
         LEFT JOIN addresses a ON a.id = p.address_id
         JOIN coins c ON c.block_height = b.height
         LEFT JOIN (SELECT block_height, array_agg("flag") flags
                    FROM block_flags
                    GROUP BY block_height) bf ON bf.block_height = b.height
ORDER BY b.height DESC