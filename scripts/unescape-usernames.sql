-- One-off fix for usernames stored HTML-escaped by older versions.
-- Run exactly once, in the same deploy as the release that stops escaping.
-- Running it twice would corrupt a legitimate username such as "a&amp;b".
--
-- 1. Check how many rows are affected (if none, skip the update):
--      SELECT username FROM auth WHERE username ~ '&(amp|lt|gt|#39|#34);';
-- 2. Review for collisions after unescaping:
--      SELECT unescaped, count(*) FROM (
--        SELECT replace(replace(replace(replace(replace(username,
--          '&lt;','<'),'&gt;','>'),'&#39;',''''),'&#34;','"'),'&amp;','&') AS unescaped
--        FROM auth) t GROUP BY unescaped HAVING count(*) > 1;

BEGIN;
UPDATE auth
SET username = replace(replace(replace(replace(replace(username,
    '&lt;', '<'), '&gt;', '>'), '&#39;', ''''), '&#34;', '"'), '&amp;', '&')
WHERE username ~ '&(amp|lt|gt|#39|#34);';
COMMIT;
