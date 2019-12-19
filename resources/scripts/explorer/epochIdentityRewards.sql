select a.address,
       ei.epoch,
       0         block_height,
       vr.balance,
       vr.stake,
       dert.name "type"
from validation_rewards vr
         join epoch_identities ei on ei.id = vr.epoch_identity_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
         join dic_epoch_reward_types dert on dert.id = vr.type
where ei.epoch = $1
  and lower(a.address) = lower($2)