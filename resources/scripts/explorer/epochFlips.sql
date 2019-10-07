select f.Cid,
       f.Size,
       a.address                      author,
       b.epoch,
       COALESCE(f.Status, '')         status,
       COALESCE(f.Answer, '')         answer,
       COALESCE(f.wrong_words, false) wrongWords,
       COALESCE(ww.cnt, 0)            wrong_words_votes,
       COALESCE(short.answers, 0),
       COALESCE(long.answers, 0),
       b.timestamp,
       coalesce(fi."data", ''::bytea) icon,
       coalesce(fw.word_1, 0)         word_id_1,
       coalesce(wd1.name, '')         word_name_1,
       coalesce(wd1.description, '')  word_desc_1,
       coalesce(fw.word_2, 0)         word_id_2,
       coalesce(wd2.name, '')         word_name_2,
       coalesce(wd2.description, '')  word_desc_2
from flips f
         join transactions t on t.id = f.tx_id
         join addresses a on a.id = t.from
         join blocks b on b.height = t.block_height
         left join (select a.flip_id, count(*) answers from answers a where a.is_short = true group by a.flip_id) short
                   on short.flip_id = f.id
         left join (select a.flip_id, count(*) answers from answers a where a.is_short = false group by a.flip_id) long
                   on long.flip_id = f.id
         left join (select a.flip_id, count(*) cnt
                    from answers a
                    where not a.is_short
                      and a.wrong_words
                    group by a.flip_id) ww
                   on ww.flip_id = f.id
         left join flips_data fd on fd.flip_id = f.id
         left join flip_icons fi on fi.flip_data_id = fd.id
         left join flip_words fw on fw.flip_id = f.id
         left join words_dictionary wd1 on wd1.id = fw.word_1
         left join words_dictionary wd2 on wd2.id = fw.word_2
where b.epoch = $1
order by b.height desc
limit $3
offset
$2