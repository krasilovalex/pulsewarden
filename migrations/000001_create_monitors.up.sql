CREATE TABLE monitors (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    method TEXT NOT NULL DEFAULT 'GET',
    interval_seconds INTEGER NOT NULL,
    timeout_milliseconds INTEGER NOT NULL,
    expected_status_from SMALLINT NOT NULL DEFAULT 200,
    expected_status_to SMALLINT NOT NULL DEFAULT 299,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    next_check_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT monitors_name_not_blank
        CHECK (BTRIM(name) <> ''),

    CONSTRAINT monitors_url_not_blank
        CHECK (BTRIM(url) <> ''),

    CONSTRAINT monitors_method_supported
        CHECK (method IN ('GET')),

    CONSTRAINT monitors_interval_positive
        CHECK (interval_seconds > 0),

    CONSTRAINT monitors_timeout_positive
        CHECK (timeout_milliseconds > 0),

    CONSTRAINT monitors_expected_status_from_valid
        CHECK (
            expected_status_from >= 100
            AND expected_status_from <= 599
        ),

    CONSTRAINT monitors_expected_status_to_valid
        CHECK (
            expected_status_to >= 100
            AND expected_status_to <= 599
        ),

    CONSTRAINT monitors_expected_status_range_valid
        CHECK (expected_status_from <= expected_status_to)
);

CREATE INDEX monitors_due_idx
    ON monitors (next_check_at, id)
    WHERE enabled = TRUE;