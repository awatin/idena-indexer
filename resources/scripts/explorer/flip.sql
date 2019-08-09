select a.address              author,
       f.size,
       b.timestamp,
       coalesce(f.answer, '') answer,
       coalesce(f.status, '') status,
       t.hash                 tx_hash,
       b.hash                 block_hash,
       b.height               block_height,
       e.epoch                epoch
from flips f
         join transactions t on t.id = f.tx_id
         join blocks b on b.id = t.block_id
         join epochs e on e.id = b.epoch_id
         join addresses a on a.id = t.from
where LOWER(f.cid) = LOWER($1)