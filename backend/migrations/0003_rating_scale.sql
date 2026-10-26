-- +goose Up
ALTER TABLE ratings DROP CONSTRAINT IF EXISTS ratings_value_check;
ALTER TABLE ratings
    ALTER COLUMN value TYPE SMALLINT
    USING CASE value
        WHEN 'cool' THEN 5
        WHEN 'bad' THEN 1
        ELSE NULL
    END;
ALTER TABLE ratings ADD CONSTRAINT ratings_value_check CHECK (value BETWEEN 1 AND 5);

-- +goose Down
ALTER TABLE ratings DROP CONSTRAINT IF EXISTS ratings_value_check;
ALTER TABLE ratings
    ALTER COLUMN value TYPE TEXT
    USING CASE
        WHEN value >= 4 THEN 'cool'
        ELSE 'bad'
    END;
ALTER TABLE ratings ADD CONSTRAINT ratings_value_check CHECK (value IN ('cool', 'bad'));
