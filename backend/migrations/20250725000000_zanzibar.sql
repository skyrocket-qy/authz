CREATE OR REPLACE FUNCTION check(
    in_sbj_ns TEXT,
    in_sbj_name TEXT,
    in_rel TEXT,
    in_obj_ns TEXT,
    in_obj_name TEXT
) RETURNS BOOLEAN AS $$
DECLARE
    sbj_ns_id BIGINT;
    sbj_id BIGINT;
    sbj_rel_id BIGINT;
    rel_id BIGINT;
    obj_ns_id BIGINT;
    obj_id BIGINT;
BEGIN
    -- Resolve all IDs
    SELECT id INTO sbj_ns_id FROM namespaces WHERE name = in_sbj_ns;
    IF sbj_ns_id IS NULL THEN RETURN FALSE; END IF;

    SELECT id INTO sbj_id FROM instances WHERE name = in_sbj_name;
    IF sbj_id IS NULL THEN RETURN FALSE; END IF;

    SELECT id INTO rel_id FROM relations WHERE name = in_rel;
    IF rel_id IS NULL THEN RETURN FALSE; END IF;

    SELECT id INTO obj_ns_id FROM namespaces WHERE name = in_obj_ns;
    IF obj_ns_id IS NULL THEN RETURN FALSE; END IF;

    SELECT id INTO obj_id FROM instances WHERE name = in_obj_name;
    IF obj_id IS NULL THEN RETURN FALSE; END IF;

    -- Run permission check using resolved IDs
    RETURN EXISTS (
        WITH RECURSIVE check_cte AS (
            SELECT obj_ns_id, obj_id, relation_id
            FROM tuples
            WHERE sbj_ns_id = sbj_ns_id
            AND sbj_id = sbj_id

            UNION

            SELECT t.obj_ns_id, t.obj_id, t.relation_id
            FROM tuples t
            INNER JOIN check_cte pc
            ON t.sbj_ns_id = pc.obj_ns_id
            AND t.sbj_id = pc.obj_id
            AND t.sbj_relation_id = pc.relation_id
        )
        SELECT 1
        FROM check_cte
        WHERE relation_id = rel_id
        AND obj_ns_id = obj_ns_id
        AND obj_id = obj_id
        LIMIT 1
    );
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION lookup(
    in_sbj_ns TEXT,
    in_sbj_name TEXT,
    in_rel TEXT
) RETURNS TABLE (
    obj_ns TEXT,
    obj_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE objs AS (
        SELECT
            t.obj_ns_id,
            t.obj_id,
            t.relation_id
        FROM tuples t
        JOIN namespaces ns ON ns.id = t.obj_ns_id
        JOIN instances inst ON inst.id = t.obj_id
        WHERE
            t.sbj_ns_id = (SELECT id FROM namespaces WHERE name = in_sbj_ns)
            AND t.sbj_id = (SELECT id FROM instances WHERE name = in_sbj_name)
            AND t.relation_id = (SELECT id FROM relations WHERE name = in_rel)

        UNION

        SELECT
            t.obj_ns_id,
            t.obj_id,
            t.relation_id
        FROM tuples t
        JOIN objs o ON t.sbj_ns_id = o.obj_ns_id AND t.sbj_id = o.obj_id AND t.sbj_relation_id = o.relation_id
    )
    SELECT ns.name, inst.name
    FROM objs
    JOIN namespaces ns ON ns.id = objs.obj_ns_id
    JOIN instances inst ON inst.id = objs.obj_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION expand(
    in_ns TEXT,
    in_name TEXT,
    in_rel TEXT
) RETURNS TABLE(ns TEXT, name TEXT) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE expand_cte(ns, name) AS (
        SELECT s_ns.name, s.name
        FROM tuples t
        JOIN namespaces o_ns ON o_ns.id = t.obj_ns_id
        JOIN instances o ON o.id = t.obj_id
        JOIN relations r ON r.id = t.relation_id
        JOIN instances s ON s.id = t.sbj_id
        JOIN namespaces s_ns ON s_ns.id = t.sbj_ns_id
        WHERE o_ns.name = in_ns AND o.name = in_name AND r.name = in_rel

        UNION

        SELECT s_ns.name, s.name
        FROM tuples t
        JOIN relations r ON r.id = t.relation_id
        JOIN namespaces s_ns ON s_ns.id = t.sbj_ns_id
        JOIN instances s ON s.id = t.sbj_id
        JOIN expand_cte ec ON t.obj_ns_id = (SELECT id FROM namespaces WHERE name = ec.ns)
                        AND t.obj_id = (SELECT id FROM instances WHERE name = ec.name)
    )
    SELECT * FROM expand_cte;
END;
$$ LANGUAGE plpgsql;