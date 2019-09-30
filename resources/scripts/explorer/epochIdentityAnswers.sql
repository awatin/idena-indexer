select f.cid,
       ''                             address,
       a.answer,
       a.wrong_words,
       coalesce(f.answer, ''),
       COALESCE(f.wrong_words, false) wrongWords,
       coalesce(f.status),
       a.point
from answers a
         join flips f on f.id = a.flip_id
         join epoch_identities ei on ei.id = a.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses ad on ad.id = s.address_id
where ei.epoch = $1
  and lower(ad.address) = lower($2)
  and a.is_short = $3